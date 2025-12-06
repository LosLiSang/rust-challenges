package utils

import "bytes"

var ipad []byte = bytes.Repeat([]byte{0x5c}, 64)
var opad []byte = bytes.Repeat([]byte{0x36}, 64)

func EncodeToHMACSHA256(key []byte, message string) []byte {
	paddedKey := padKey(key)
	innerKey := make([]byte, 64)
	for i := 0; i < len(paddedKey); i++ {
		innerKey[i] = paddedKey[i] ^ opad[i]
	}
	outerKey := make([]byte, 64)
	for i := 0; i < len(paddedKey); i++ {
		outerKey[i] = paddedKey[i] ^ ipad[i]
	}
	innerHash := EncodeToSHA256(append(innerKey, []byte(message)...))
	outerHash := EncodeToSHA256(append(outerKey, []byte(innerHash)...))
	return outerHash
}

func padKey(key []byte) []byte {
	if len(key) > 64 {
		key = EncodeToSHA256(key)
	}
	if len(key) < 64 {
		key = append(key, bytes.Repeat([]byte{0x00}, 64-len(key))...)
	}
	return key
}
