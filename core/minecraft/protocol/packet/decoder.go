package packet

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"io"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
)

// Decoder handles the decoding of Minecraft packets sent through an io.Reader. These packets in turn contain
// multiple compressed packets.
type Decoder struct {
	// r holds the io.Reader that packets are read from if the reader does not implement packetReader. When
	// this is the case, the buf field has a non-zero length.
	r   io.Reader
	buf []byte

	// pr holds a packetReader (and io.Reader) that packets are read from if the io.Reader passed to
	// NewDecoder implements the packetReader interface.
	pr packetReader

	decompress         bool
	compression        Compression
	readCompressID     bool
	maxDecompressedLen int
	encrypt            *encrypt

	checkPacketLimit bool
}

// packetReader is used to read packets immediately instead of copying them in a buffer first. This is a
// specific case made to reduce RAM usage.
type packetReader interface {
	ReadPacket() ([]byte, error)
}

// NewDecoder returns a new decoder decoding data from the io.Reader passed. One read call from the reader is
// assumed to consume an entire packet.
func NewDecoder(reader io.Reader) *Decoder {
	if pr, ok := reader.(packetReader); ok {
		return &Decoder{checkPacketLimit: true, pr: pr}
	}
	return &Decoder{
		r:                reader,
		buf:              make([]byte, 1024*1024*3),
		checkPacketLimit: true,
	}
}

// EnableEncryption enables encryption for the Decoder using the secret key bytes passed. Each packet received
// will be decrypted.
func (decoder *Decoder) EnableEncryption(keyBytes [32]byte) {
	block, _ := aes.NewCipher(keyBytes[:])
	first12 := append([]byte(nil), keyBytes[:12]...)
	stream := cipher.NewCTR(block, append(first12, 0, 0, 0, 2))
	decoder.encrypt = newEncrypt(keyBytes[:], stream)
}

// EnableCompression enables compression for the Decoder.
// Note that NetEase compression is no need to read compress ID.
func (decoder *Decoder) EnableCompression(maxDecompressedLen int, readCompressID bool) {
	decoder.decompress = true
	decoder.readCompressID = readCompressID
	decoder.maxDecompressedLen = maxDecompressedLen
}

// DisableBatchPacketLimit disables the check that limits the number of packets allowed in a single packet
// batch. This should typically be called for Decoders decoding from a server connection.
func (decoder *Decoder) DisableBatchPacketLimit() {
	decoder.checkPacketLimit = false
}

const (
	// header is the header of compressed 'batches' from Minecraft.
	header = 0xfe
	// maximumInBatch is the maximum amount of packets that may be found in a batch. If a compressed batch has
	// more than this amount, decoding will fail.
	maximumInBatch = 812
)

// setCompression sets underlying compression by compressID.
// If compression is not set, the init a new one and save it.
// If compressID is unknown or compressID changed, then return non-nil error.
func (decoder *Decoder) setCompression(compressID uint16) error {
	// Get compress func
	compressFunc, found := CompressFuncByID(compressID)
	if !found {
		return fmt.Errorf("setCompression: unknown compression algorithm %d", compressID)
	}
	// If compression is nil, then set and return
	if decoder.compression == nil {
		decoder.compression = compressFunc()
		return nil
	}
	// Check compress ID
	if compressID != decoder.compression.EncodeCompression() {
		return fmt.Errorf(
			"setCompression: attempt to use another compression algorithm (origin = %d, current = %d)",
			decoder.compression.EncodeCompression(), compressID,
		)
	}
	// Return
	return nil
}

// Decode decodes one 'packet' from the io.Reader passed in NewDecoder(), producing a slice of packets that it
// held and an error if not successful.
func (decoder *Decoder) Decode() (packets [][]byte, err error) {
	var data []byte
	if decoder.pr == nil {
		var n int
		n, err = decoder.r.Read(decoder.buf)
		data = decoder.buf[:n]
	} else {
		data, err = decoder.pr.ReadPacket()
	}
	if err != nil {
		return nil, fmt.Errorf("read batch: %w", err)
	}
	if len(data) == 0 {
		return nil, nil
	}
	if data[0] != header {
		return nil, fmt.Errorf("decode batch: invalid header %x, expected %x", data[0], header)
	}
	data = data[1:]
	if decoder.encrypt != nil {
		decoder.encrypt.decrypt(data)
		if err := decoder.encrypt.verify(data); err != nil {
			// The packet did not have a correct checksum.
			return nil, fmt.Errorf("verify batch: %w", err)
		}
		data = data[:len(data)-8]
	}

	if decoder.decompress {
		// If need read compress ID as the prefix
		if decoder.readCompressID {
			if data[0] == 0xff {
				// No need to decompress
				data = data[1:]
			} else {
				// Set compression
				err = decoder.setCompression(uint16(data[0]))
				if err != nil {
					return nil, fmt.Errorf("Decode: %v", err)
				}
				// Do decompress
				data, err = decoder.compression.Decompress(data[1:], decoder.maxDecompressedLen)
				if err != nil {
					return nil, fmt.Errorf("decompress batch: %w", err)
				}
			}
		} else {
			_ = decoder.setCompression(CompressionAlgorithmNetEase)
			data, err = decoder.compression.Decompress(data, decoder.maxDecompressedLen)
			if err != nil {
				return nil, fmt.Errorf("decompress batch: %w", err)
			}
		}
	}

	b := bytes.NewBuffer(data)
	for b.Len() != 0 {
		var length uint32
		if err := protocol.Varuint32(b, &length); err != nil {
			return nil, fmt.Errorf("decode batch: read packet length: %w", err)
		}
		packets = append(packets, b.Next(int(length)))
	}
	if len(packets) > maximumInBatch && decoder.checkPacketLimit {
		return nil, fmt.Errorf("decode batch: number of packets %v exceeds max=%v", len(packets), maximumInBatch)
	}
	return packets, nil
}
