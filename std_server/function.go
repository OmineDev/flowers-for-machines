package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/OmineDev/flowers-for-machines/core/minecraft/nbt"
	"github.com/OmineDev/flowers-for-machines/core/minecraft/protocol"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/block_helper"
	"github.com/OmineDev/flowers-for-machines/nbt_assigner/nbt_console"
	"github.com/OmineDev/flowers-for-machines/utils"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

func CheckAlive(c *gin.Context) {
	err := mcClient.Conn().Flush()
	if err != nil {
		c.JSON(http.StatusOK, CheckAliveResponse{
			Alive:     false,
			ErrorInfo: fmt.Sprintf("Bot is dead; err = %v", err),
		})
		return
	}
	c.JSON(http.StatusOK, CheckAliveResponse{Alive: true})
}

func ProcessExist(c *gin.Context) {
	_, _ = gameInterface.Commands().SendWSCommandWithResp("deop @s")
	_ = mcClient.Conn().Close()
	go func() {
		time.Sleep(time.Second)
		os.Exit(0)
	}()
}

func ChangeConsolePosition(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	var request ChangeConsolePosRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, ChangeConsolePosResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	err = console.ChangeConsolePosition(
		request.DimensionID,
		protocol.BlockPos{
			request.CenterX,
			request.CenterY,
			request.CenterZ,
		},
	)
	if err != nil {
		c.JSON(http.StatusOK, ChangeConsolePosResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Change console position failed; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, ChangeConsolePosResponse{Success: true})
}

func PlaceNBTBlock(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()

	var request PlaceNBTBlockRequest
	var blockNBT map[string]any

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, PlaceNBTBlockResponse{
			Success:   false,
			ErrorType: ResponseErrorTypeParseError,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	blockNBTBytes, err := base64.StdEncoding.DecodeString(request.BlockNBTBase64String)
	if err != nil {
		c.JSON(http.StatusOK, PlaceNBTBlockResponse{
			Success:   false,
			ErrorType: ResponseErrorTypeParseError,
			ErrorInfo: fmt.Sprintf("Failed to parse block NBT base64 string; err = %v", err),
		})
		return
	}
	err = nbt.UnmarshalEncoding(blockNBTBytes, &blockNBT, nbt.LittleEndian)
	if err != nil {
		c.JSON(http.StatusOK, PlaceNBTBlockResponse{
			Success:   false,
			ErrorType: ResponseErrorTypeParseError,
			ErrorInfo: fmt.Sprintf("Block NBT bytes is broken; err = %v", err),
		})
		return
	}

	canFast, uniqueID, offset, err := wrapper.PlaceNBTBlock(
		request.BlockName,
		utils.ParseBlockStatesString(request.BlockStatesString),
		blockNBT,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceNBTBlockResponse{
			Success:   false,
			ErrorType: ResponseErrorTypeRuntimeError,
			ErrorInfo: fmt.Sprintf("Runtime error: Failed to place NBT block; err = %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, PlaceNBTBlockResponse{
		Success:           true,
		CanFast:           canFast,
		StructureUniqueID: uniqueID.String(),
		StructureName:     utils.MakeUUIDSafeString(uniqueID),
		OffsetX:           offset.X(),
		OffsetY:           offset.Y(),
		OffsetZ:           offset.Z(),
	})
}

func PlaceLargeChest(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	var request PlaceLargeChestRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	// Step 1: Prepare
	chestBlockStates := utils.ParseBlockStatesString(request.BlockStatesString)
	center := console.Center()
	pairleadPos := protocol.BlockPos{
		center[0], center[1] + 1, center[2],
	}
	pairedPos := protocol.BlockPos{
		center[0] + 1, center[1] + 1, center[2],
	}

	// Step 2.1: Place pairlead chest
	if !request.PairleadChestStructureExist {
		err = gameInterface.SetBlock().SetBlock(pairleadPos, request.BlockName, request.BlockStatesString)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Place pairlead chest failed; err = %v", err),
			})
			return
		}
	} else {
		uniqueID, err := uuid.Parse(request.PairleadChestUniqueID)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Parse structure unique ID of pairlead chest failed; err = %v", err),
			})
			return
		}
		err = gameInterface.StructureBackup().RevertStructure(uniqueID, pairleadPos)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Revert structure for pairlead chest failed; err = %v", err),
			})
			return
		}
	}
	nearBlock := console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.ComplexBlock{
		KnownStates: true,
		Name:        request.BlockName,
		States:      chestBlockStates,
	}

	// Step 2.2: Place paired chest
	if !request.PairedChestStructureExist {
		err = gameInterface.SetBlock().SetBlock(pairedPos, request.BlockName, request.BlockStatesString)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Place paired chest failed; err = %v", err),
			})
			return
		}
	} else {
		uniqueID, err := uuid.Parse(request.PairedChestUniqueID)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Parse structure unique ID of paired chest failed; err = %v", err),
			})
			return
		}
		err = gameInterface.StructureBackup().RevertStructure(uniqueID, pairedPos)
		if err != nil {
			c.JSON(http.StatusOK, PlaceLargeChestResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Revert structure for paired chest failed; err = %v", err),
			})
			return
		}
	}

	// Step 3: Backup loaded chests
	tempStructure, err := gameInterface.StructureBackup().BackupOffset(pairleadPos, [3]int32{1, 0, 0})
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Backup temp structure failed; err = %v", err),
		})
		return
	}
	defer gameInterface.StructureBackup().DeleteStructure(tempStructure)

	// Step 4.1: Clean loaded chests
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"fill %d %d %d %d %d %d air",
			pairleadPos[0], pairleadPos[1], pairleadPos[2],
			pairedPos[0], pairedPos[1], pairedPos[2],
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Clean loaded chest failed; err = %v", err),
		})
		return
	}

	// Step 4.2: Wait clean down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes general failed (stage 1); err = %v", err),
		})
		return
	}
	nearBlock = console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.Air{}

	// Step 5.1: Revert loaded chests on console center
	err = gameInterface.StructureBackup().RevertStructure(tempStructure, console.Center())
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Revert temp structure failed; err = %v", err),
		})
		return
	}

	// Step 5.2: Sync changes to console
	nearBlock = console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{1, 0, 0})
	console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
		KnownStates: true,
		Name:        request.BlockName,
		States:      chestBlockStates,
	})
	*nearBlock = block_helper.ComplexBlock{
		KnownStates: true,
		Name:        request.BlockName,
		States:      chestBlockStates,
	}

	// Step 6.1: Clone revert structures to ~ ~1 ~
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"clone %d %d %d %d %d %d %d %d %d",
			center[0], center[1], center[2],
			center[0]+1, center[1], center[2],
			center[0], center[1]+1, center[2],
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Clone commands failed; err = %v", err),
		})
		return
	}

	// Step 6.2: Wait clone down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes general failed (stage 2); err = %v", err),
		})
		return
	}
	nearBlock = console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.ComplexBlock{
		KnownStates: true,
		Name:        request.BlockName,
		States:      chestBlockStates,
	}

	// Step 7.1: Get final structure (that included a large chest)
	finalStructure, err := gameInterface.StructureBackup().BackupOffset(pairleadPos, [3]int32{1, 0, 0})
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Get final structure failed; err = %v", err),
		})
		return
	}

	// Step 8.1: Clean loaded large chest
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"fill %d %d %d %d %d %d air",
			pairleadPos[0], pairleadPos[1], pairleadPos[2],
			pairedPos[0], pairedPos[1], pairedPos[2],
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Clean loaded chest failed; err = %v", err),
		})
		return
	}

	// Step 8.2: Wait clean down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes general failed (stage 3); err = %v", err),
		})
		return
	}
	nearBlock = console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.Air{}

	c.JSON(http.StatusOK, PlaceLargeChestResponse{
		Success:           true,
		StructureUniqueID: finalStructure.String(),
		StructureName:     utils.MakeUUIDSafeString(finalStructure),
	})
}

func PlaceWaterloggedBlock(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	var request PlaceWaterloggedBlockRequest

	err := c.BindJSON(&request)
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Failed to parse request; err = %v", err),
		})
		return
	}

	blockStates := utils.ParseBlockStatesString(request.NBTBlockStatesString)
	offsetBlockStates := utils.ParseBlockStatesString(request.OffsetNBTBlockStatesString)
	center := console.Center()
	offset := [3]int32{request.NBTStructureOffsetX, 0, request.NBTStructureOffsetZ}

	// Step 1: Place NBT block first
	if request.NBTStructureExist {
		uniqueID, err := uuid.Parse(request.NBTStructureUniqueID)
		if err != nil {
			c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Parse NBT structure unique ID failed; err = %v", err),
			})
			return
		}

		err = gameInterface.StructureBackup().RevertStructure(uniqueID, console.Center())
		if err != nil {
			c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Load NBT structure failed; err = %v", err),
			})
			return
		}

		console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
			KnownStates: true,
			Name:        request.NBTBlockName,
			States:      blockStates,
		})
		if offset != [3]int32{0, 0, 0} {
			nearBlock := console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, offset)
			*nearBlock = block_helper.ComplexBlock{
				KnownStates: true,
				Name:        request.NBTBlockName,
				States:      offsetBlockStates,
			}
		}
	} else {
		err = gameInterface.SetBlock().SetBlock(console.Center(), request.NBTBlockName, request.NBTBlockStatesString)
		if err != nil {
			c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
				Success:   false,
				ErrorInfo: fmt.Sprintf("Place NBT block failed; err = %v", err),
			})
			return
		}
		console.UseHelperBlock(nbt_console.RequesterUser, nbt_console.ConsoleIndexCenterBlock, block_helper.ComplexBlock{
			KnownStates: true,
			Name:        request.NBTBlockName,
			States:      blockStates,
		})
	}

	// Step 2.1: Place water
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"fill %d %d %d %d %d %d %s %s",
			center[0]+request.WaterStartOffsetX, center[1]+1, center[2]+request.WaterStartOffsetZ,
			center[0]+request.WaterEndOffsetX, center[1]+1, center[2]+request.WaterEndOffsetZ,
			request.WaterBlockName, request.WaterBlockStatesString,
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Place water failed; err = %v", err),
		})
		return
	}

	// Step 2.2: Wait water place down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes failed (stage 1); err = %v", err),
		})
		return
	}

	// Step 3.1: Clone loaded blocks to the water
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"clone %d %d %d %d %d %d %d %d %d",
			center[0], center[1], center[2],
			center[0]+offset[0], center[1], center[2]+offset[2],
			center[0], center[1]+1, center[2],
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceLargeChestResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Clone commands failed; err = %v", err),
		})
		return
	}

	// Step 3.2: Wait clone down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes failed (stage 2); err = %v", err),
		})
		return
	}
	nearBlock := console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.ComplexBlock{
		Name:   request.NBTBlockName,
		States: blockStates,
	}

	// Step 4: Get final structure (that included water logged NBT block)
	finalStructure, err := gameInterface.StructureBackup().BackupOffset(
		protocol.BlockPos{center[0], center[1] + 1, center[2]},
		offset,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Get final structure failed; err = %v", err),
		})
		return
	}

	// Step 5.1: Clean water logged block
	err = gameInterface.Commands().SendSettingsCommand(
		fmt.Sprintf(
			"fill %d %d %d %d %d %d air",
			center[0], center[1]+1, center[2],
			center[0]+offset[0], center[1]+1, center[2]+offset[2],
		),
		true,
	)
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Clean water logged block failed; err = %v", err),
		})
		return
	}

	// Step 5.2: Wait clean down
	err = gameInterface.Commands().AwaitChangesGeneral()
	if err != nil {
		c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
			Success:   false,
			ErrorInfo: fmt.Sprintf("Await changes general failed (stage 3); err = %v", err),
		})
		return
	}
	nearBlock = console.NearBlockByIndex(nbt_console.ConsoleIndexCenterBlock, protocol.BlockPos{0, 1, 0})
	*nearBlock = block_helper.Air{}

	c.JSON(http.StatusOK, PlaceWaterloggedBlockResponse{
		Success:           true,
		StructureUniqueID: finalStructure.String(),
		StructureName:     utils.MakeUUIDSafeString(finalStructure),
	})
}
