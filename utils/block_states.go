package utils

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// "color":"orange"
// or
// "color"="orange" [current]
const BlockStatesDefaultSeparator string = "="

// MarshalBlockStates 将 blockStates 格式化为其字符串表示。
// 该函数不会返回错误，即便 blockStates 包含了不正确的数据类型。
// 其在被设计时被设计为“更加方便的格式化工具”。
//
// 可以保证 MarshalBlockStates 的输出是稳定的，因为其内部使用了
// 稳定排序 (slices.SortStableFunc)
func MarshalBlockStates(blockStates map[string]any) string {
	result := []string{}
	separator := BlockStatesDefaultSeparator

	keys := []string{}
	for key := range blockStates {
		keys = append(keys, key)
	}
	slices.SortStableFunc(keys, func(a string, b string) int {
		return strings.Compare(a, b)
	})

	for _, keyName := range keys {
		switch val := blockStates[keyName].(type) {
		// e.g. "color"="orange"
		case string:
			result = append(result, fmt.Sprintf(
				"%#v%s%#v", keyName, separator, val,
			))
		// e.g. "open_bit"=true
		case byte:
			if val == 0 {
				result = append(result, fmt.Sprintf("%#v%sfalse", keyName, separator))
			} else {
				result = append(result, fmt.Sprintf("%#v%strue", keyName, separator))
			}
		// e.g. "facing_direction"=0
		case int32:
			result = append(result, fmt.Sprintf("%#v%s%d", keyName, separator, val))
		}
	}

	return fmt.Sprintf("[%s]", strings.Join(result, ","))
}

// ParseBlockStatesString 将 blockStatesString 解析为 Go 地图类型。
// 无论给出的 blockStatesString 是否正确，可以保证返回的地图不是 nil。
//
// ParseBlockStatesString 被尽可能的设计简单，这意味着 blockStatesString 不
// 应该包含复杂的转义符，并且应该是正确的方块状态字符串，否则解析结果是不正确的。
//
// 例如，虽然 ["a,\"b"=true] 在语法上是正确的，但由于使用该实现使用逗号作为分割符，
// 因此最终将会得到 map[string]any{`\b`: byte(1)} 而非 map[string]any{`a,"b`: byte(1)}。
// 然而，我们预期正确的方块状态永远不可能在键或值中使用逗号(和复杂的转义)，所以这种问题不应发生。
//
// 该函数不会返回错误，并尽可能解析可能的每个字段，即便 blockStatesString 不完整。
// ParseBlockStatesString 在被设计时被设计为“更加方便的解析工具”
func ParseBlockStatesString(blockStatesString string) (result map[string]any) {
	if len(blockStatesString) < 2 {
		return make(map[string]any)
	}
	if blockStatesString[0] != '[' || blockStatesString[len(blockStatesString)-1] != ']' {
		return make(map[string]any)
	}

	separator := BlockStatesDefaultSeparator
	blockStatesString = blockStatesString[1 : len(blockStatesString)-1]
	result = make(map[string]any)

	for state := range strings.SplitSeq(blockStatesString, ",") {
		state := strings.TrimSpace(state)
		keyAndValue := strings.Split(state, separator)
		if len(keyAndValue) != 2 {
			continue
		}

		key := strings.ReplaceAll(strings.TrimSpace(keyAndValue[0]), `"`, "")
		value := strings.TrimSpace(keyAndValue[1])
		if len(value) < 1 {
			continue
		}

		switch value[0] {
		case '"':
			result[key] = strings.ReplaceAll(value, `"`, "")
		case 't', 'T':
			result[key] = byte(1)
		case 'f', 'F':
			result[key] = byte(0)
		default:
			val, err := strconv.ParseInt(value, 10, 32)
			if err != nil {
				continue
			}
			result[key] = int32(val)
		}
	}

	return
}
