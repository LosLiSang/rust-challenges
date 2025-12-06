package parse

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"strconv"
	"strings"

	lzf "github.com/zhuyie/golzf"
)

var (
	// ErrLZFEncoding 用于标识 LZF 压缩字符串的特殊长度编码
	ErrLZFEncoding = errors.New("lzf encoding")
	ErrIntEncoding = errors.New("int encoding")
)

func DecodeLength(r *strings.Reader) (uint, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	switch b >> 6 {
	case 0:
		// 6 bit length
		return uint(b & 0x3F), nil
	case 1:
		// 14 bit length (6 bits in b, 8 bits in next byte)
		b2, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		return uint((uint(b)&0x3F)<<8 | uint(b2)), nil
	case 2:
		// 32 bit length (next 4 bytes, little endian)
		buf := make([]byte, 4)
		if _, err := io.ReadFull(r, buf); err != nil {
			return 0, err
		}
		return uint(binary.LittleEndian.Uint32(buf)), nil
	case 3:
		// special encoding: lower 6 bits indicate type
		typ := b & 0x3F
		switch typ {
		case 0:
			// 8 bit integer
			v, err := r.ReadByte()
			if err != nil {
				return 0, err
			}
			return uint(v), ErrIntEncoding
		case 1:
			// 16 bit integer (little endian)
			buf := make([]byte, 2)
			if _, err := io.ReadFull(r, buf); err != nil {
				return 0, err
			}
			return uint(binary.LittleEndian.Uint16(buf)), ErrIntEncoding
		case 2:
			// 32 bit integer (little endian)
			buf := make([]byte, 4)
			if _, err := io.ReadFull(r, buf); err != nil {
				return 0, err
			}
			return uint(binary.LittleEndian.Uint32(buf)), ErrIntEncoding
		case 3:
			// LZF compressed string marker
			return 0, ErrLZFEncoding
		default:
			return 0, fmt.Errorf("unsupported special length encoding: %d", typ)
		}
	}
	return 0, fmt.Errorf("invalid length encoding")
}

func DecodeString(r *strings.Reader) (string, error) {
	n, err := DecodeLength(r)
	if err != nil {
		switch err {
		case ErrIntEncoding:
			// 整数编码，读取整数值
			return strconv.Itoa(int(n)), nil
		case ErrLZFEncoding:
			// LZF 压缩字符串，使用专门的解析器
			return DecodeLFString(r)
		}
		return "", err
	}
	if n == 0 {
		return "", nil
	}
	// Length Prefix
	buf := make([]byte, int(n))
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

func DecodeLFString(r *strings.Reader) (string, error) {
	clen, err := DecodeLength(r)
	if err != nil {
		return "", err
	}
	uncompressLen, err := DecodeLength(r)
	if err != nil {
		return "", err
	}

	inputBuf := make([]byte, int(clen))
	if _, err := io.ReadFull(r, inputBuf); err != nil {
		return "", err
	}

	outputBuf := make([]byte, uncompressLen)
	if _, err := lzf.Decompress(inputBuf, outputBuf); err != nil {
		return "", err
	}
	log.Println(string(outputBuf[:uncompressLen]))
	return string(outputBuf[:uncompressLen]), nil
}

func DecodeList(r *strings.Reader) ([]string, error) {
	size, err := DecodeLength(r)
	if err != nil {
		return nil, err
	}

	// 预分配切片以避免重复分配
	result := make([]string, 0, int(size))
	for i := uint(0); i < size; i++ {
		str, err := DecodeString(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode list item %d: %w", i, err)
		}
		result = append(result, str)
	}
	return result, nil
}

func DecodeSortedSet(r *strings.Reader) (map[string]struct{}, error) {
	size, err := DecodeLength(r)
	if err != nil {
		return nil, err
	}

	result := make(map[string]struct{}, int(size))
	for i := uint(0); i < size; i++ {
		// 读取值
		value, err := DecodeString(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode sorted set value %d: %w", i, err)
		}

		// 读取分数：先读取一个字节判断类型
		score_len_byte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		var score float64
		switch score_len_byte {
		case 255:
			// 负无穷
			score = math.Inf(-1)
		case 254:
			// 正无穷
			score = math.Inf(1)
		case 253:
			// NaN
			score = math.NaN()
		default:
			// 普通分数，这个字节表示分数字符串的长度
			score_buf := make([]byte, int(score_len_byte))
			if _, err := io.ReadFull(r, score_buf); err != nil {
				return nil, err
			}
			score, err = strconv.ParseFloat(string(score_buf), 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse score: %w", err)
			}
		}

		fmt.Printf("Sorted Set Value: %s, Score: %v\n", value, score)
		result[value] = struct{}{}
	}
	return result, nil
}

func DecodeHash(r *strings.Reader) (map[string]string, error) {
	size, err := DecodeLength(r)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string, int(size))
	for i := uint(0); i < size; i++ {
		key, err := DecodeString(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash key %d: %w", i, err)
		}
		value, err := DecodeString(r)
		if err != nil {
			return nil, fmt.Errorf("failed to decode hash value %d: %w", i, err)
		}
		result[key] = value
	}
	return result, nil
}

func DecodeZipmap(r *strings.Reader) (map[string]string, error) {
	// 读取 zmlen (zipmap 长度)，如果 >= 254，需要遍历整个 zipmap
	zmlen, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	_ = zmlen // 如果 >= 254，这个值无效，需要遍历

	result := make(map[string]string)

	for {
		// 读取 key 的长度
		keyLenByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		// 如果是 255，表示 zipmap 结束
		if keyLenByte == 255 {
			break
		}

		var keyLen uint
		if keyLenByte == 253 {
			// 长度存储在接下来的 4 字节中
			buf := make([]byte, 4)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			keyLen = uint(binary.LittleEndian.Uint32(buf))
		} else if keyLenByte == 254 || keyLenByte == 255 {
			return nil, fmt.Errorf("invalid key length: %d", keyLenByte)
		} else {
			keyLen = uint(keyLenByte)
		}

		// 读取 key
		key := make([]byte, keyLen)
		if _, err := io.ReadFull(r, key); err != nil {
			return nil, err
		}

		// 读取 value 的长度
		valLenByte, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		var valLen uint
		if valLenByte == 253 {
			buf := make([]byte, 4)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			valLen = uint(binary.LittleEndian.Uint32(buf))
		} else if valLenByte == 254 || valLenByte == 255 {
			return nil, fmt.Errorf("invalid value length: %d", valLenByte)
		} else {
			valLen = uint(valLenByte)
		}

		// 读取 free 字节数
		freeLen, err := r.ReadByte()
		if err != nil {
			return nil, err
		}

		// 读取 value
		value := make([]byte, valLen)
		if _, err := io.ReadFull(r, value); err != nil {
			return nil, err
		}

		// 跳过 free 字节
		if freeLen > 0 {
			freeBuf := make([]byte, freeLen)
			if _, err := io.ReadFull(r, freeBuf); err != nil {
				return nil, err
			}
		}

		result[string(key)] = string(value)
	}

	return result, nil
}

// DecodeZiplist 解码 Ziplist 编码的数据
// Ziplist 是一个紧凑的列表编码，包含 zlbytes, zltail, zllen, entries, zlend
func DecodeZiplist(r *strings.Reader) ([]interface{}, error) {
	// 读取 zlbytes (4 字节，小端序)
	zlbytesRaw := make([]byte, 4)
	if _, err := io.ReadFull(r, zlbytesRaw); err != nil {
		return nil, fmt.Errorf("failed to read zlbytes: %w", err)
	}
	zlbytes := binary.LittleEndian.Uint32(zlbytesRaw)

	// 读取 zltail (4 字节，小端序) - 指向最后一个元素的偏移量
	zltailRaw := make([]byte, 4)
	if _, err := io.ReadFull(r, zltailRaw); err != nil {
		return nil, fmt.Errorf("failed to read zltail: %w", err)
	}

	// 读取 zllen (2 字节，小端序) - 元素数量
	zllenRaw := make([]byte, 2)
	if _, err := io.ReadFull(r, zllenRaw); err != nil {
		return nil, fmt.Errorf("failed to read zllen: %w", err)
	}
	zllen := binary.LittleEndian.Uint16(zllenRaw)

	result := make([]interface{}, 0, int(zllen))

	// 读取 entries
	for i := 0; i < int(zllen); i++ {
		// 读取前一个元素的长度
		prevLen, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read prev length: %w", err)
		}
		if prevLen == 254 {
			// 如果是 254，后续 4 字节才是长度
			skipBuf := make([]byte, 4)
			if _, err := io.ReadFull(r, skipBuf); err != nil {
				return nil, err
			}
		}

		// 读取 special flag
		flag, err := r.ReadByte()
		if err != nil {
			return nil, fmt.Errorf("failed to read flag: %w", err)
		}

		// 根据 flag 解码数据
		var value interface{}
		switch {
		case (flag >> 6) == 0: // 00pppppp - 字符串，长度 <= 63
			length := int(flag & 0x3F)
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			value = string(buf)

		case (flag >> 6) == 1: // 01pppppp qqqqqqqq - 字符串，长度 <= 16383
			nextByte, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			length := int((uint(flag)&0x3F)<<8 | uint(nextByte))
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			value = string(buf)

		case (flag >> 6) == 2: // 10______ - 字符串，长度 >= 16384
			lengthBuf := make([]byte, 4)
			if _, err := io.ReadFull(r, lengthBuf); err != nil {
				return nil, err
			}
			// 注意：这里需要使用 Big Endian（根据规范）
			length := binary.BigEndian.Uint32(lengthBuf)
			buf := make([]byte, length)
			if _, err := io.ReadFull(r, buf); err != nil {
				return nil, err
			}
			value = string(buf)

		case (flag & 0xF0) == 0xC0: // 1100____ - 16 位整数
			intBuf := make([]byte, 2)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value = int64(int16(binary.LittleEndian.Uint16(intBuf)))

		case (flag & 0xF0) == 0xD0: // 1101____ - 32 位整数
			intBuf := make([]byte, 4)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value = int64(int32(binary.LittleEndian.Uint32(intBuf)))

		case (flag & 0xF0) == 0xE0: // 1110____ - 64 位整数
			intBuf := make([]byte, 8)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value = int64(binary.LittleEndian.Uint64(intBuf))

		case (flag & 0xF0) == 0xF0: // 1111____ - 24 位整数或其他
			lowerBits := flag & 0x0F
			switch lowerBits {
			case 0x0E: // 11111110 - 8 位整数
				intVal, err := r.ReadByte()
				if err != nil {
					return nil, err
				}
				value = int64(int8(intVal))
			case 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D:
				// 1111xxxx - 4 位立即整数 (0-12)，实际值是 xxxx - 1
				value = int64(lowerBits - 1)
			case 0x0F:
				// 这是特殊情况，需要读取 3 字节的 24 位整数
				intBuf := make([]byte, 3)
				if _, err := io.ReadFull(r, intBuf); err != nil {
					return nil, err
				}
				// 24 位有符号整数，需要手动转换
				val := int32(intBuf[0]) | int32(intBuf[1])<<8 | int32(intBuf[2])<<16
				// 处理符号位
				if val&0x800000 != 0 {
					val |= ^0xFFFFFF
				}
				value = int64(val)
			default:
				return nil, fmt.Errorf("invalid ziplist flag lower bits: 0x%02X", lowerBits)
			}

		default:
			return nil, fmt.Errorf("unsupported ziplist flag: 0x%02X", flag)
		}

		result = append(result, value)
	}

	// 读取 zlend (应该是 0xFF)
	zlend, err := r.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("failed to read zlend: %w", err)
	}
	if zlend != 0xFF {
		return nil, fmt.Errorf("invalid zlend: expected 0xFF, got 0x%02X", zlend)
	}

	// 验证读取的字节数
	_ = zlbytes // zlbytes 用于验证，这里简化处理

	return result, nil
}

// DecodeIntset 解码 Intset 编码的数据
// Intset 是一个整数集合，所有元素都是整数并按顺序存储
func DecodeIntset(r *strings.Reader) ([]int64, error) {
	// 读取 encoding (4 字节，小端序)
	encodingBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, encodingBuf); err != nil {
		return nil, fmt.Errorf("failed to read encoding: %w", err)
	}
	encoding := binary.LittleEndian.Uint32(encodingBuf)

	// encoding 可以是 2, 4, 或 8，表示每个整数的字节数
	if encoding != 2 && encoding != 4 && encoding != 8 {
		return nil, fmt.Errorf("invalid intset encoding: %d", encoding)
	}

	// 读取 length (4 字节，小端序)
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lengthBuf); err != nil {
		return nil, fmt.Errorf("failed to read length: %w", err)
	}
	length := binary.LittleEndian.Uint32(lengthBuf)

	result := make([]int64, 0, int(length))

	// 根据 encoding 读取整数
	for i := uint32(0); i < length; i++ {
		switch encoding {
		case 2: // 16 位整数
			intBuf := make([]byte, 2)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value := int64(int16(binary.LittleEndian.Uint16(intBuf)))
			result = append(result, value)

		case 4: // 32 位整数
			intBuf := make([]byte, 4)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value := int64(int32(binary.LittleEndian.Uint32(intBuf)))
			result = append(result, value)

		case 8: // 64 位整数
			intBuf := make([]byte, 8)
			if _, err := io.ReadFull(r, intBuf); err != nil {
				return nil, err
			}
			value := int64(binary.LittleEndian.Uint64(intBuf))
			result = append(result, value)
		}
	}

	return result, nil
}

// DecodeSortedSetZiplist 解码 Ziplist 编码的 Sorted Set
// 在 ziplist 中，每个元素后面跟着它的分数
func DecodeSortedSetZiplist(ziplistData []interface{}) (map[string]float64, error) {
	if len(ziplistData)%2 != 0 {
		return nil, fmt.Errorf("invalid sorted set ziplist: odd number of elements")
	}

	result := make(map[string]float64, len(ziplistData)/2)

	for i := 0; i < len(ziplistData); i += 2 {
		// 元素
		var member string
		switch v := ziplistData[i].(type) {
		case string:
			member = v
		case int64:
			member = strconv.FormatInt(v, 10)
		default:
			return nil, fmt.Errorf("invalid member type: %T", v)
		}

		// 分数
		var score float64
		switch v := ziplistData[i+1].(type) {
		case string:
			var err error
			score, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse score: %w", err)
			}
		case int64:
			score = float64(v)
		default:
			return nil, fmt.Errorf("invalid score type: %T", v)
		}

		result[member] = score
	}

	return result, nil
}

// DecodeHashZiplist 解码 Ziplist 编码的 Hash
// 在 ziplist 中，key 和 value 交替存储
func DecodeHashZiplist(ziplistData []interface{}) (map[string]string, error) {
	if len(ziplistData)%2 != 0 {
		return nil, fmt.Errorf("invalid hash ziplist: odd number of elements")
	}

	result := make(map[string]string, len(ziplistData)/2)

	for i := 0; i < len(ziplistData); i += 2 {
		// Key
		var key string
		switch v := ziplistData[i].(type) {
		case string:
			key = v
		case int64:
			key = strconv.FormatInt(v, 10)
		default:
			return nil, fmt.Errorf("invalid key type: %T", v)
		}

		// Value
		var value string
		switch v := ziplistData[i+1].(type) {
		case string:
			value = v
		case int64:
			value = strconv.FormatInt(v, 10)
		default:
			return nil, fmt.Errorf("invalid value type: %T", v)
		}

		result[key] = value
	}

	return result, nil
}

// DecodeQuicklist 解码 Quicklist 编码的数据
// Quicklist 是一个 ziplist 的链表
func DecodeQuicklist(r *strings.Reader) ([]interface{}, error) {
	// 读取节点数量
	nodeCount, err := DecodeLength(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read quicklist node count: %w", err)
	}

	var result []interface{}

	// 读取每个节点（每个节点是一个 ziplist）
	for i := uint(0); i < nodeCount; i++ {
		// 读取 ziplist（以 String 编码包装）
		ziplistStr, err := DecodeString(r)
		if err != nil {
			return nil, fmt.Errorf("failed to read ziplist %d: %w", i, err)
		}

		// 解析 ziplist
		ziplistReader := strings.NewReader(ziplistStr)
		ziplistData, err := DecodeZiplist(ziplistReader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ziplist %d: %w", i, err)
		}

		// 将 ziplist 中的元素添加到结果中
		result = append(result, ziplistData...)
	}

	return result, nil
}
