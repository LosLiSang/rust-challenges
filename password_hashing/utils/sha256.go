package utils

import (
	"encoding/binary"
)

var chunk = 64 // 512 bits

var h = [8]uint32{
	0x6a09e667,
	0xbb67ae85,
	0x3c6ef372,
	0xa54ff53a,
	0x510e527f,
	0x9b05688c,
	0x1f83d9ab,
	0x5be0cd19,
}

var k = [64]uint32{
	0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5,
	0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5,
	0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3,
	0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174,
	0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc,
	0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da,
	0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7,
	0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967,
	0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13,
	0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85,
	0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3,
	0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070,
	0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5,
	0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3,
	0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208,
	0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2,
}

func EncodeToSHA256(input []byte) []byte {
	// Implementation goes here
	inputBytes := input
	paddedInput := padding(inputBytes)
	blocks := parsing(paddedInput)
	hash := [8]uint32{}
	copy(hash[:], h[:])
	for _, block := range blocks {
		schedule := prepareSchedule(block)
		hash = processSchedule(hash, schedule)
	}
	// Convert final hash to hexadecimal string
	var result []byte
	for _, hPart := range hash {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, hPart)
		result = append(result, b...)
	}
	return result
}

func padding(m []byte) []byte {
	// length of the message m in bits
	l := len(m) * 8
	// 复制原始消息，避免修改输入
	result := make([]byte, len(m))
	copy(result, m)
	// add the '1' bit to the end of the message
	result = append(result, 0x80)
	// 计算需要的零填充：消息长度 + 1字节(0x80) + 零填充 + 8字节(长度) 必须是64字节的倍数
	zeroPadding := (64 - ((len(m) + 1 + 8) % 64)) % 64
	// 添加零填充
	result = append(result, make([]byte, zeroPadding)...)
	// length of the original message in binary (big-endian)
	mlen := make([]byte, 8)
	binary.BigEndian.PutUint64(mlen, uint64(l))
	// append the length of the original message in bits
	result = append(result, mlen...)
	return result
}

func parsing(m []byte) [][]byte {
	var mb [][]byte
	n := len(m) / chunk
	for i := 0; i < n; i++ {
		mb = append(mb, m[i*chunk:(i+1)*chunk])
	}
	return mb
}

func prepareSchedule(bl []byte) []uint32 {
	sch := make([]uint32, 64)
	for i := 0; i < 16; i++ {
		sch[i] = binary.BigEndian.Uint32(bl[i*4 : (i+1)*4])

	}
	for i := 16; i < 64; i++ {
		s0 := rotateRight(sch[i-15], 7) ^ rotateRight(sch[i-15], 18) ^ rightShift(sch[i-15], 3)
		s1 := rotateRight(sch[i-2], 17) ^ rotateRight(sch[i-2], 19) ^ rightShift(sch[i-2], 10)
		sch[i] = s1 + sch[i-7] + s0 + sch[i-16]

	}
	return sch
}

func processSchedule(H [8]uint32, ws []uint32) [8]uint32 {

	// initialize the eight working vars
	// with the (i-1) hash value
	a, b, c, d, e, f, g, h := H[0], H[1], H[2], H[3], H[4], H[5], H[6], H[7]
	// rotation processing
	for t, w := range ws {
		S0 := rotateRight(a, 2) ^ rotateRight(a, 13) ^ rotateRight(a, 22)
		S1 := rotateRight(e, 6) ^ rotateRight(e, 11) ^ rotateRight(e, 25)
		t1 := h + S1 + ch(e, f, g) + k[t] + w
		t2 := S0 + maj(a, b, c)
		// rotate working vars
		h = g
		g = f
		f = e
		e = d + t1
		d = c
		c = b
		b = a
		a = t1 + t2

	}
	// compute the ith intermidate hash value
	H[0] += a
	H[1] += b
	H[2] += c
	H[3] += d
	H[4] += e
	H[5] += f
	H[6] += g
	H[7] += h
	return H
}

func ch(x, y, z uint32) uint32 {
	return (x & y) ^ (^x & z)
}

func maj(x, y, z uint32) uint32 {
	return (x & y) ^ (x & z) ^ (y & z)
}

func rotateRight(x uint32, n uint) uint32 {
	return (x >> n) | (x << (32 - n))
}

func rightShift(x uint32, n uint) uint32 {
	return x >> n
}
