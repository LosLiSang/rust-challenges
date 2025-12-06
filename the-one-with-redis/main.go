package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"the-one-with-redis/model"
	"the-one-with-redis/parse"
	"the-one-with-redis/utils"
	"unicode/utf8"
)

var response = utils.Solution{
	DbCount:       0,
	EmojiKeyValue: "",
	ExpiryMillis:  0,
	Extra:         map[string]interface{}{},
}

func main() {
	accessToken := "aa8e14a982d0994e"
	res, e := utils.DownloadProblem(accessToken)
	if e != nil {
		fmt.Println("Failed to download problem:", e)
		return
	}
	fmt.Println("Check type of:", res.Requirements.CheckTypeOf)

	rdb_data, err := base64.StdEncoding.DecodeString(res.RDB)
	if err != nil {
		fmt.Println("decode error:", err)
		return
	}

	rdb_stream := strings.NewReader(string(rdb_data))

	// Read RDB file header
	magic_buf := make([]byte, 5)
	if _, err := io.ReadFull(rdb_stream, magic_buf); err != nil {
		fmt.Println("Failed to read magic:", err)
		return
	}
	fmt.Printf("Magic: %s\n", string(magic_buf))

	version_buf := make([]byte, 4)
	if _, err := io.ReadFull(rdb_stream, version_buf); err != nil {
		fmt.Println("Failed to read version:", err)
		return
	}
	fmt.Printf("Version: %s\n", string(version_buf))

	// Parse RDB body
	if err := parseRDBBody(rdb_stream, res.Requirements.CheckTypeOf); err != nil {
		fmt.Println("Failed to parse RDB:", err)
		return
	}
	fmt.Printf("%v\n", response)
	r, err := utils.UploadSolution(accessToken, &response)
	if err != nil {
		fmt.Println("Failed to upload solution:", err)
	}
	fmt.Println(r)
}

func parseRDBBody(rdb_stream *strings.Reader, checkTypeOf string) error {
	var currentDB uint
	var expiryTime int64 = -1 // -1 表示没有过期时间
	var expiryInMs bool

	for {
		b, err := rdb_stream.ReadByte()
		if err != nil {
			if err == io.EOF {
				fmt.Println("Unexpected EOF")
				return err
			}
			return err
		}

		switch b {
		case model.EOF:
			fmt.Println("=== Reached EOF ===")
			// 读取 CRC64 校验和（如果存在）
			checksum := make([]byte, 8)
			if _, err := io.ReadFull(rdb_stream, checksum); err != nil {
				if err != io.EOF {
					fmt.Println("Warning: Failed to read checksum:", err)
				}
			} else {
				fmt.Printf("CRC64 Checksum: %x\n", checksum)
			}
			return nil

		case model.SELECTDB:
			db_number, err := parse.DecodeLength(rdb_stream)
			if err != nil {
				return fmt.Errorf("failed to decode DB number: %w", err)
			}
			currentDB = db_number
			fmt.Printf("\n=== Switching to Database %d ===\n", db_number)

			// Service Details Requirement
			response.DbCount++

		case model.RESIZEDB:
			dbHashTableSize, err := parse.DecodeLength(rdb_stream)
			if err != nil {
				return fmt.Errorf("failed to decode DB hash table size: %w", err)
			}
			expiryHashTableSize, err := parse.DecodeLength(rdb_stream)
			if err != nil {
				return fmt.Errorf("failed to decode expiry hash table size: %w", err)
			}
			fmt.Printf("DB %d - HashTable Size: %d, Expiry HashTable Size: %d\n",
				currentDB, dbHashTableSize, expiryHashTableSize)

		case model.AUX:
			aux_key, err := parse.DecodeString(rdb_stream)
			if err != nil {
				return fmt.Errorf("failed to decode AUX key: %w", err)
			}
			aux_value, err := parse.DecodeString(rdb_stream)
			if err != nil {
				return fmt.Errorf("failed to decode AUX value: %w", err)
			}
			fmt.Printf("AUX - %s: %s\n", aux_key, aux_value)

		case model.EXPIRETIME:
			// 4 字节，秒级时间戳
			expire_time_buf := make([]byte, 4)
			if _, err := io.ReadFull(rdb_stream, expire_time_buf); err != nil {
				return fmt.Errorf("failed to read expiry time: %w", err)
			}
			expiryTime = int64(binary.LittleEndian.Uint32(expire_time_buf))
			expiryInMs = false
			// 注意：不要在这里 break，继续读取后面的值类型

		case model.EXPIRETIMEMS:
			// 8 字节，毫秒级时间戳
			expire_time_buf := make([]byte, 8)
			if _, err := io.ReadFull(rdb_stream, expire_time_buf); err != nil {
				return fmt.Errorf("failed to read expiry time ms: %w", err)
			}
			expiryTime = int64(binary.LittleEndian.Uint64(expire_time_buf))
			expiryInMs = true
			// 注意：不要在这里 break，继续读取后面的值类型

			// Service Details Requirement
			response.ExpiryMillis = expiryTime

		default:
			// 这是一个值类型，需要读取 key 和 value
			valueType := b
			if err := parseKeyValue(rdb_stream, valueType, expiryTime, expiryInMs, checkTypeOf); err != nil {
				return err
			}
			// 重置过期时间
			expiryTime = -1
			expiryInMs = false
		}
	}
}

func parseKeyValue(rdb_stream *strings.Reader, valueType byte, expiryTime int64, expiryInMs bool, checkTypeOf string) error {
	// 读取 key
	key, err := parse.DecodeString(rdb_stream)
	if err != nil {
		return fmt.Errorf("failed to decode key: %w", err)
	}

	// 打印过期时间信息
	expiryInfo := ""
	if expiryTime > 0 {
		if expiryInMs {
			expiryInfo = fmt.Sprintf(" [Expires at: %d ms]", expiryTime)
		} else {
			expiryInfo = fmt.Sprintf(" [Expires at: %d s]", expiryTime)
		}
	}

	// 检查是否需要报告此键的类型
	if key == checkTypeOf {
		typeName := getTypeName(valueType)
		fmt.Printf("\n*** FOUND CHECK_TYPE_OF KEY: %s => TYPE: %s ***\n\n", key, typeName)
		// Service Details Requirement
		response.Extra[checkTypeOf] = typeName
	}

	// 根据值类型解析
	switch valueType {
	case model.STRING:
		value, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode string value: %w", err)
		}
		fmt.Printf("STRING - Key: %s, Value: %s%s\n", key, value, expiryInfo)
		if _, len := utf8.DecodeRuneInString(key); len > 3 {
			response.EmojiKeyValue = value
		}

	case model.LIST:
		list, err := parse.DecodeList(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode list: %w", err)
		}
		fmt.Printf("LIST - Key: %s, Size: %d%s\n", key, len(list), expiryInfo)
		for i, value := range list {
			fmt.Printf("  [%d]: %s\n", i, value)
		}

	case model.SET:
		set, err := parse.DecodeList(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode set: %w", err)
		}
		fmt.Printf("SET - Key: %s, Size: %d%s\n", key, len(set), expiryInfo)
		for _, value := range set {
			fmt.Printf("  - %s\n", value)
		}

	case model.SORTED_SET:
		sortedSet, err := parse.DecodeSortedSet(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode sorted set: %w", err)
		}
		fmt.Printf("SORTED_SET - Key: %s, Size: %d%s\n", key, len(sortedSet), expiryInfo)

	case model.HASH:
		hashMap, err := parse.DecodeHash(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode hash: %w", err)
		}
		fmt.Printf("HASH - Key: %s, Size: %d%s\n", key, len(hashMap), expiryInfo)
		for field, value := range hashMap {
			fmt.Printf("  %s => %s\n", field, value)
		}

	case model.ZIPMAP:
		// Zipmap 被包装在字符串中
		zipmapStr, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode zipmap string: %w", err)
		}
		zipmapReader := strings.NewReader(zipmapStr)
		hashMap, err := parse.DecodeZipmap(zipmapReader)
		if err != nil {
			return fmt.Errorf("failed to decode zipmap: %w", err)
		}
		fmt.Printf("ZIPMAP - Key: %s, Size: %d%s\n", key, len(hashMap), expiryInfo)
		for field, value := range hashMap {
			fmt.Printf("  %s => %s\n", field, value)
		}

	case model.ZIPLIST:
		// Ziplist 被包装在字符串中
		ziplistStr, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode ziplist string: %w", err)
		}
		ziplistReader := strings.NewReader(ziplistStr)
		data, err := parse.DecodeZiplist(ziplistReader)
		if err != nil {
			return fmt.Errorf("failed to decode ziplist: %w", err)
		}
		fmt.Printf("ZIPLIST - Key: %s, Size: %d%s\n", key, len(data), expiryInfo)
		for i, item := range data {
			fmt.Printf("  [%d]: %v\n", i, item)
		}

	case model.INTSET:
		// Intset 被包装在字符串中
		intsetStr, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode intset string: %w", err)
		}
		intsetReader := strings.NewReader(intsetStr)
		intSet, err := parse.DecodeIntset(intsetReader)
		if err != nil {
			return fmt.Errorf("failed to decode intset: %w", err)
		}
		fmt.Printf("INTSET - Key: %s, Size: %d%s\n", key, len(intSet), expiryInfo)
		for _, num := range intSet {
			fmt.Printf("  %d\n", num)
		}

	case model.SORTED_SET_ZIPLIST:
		// Sorted Set in Ziplist
		ziplistStr, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode sorted set ziplist string: %w", err)
		}
		ziplistReader := strings.NewReader(ziplistStr)
		ziplistData, err := parse.DecodeZiplist(ziplistReader)
		if err != nil {
			return fmt.Errorf("failed to decode sorted set ziplist: %w", err)
		}
		sortedSet, err := parse.DecodeSortedSetZiplist(ziplistData)
		if err != nil {
			return fmt.Errorf("failed to convert sorted set ziplist: %w", err)
		}
		fmt.Printf("SORTED_SET_ZIPLIST - Key: %s, Size: %d%s\n", key, len(sortedSet), expiryInfo)
		for member, score := range sortedSet {
			fmt.Printf("  %s => %.6f\n", member, score)
		}

	case model.HASH_ZIPLIST:
		// Hash in Ziplist
		ziplistStr, err := parse.DecodeString(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode hash ziplist string: %w", err)
		}
		ziplistReader := strings.NewReader(ziplistStr)
		ziplistData, err := parse.DecodeZiplist(ziplistReader)
		if err != nil {
			return fmt.Errorf("failed to decode hash ziplist: %w", err)
		}
		hashMap, err := parse.DecodeHashZiplist(ziplistData)
		if err != nil {
			return fmt.Errorf("failed to convert hash ziplist: %w", err)
		}
		fmt.Printf("HASH_ZIPLIST - Key: %s, Size: %d%s\n", key, len(hashMap), expiryInfo)
		for field, value := range hashMap {
			fmt.Printf("  %s => %s\n", field, value)
		}

	case model.QUICKLIST:
		list, err := parse.DecodeQuicklist(rdb_stream)
		if err != nil {
			return fmt.Errorf("failed to decode quicklist: %w", err)
		}
		fmt.Printf("QUICKLIST - Key: %s, Size: %d%s\n", key, len(list), expiryInfo)
		for i, item := range list {
			fmt.Printf("  [%d]: %v\n", i, item)
		}

	default:
		return fmt.Errorf("unsupported value type: %d", valueType)
	}

	return nil
}

func getTypeName(valueType byte) string {
	switch valueType {
	case model.STRING:
		return "string"
	case model.LIST:
		return "list"
	case model.SET:
		return "set"
	case model.SORTED_SET:
		return "zset"
	case model.HASH:
		return "hash"
	case model.ZIPMAP:
		return "hash" // Zipmap 实际上是 hash
	case model.ZIPLIST:
		return "list" // Ziplist 可能是 list
	case model.INTSET:
		return "set" // Intset 是 set
	case model.SORTED_SET_ZIPLIST:
		return "zset"
	case model.HASH_ZIPLIST:
		return "hash"
	case model.QUICKLIST:
		return "list"
	default:
		return fmt.Sprintf("unknown(%d)", valueType)
	}
}
