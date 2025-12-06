package utils

import "encoding/binary"

func EncodeToScrypt(password string, salt []byte, N int, r int, p int, buflen int) []byte {
	blockSize := 128 * r
	// 1) B = PBKDF2(P, S, 1, p * 128 * r)
	B := EncodePBKDF2(password, salt, 1, p*blockSize) // 这里 iterations=1

	// 2) 把 B 拆成 p 个块，每块 128*r 字节
	Bblocks := make([][]byte, p)
	for i := 0; i < p; i++ {
		start := i * blockSize
		end := start + blockSize
		Bblocks[i] = make([]byte, blockSize)
		copy(Bblocks[i], B[start:end])
	}

	// 3) 对每个块做 ROMix(r, N)
	for i := 0; i < p; i++ {
		Bblocks[i] = ROMix(Bblocks[i], r, N)
	}

	// 4) 把处理后的块拼回一个大 B'
	mixed := make([]byte, 0, p*blockSize)
	for i := 0; i < p; i++ {
		mixed = append(mixed, Bblocks[i]...)
	}

	// 5) 再做一次 PBKDF2(P, B', 1, dkLen)
	DK := EncodePBKDF2(password, mixed, 1, buflen)
	return DK
}

func ROMix(b []byte, r, N int) []byte {
	blockLen := 128 * r

	X := make([]byte, blockLen)
	copy(X, b)
	var v [][]byte = make([][]byte, N)

	for i := 0; i < N; i++ {
		v[i] = make([]byte, blockLen)
		copy(v[i], X)
		X = BlockMix(X, r)
	}
	// 3) 后半段：随机访问 V
	for i := 0; i < N; i++ {
		j := Integerify(X, r) % N
		X = xorBytes(X, v[j]) // 逐字节 XOR
		X = BlockMix(X, r)
	}
	return X
}

func xorBytes(X []byte, b []byte) []byte {
	result := make([]byte, len(X))
	for i := 0; i < len(X); i++ {
		result[i] = X[i] ^ b[i]
	}
	return result
}


func Integerify(X []byte, r int) int {
    // 修正：读取 X 的最后一个 64 字节块的前 8 字节
    // X 的结构是 2r 个 64字节块: B[0] ... B[2r-1]
    // 我们需要 B[2r-1]
    lenX := len(X)
    blockSize := 64
    
    // 最后一个块的偏移量
    offset := lenX - blockSize
    
    // 从该偏移量读取 uint64
    v := binary.LittleEndian.Uint64(X[offset : offset+8])
    
    // 注意：虽然结果被截断为 int，但在 Scrypt 中 N 也是 int
    // 且通常 N 是 2 的幂，所以取模操作 v % N 实际上只依赖低位，这在 32位/64位系统上通常是安全的
    return int(v & 0x7FFFFFFFFFFFFFFF)
}

func BlockMix(X []byte, r int) []byte {
	blockSize := 64 // 64 字节
	numBlocks := 2 * r
	if len(X) != numBlocks*blockSize {
		panic("wrong length")
	}

	// 将 X 分成 2*r 个 64 字节的块
	blocks := make([][]byte, numBlocks)
	for i := 0; i < numBlocks; i++ {
		blocks[i] = make([]byte, blockSize)
		copy(blocks[i], X[i*blockSize:(i+1)*blockSize])
	}

	// Y 用于存储结果
	Y := make([][]byte, numBlocks)

	// 1. X = B[2*r - 1]
	tmpX := make([]byte, blockSize)
	copy(tmpX, blocks[numBlocks-1])

	// 2. for i = 0 to 2r - 1 do
	for i := 0; i < numBlocks; i++ {
		// T = X xor B[i]
		for j := 0; j < blockSize; j++ {
			tmpX[j] ^= blocks[i][j]
		}
		// X = Salsa20/8(T)
		tmpX = salsa208(tmpX)
		// Y[i] = X
		Y[i] = make([]byte, blockSize)
		copy(Y[i], tmpX)
	}

	// 3. 重新排列 Y: 偶数索引在前，奇数索引在后
	result := make([]byte, 0, len(X))
	// 先添加所有偶数索引的块
	for i := 0; i < numBlocks; i += 2 {
		result = append(result, Y[i]...)
	}
	// 再添加所有奇数索引的块
	for i := 1; i < numBlocks; i += 2 {
		result = append(result, Y[i]...)
	}

	return result
}

// salsa208 实现 Salsa20/8 核心函数
func salsa208(input []byte) []byte {
	if len(input) != 64 {
		panic("salsa208: input must be 64 bytes")
	}

	// 将输入转换为 16 个 uint32
	x := make([]uint32, 16)
	for i := 0; i < 16; i++ {
		x[i] = binary.LittleEndian.Uint32(input[i*4 : (i+1)*4])
	}

	// 保存初始状态用于最后的加法
	initial := make([]uint32, 16)
	copy(initial, x)

	// 执行 8 轮（4 个双轮）
	for i := 0; i < 4; i++ {
		// 列轮
		quarterRound(&x[0], &x[4], &x[8], &x[12])
		quarterRound(&x[5], &x[9], &x[13], &x[1])
		quarterRound(&x[10], &x[14], &x[2], &x[6])
		quarterRound(&x[15], &x[3], &x[7], &x[11])
		// 行轮
		quarterRound(&x[0], &x[1], &x[2], &x[3])
		quarterRound(&x[5], &x[6], &x[7], &x[4])
		quarterRound(&x[10], &x[11], &x[8], &x[9])
		quarterRound(&x[15], &x[12], &x[13], &x[14])
	}

	// 加上初始状态
	for i := 0; i < 16; i++ {
		x[i] += initial[i]
	}

	// 转换回字节数组
	output := make([]byte, 64)
	for i := 0; i < 16; i++ {
		binary.LittleEndian.PutUint32(output[i*4:(i+1)*4], x[i])
	}

	return output
}

// quarterRound 是 Salsa20 的四分之一轮函数
func quarterRound(y0, y1, y2, y3 *uint32) {
	*y1 ^= rotateLeft(*y0+*y3, 7)
	*y2 ^= rotateLeft(*y1+*y0, 9)
	*y3 ^= rotateLeft(*y2+*y1, 13)
	*y0 ^= rotateLeft(*y3+*y2, 18)
}

// rotateLeft 执行循环左移
func rotateLeft(x uint32, n uint) uint32 {
	return (x << n) | (x >> (32 - n))
}
