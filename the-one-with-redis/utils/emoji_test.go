package utils

import (
	"testing"
)

func TestIsEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "单个笑脸",
			input:    "😀",
			expected: true,
		},
		{
			name:     "火箭",
			input:    "🚀",
			expected: true,
		},
		{
			name:     "心形",
			input:    "❤️",
			expected: true,
		},
		{
			name:     "竖起大拇指",
			input:    "👍",
			expected: true,
		},
		{
			name:     "多个 Emoji",
			input:    "😀🚀",
			expected: true,
		},
		{
			name:     "普通文本",
			input:    "hello",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
		{
			name:     "中文",
			input:    "你好",
			expected: false,
		},
		{
			name:     "数字",
			input:    "123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmoji(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestContainsEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "纯 Emoji",
			input:    "😀",
			expected: true,
		},
		{
			name:     "文本加 Emoji",
			input:    "Hello 😀",
			expected: true,
		},
		{
			name:     "Emoji 在中间",
			input:    "Hello 😀 World",
			expected: true,
		},
		{
			name:     "纯文本",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "多个 Emoji",
			input:    "😀🚀❤️",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ContainsEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("ContainsEmoji(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPureEmoji(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "单个 Emoji",
			input:    "😀",
			expected: true,
		},
		{
			name:     "多个 Emoji",
			input:    "😀🚀❤️",
			expected: true,
		},
		{
			name:     "Emoji 加文本",
			input:    "😀hello",
			expected: false,
		},
		{
			name:     "Emoji 加数字",
			input:    "😀123",
			expected: false,
		},
		{
			name:     "纯文本",
			input:    "hello",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPureEmoji(tt.input)
			if result != tt.expected {
				t.Errorf("IsPureEmoji(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsEmojiKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "单个 Emoji",
			input:    "😀",
			expected: true,
		},
		{
			name:     "多个 Emoji",
			input:    "😀🚀",
			expected: true,
		},
		{
			name:     "Emoji 占多数",
			input:    "😀a",
			expected: true,
		},
		{
			name:     "文本占多数",
			input:    "hello😀",
			expected: false,
		},
		{
			name:     "纯文本",
			input:    "hello",
			expected: false,
		},
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsEmojiKey(tt.input)
			if result != tt.expected {
				t.Errorf("IsEmojiKey(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// 测试一些常见的 Emoji
func TestCommonEmojis(t *testing.T) {
	commonEmojis := []string{
		"😀", "😃", "😄", "😁", "😆", // 笑脸
		"❤️", "💙", "💚", "💛", "🧡", // 心形
		"🚀", "✈️", "🚗", "🚕", "🚙", // 交通工具
		"🍎", "🍊", "🍋", "🍌", "🍉", // 水果
		"⚽", "🏀", "🏈", "⚾", "🎾", // 运动
		"🎵", "🎶", "🎤", "🎧", "🎹", // 音乐
		"👍", "👎", "👏", "🙌", "🤝", // 手势
	}

	for _, emoji := range commonEmojis {
		t.Run("Emoji_"+emoji, func(t *testing.T) {
			if !IsEmoji(emoji) {
				t.Errorf("IsEmoji(%q) should be true", emoji)
			}
			if !IsPureEmoji(emoji) {
				t.Errorf("IsPureEmoji(%q) should be true", emoji)
			}
			if !ContainsEmoji(emoji) {
				t.Errorf("ContainsEmoji(%q) should be true", emoji)
			}
		})
	}
}
