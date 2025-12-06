package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestSHA256(testing *testing.T) {
	input := []byte("hello world")
	expected := sha256.Sum256(input)
	result := EncodeToSHA256(input)
	testing.Logf("SHA256 result: %s", hex.EncodeToString(result))
	testing.Logf("Expected: %s", hex.EncodeToString(expected[:]))
}
