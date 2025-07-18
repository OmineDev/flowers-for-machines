package utils

import (
	"fmt"

	"github.com/OmineDev/flowers-for-machines/mapping"
)

// FormatBool 将 input 格式化为中文表示
func FormatBool(input bool) string {
	if input {
		return "是"
	}
	return "否"
}

// FormatByte 将 input 格式化为中文表示
func FormatByte(input uint8) string {
	if input == 0 {
		return "否"
	}
	return "是"
}

// FormatEnch 将格式化一个等级为 level 且 ID 为 id 的魔咒
func FormatEnch(id int16, level int16) string {
	levelString := ""

	index := int(level - 1)
	if index >= 0 && index < len(mapping.EnchLevelFormat) {
		levelString = mapping.EnchLevelFormat[index]
	} else {
		levelString = fmt.Sprintf("%d", level)
	}

	return mapping.EnchantFormat[id] + " " + levelString
}
