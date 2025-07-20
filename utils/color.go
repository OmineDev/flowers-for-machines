package utils

import "math"

// 计算两个 RGB 颜色 colorA 和 colorB 的欧式距离
func CalculateColorDistance(colorA [3]uint8, colorB [3]uint8) float64 {
	deltaR := float64(colorA[0]) - float64(colorB[0])
	deltaG := float64(colorA[1]) - float64(colorB[1])
	deltaB := float64(colorA[2]) - float64(colorB[2])
	return deltaR*deltaR + deltaG*deltaG + deltaB*deltaB
}

// 从 mapping 中选出距离 color 最近的 RGB 颜色
func SearchForBestColor(color [3]uint8, mapping [][3]uint8) (result [3]uint8) {
	distance := math.Inf(1)
	for _, value := range mapping {
		if deltaC := CalculateColorDistance(color, value); deltaC < distance {
			distance = deltaC
			result = value
		}
	}
	return
}

// 从 x 解码一个 RGBA 颜色
func DecodeVarRGBA(x int32) (RGB [3]uint8, RGBA [4]uint8) {
	R, G, B := uint8(x>>16), uint8(x>>8), uint8(x)
	A := uint8(x >> 24)
	return [3]uint8{R, G, B}, [4]uint8{R, G, B, A}
}

// EncodeVarRGBA 将 RGBA 颜色编码为 int32 整数
func EncodeVarRGBA(r uint8, g uint8, b uint8, a uint8) int32 {
	return int32(
		(uint32(r) << 16) | (uint32(g) << 8) | (uint32(b)) | (uint32(a) << 24),
	)
}
