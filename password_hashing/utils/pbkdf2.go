package utils

import "encoding/binary"

func EncodePBKDF2(password string, salt []byte, iterations int, keyLength int) []byte {
	// 计算需要的块数
	numBlocks := (keyLength + 31) / 32 // 每个块32字节
	// 初始化一个结果切片
	var result []byte = make([]byte, 0, numBlocks*32)
	for i := 1; i <= numBlocks; i++ {
		// 将盐和块索引拼接
		appendedSalt := appendInt32BE(salt, i)
		// 使用伪随机函数生成一个块
		block := F(password, appendedSalt, iterations)
		// 将生成的块添加到结果中
		result = append(result, block...)
	}
	// 截取结果到所需长度
	if len(result) > keyLength {
		result = result[:keyLength]
	}
	return result

}

func appendInt32BE(salt []byte, i int) []byte {
	// 创建一个和 salt 长度多4的新 slice
	res := make([]byte, len(salt)+4)
	copy(res, salt)
	// i 强转为 uint32
	binary.BigEndian.PutUint32(res[len(salt):], uint32(i))
	return res
}

func F(password string, appendedSalt []byte, iterations int) []byte {
	// 使用 HMAC-SHA256 作为伪随机函数
	passwordByte := []byte(password)
	hmac := EncodeToHMACSHA256(passwordByte, string(appendedSalt))
	// 将 HMAC 的结果转换为字节切片
	// 初始化一个结果切片
	result := make([]byte, len(hmac))
	copy(result, hmac)
	// 重复迭代次数次
	for i := 1; i < iterations; i++ {
		hmac = EncodeToHMACSHA256(passwordByte, string(hmac))
		for j := 0; j < len(hmac); j++ {
			// 将 HMAC 的每个字节添加到结果切片中
			result[j] ^= hmac[j]
		}
	}
	return result
}
