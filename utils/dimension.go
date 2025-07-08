package utils

import "fmt"

// DimensionNameByID 返回维度 dimensionID 对应的字符串表示
func DimensionNameByID(dimensionID uint8) string {
	switch dimensionID {
	case 0:
		return "overworld"
	case 1:
		return "nether"
	case 2:
		return "the_end"
	default:
		return fmt.Sprintf("dm%d", dimensionID)
	}
}
