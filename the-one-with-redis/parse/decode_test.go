package parse

import (
	"strings"
	"testing"
)

// TestDecodeLength 测试长度解码
func TestDecodeLength(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected uint
		isError  bool
	}{
		{
			name:     "6 bit length",
			input:    []byte{0x0F}, // 00001111 = 15
			expected: 15,
			isError:  false,
		},
		{
			name:     "14 bit length",
			input:    []byte{0x40, 0xFF}, // 01000000 11111111 = 255
			expected: 255,
			isError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(string(tt.input))
			result, err := DecodeLength(reader)

			if tt.isError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.isError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

// TestDecodeString 测试字符串解码
func TestDecodeString(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected string
		isError  bool
	}{
		{
			name:     "simple string",
			input:    []byte{0x05, 'h', 'e', 'l', 'l', 'o'}, // length=5, "hello"
			expected: "hello",
			isError:  false,
		},
		{
			name:     "empty string",
			input:    []byte{0x00}, // length=0
			expected: "",
			isError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(string(tt.input))
			result, err := DecodeString(reader)

			if tt.isError && err == nil {
				t.Errorf("expected error but got none")
			}
			if !tt.isError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// TestDecodeIntset 测试 Intset 解码
func TestDecodeIntset(t *testing.T) {
	// Intset with encoding=4 (32-bit integers), length=3
	// 包含三个整数: 0x0000FFFC, 0x0000FFFD, 0x0000FFFE
	input := []byte{
		0x04, 0x00, 0x00, 0x00, // encoding = 4
		0x03, 0x00, 0x00, 0x00, // length = 3
		0xFC, 0xFF, 0x00, 0x00, // 65532
		0xFD, 0xFF, 0x00, 0x00, // 65533
		0xFE, 0xFF, 0x00, 0x00, // 65534
	}

	reader := strings.NewReader(string(input))
	result, err := DecodeIntset(reader)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 elements, got %d", len(result))
	}

	expected := []int64{65532, 65533, 65534}
	for i, v := range result {
		if v != expected[i] {
			t.Errorf("element %d: expected %d, got %d", i, expected[i], v)
		}
	}
}

// TestDecodeHashZiplist 测试 Hash Ziplist 转换
func TestDecodeHashZiplist(t *testing.T) {
	ziplistData := []interface{}{
		"key1", "value1",
		"key2", "value2",
		int64(123), "value3",
	}

	result, err := DecodeHashZiplist(ziplistData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 pairs, got %d", len(result))
	}

	if result["key1"] != "value1" {
		t.Errorf("expected value1, got %s", result["key1"])
	}

	if result["key2"] != "value2" {
		t.Errorf("expected value2, got %s", result["key2"])
	}

	if result["123"] != "value3" {
		t.Errorf("expected value3, got %s", result["123"])
	}
}

// TestDecodeSortedSetZiplist 测试 Sorted Set Ziplist 转换
func TestDecodeSortedSetZiplist(t *testing.T) {
	ziplistData := []interface{}{
		"member1", "1.5",
		"member2", int64(2),
		"member3", "3.14",
	}

	result, err := DecodeSortedSetZiplist(ziplistData)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("expected 3 members, got %d", len(result))
	}

	if result["member1"] != 1.5 {
		t.Errorf("expected 1.5, got %f", result["member1"])
	}

	if result["member2"] != 2.0 {
		t.Errorf("expected 2.0, got %f", result["member2"])
	}

	if result["member3"] != 3.14 {
		t.Errorf("expected 3.14, got %f", result["member3"])
	}
}
