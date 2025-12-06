package utils

import (
	"unicode"
)

// IsEmoji 判断一个字符串是否是 Emoji 表情符号
// 支持：
// - 基础 Emoji (U+1F300-U+1F9FF)
// - 补充 Emoji (U+2600-U+27BF, U+1F000-U+1F6FF)
// - 表情符号修饰符
// - 组合 Emoji (包含 ZWJ - Zero Width Joiner)
func IsEmoji(s string) bool {
	if s == "" {
		return false
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return false
	}

	// 检查字符串中是否包含 Emoji 字符
	hasEmoji := false
	for _, r := range runes {
		if isEmojiRune(r) {
			hasEmoji = true
			break
		}
	}

	return hasEmoji
}

// isEmojiRune 判断一个 rune 是否是 Emoji 字符
func isEmojiRune(r rune) bool {
	// Emoji 基本范围
	if r >= 0x1F300 && r <= 0x1F9FF {
		return true
	}

	// 补充符号和象形文字
	if r >= 0x2600 && r <= 0x27BF {
		return true
	}

	// 杂项符号
	if r >= 0x1F000 && r <= 0x1F6FF {
		return true
	}

	// 传输和地图符号
	if r >= 0x1F680 && r <= 0x1F6FF {
		return true
	}

	// Emoji 修饰符
	if r >= 0x1F3FB && r <= 0x1F3FF {
		return true
	}

	// 补充象形文字
	if r >= 0x1F900 && r <= 0x1F9FF {
		return true
	}

	// 其他常见 Emoji 符号
	if r >= 0x2300 && r <= 0x23FF {
		return true
	}

	// 特殊符号
	if r >= 0x2B00 && r <= 0x2BFF {
		return true
	}

	// 箭头
	if r >= 0x2190 && r <= 0x21FF {
		return true
	}

	// 数学运算符
	if r >= 0x2200 && r <= 0x22FF {
		return true
	}

	// 各种技术符号
	if r >= 0x2300 && r <= 0x23FF {
		return true
	}

	// 图形字符
	if r >= 0x25A0 && r <= 0x25FF {
		return true
	}

	// 零宽度连接符 (用于组合 Emoji)
	if r == 0x200D {
		return true
	}

	// 变体选择器
	if r >= 0xFE00 && r <= 0xFE0F {
		return true
	}

	return false
}

// ContainsEmoji 判断字符串是否包含 Emoji（即使不是纯 Emoji）
func ContainsEmoji(s string) bool {
	for _, r := range s {
		if isEmojiRune(r) {
			return true
		}
	}
	return false
}

// IsPureEmoji 判断字符串是否只包含 Emoji（不包含其他字符）
func IsPureEmoji(s string) bool {
	if s == "" {
		return false
	}

	runes := []rune(s)
	for _, r := range runes {
		// 跳过变体选择器和零宽连接符
		if r == 0x200D || (r >= 0xFE00 && r <= 0xFE0F) {
			continue
		}

		// 跳过空格
		if unicode.IsSpace(r) {
			continue
		}

		// 如果不是 Emoji，返回 false
		if !isEmojiRune(r) {
			return false
		}
	}

	return true
}

// IsEmojiKey 判断一个字符串是否可能是 Emoji 键
// 这是为 Redis 挑战特别设计的，判断键是否主要由 Emoji 组成
func IsEmojiKey(s string) bool {
	if s == "" {
		return false
	}

	runes := []rune(s)
	emojiCount := 0
	totalCount := 0

	for _, r := range runes {
		// 跳过特殊字符
		if r == 0x200D || (r >= 0xFE00 && r <= 0xFE0F) {
			continue
		}

		totalCount++
		if isEmojiRune(r) {
			emojiCount++
		}
	}

	// 如果至少有一个 Emoji，且 Emoji 占比超过 50%，就认为是 Emoji 键
	return emojiCount > 0 && float64(emojiCount)/float64(totalCount) >= 0.5
}
