package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"password_hashing/utils"
)

const (
	accessToken = "aa8e14a982d0994e" // 请在这里填入你的 access token
)

type Solution struct {
	SHA256 string `json:"sha256"`
	HMAC   string `json:"hmac"`
	PBKDF2 string `json:"pbkdf2"`
	Scrypt string `json:"scrypt"`
}

func main() {
	token := accessToken
	if token == "" {
		token = os.Getenv("HACKATTIC_TOKEN")
		if token == "" {
			log.Fatal("请设置 access token，可以在代码中设置或通过环境变量 HACKATTIC_TOKEN")
		}
	}

	// 1. 获取问题集
	fmt.Println("正在获取问题集...")
	problem, err := utils.GetProblemSet(token)
	if err != nil {
		log.Fatalf("获取问题集失败: %v", err)
	}

	fmt.Printf("密码: %s\n", problem.Password)
	fmt.Printf("盐(base64): %s\n", problem.Salt)

	// 解码 salt (base64 -> bytes)
	salt, err := base64.StdEncoding.DecodeString(problem.Salt)
	if err != nil {
		log.Fatalf("解码 salt 失败: %v", err)
	}

	// 2. 计算 SHA256
	fmt.Println("\n计算 SHA256...")
	sha256Result := utils.EncodeToSHA256([]byte(problem.Password))
	sha256Hex := hex.EncodeToString(sha256Result)
	fmt.Printf("SHA256: %s\n", sha256Hex)

	// 3. 计算 HMAC-SHA256
	fmt.Println("\n计算 HMAC-SHA256...")
	hmacResult := utils.EncodeToHMACSHA256(salt, problem.Password)
	hmacHex := hex.EncodeToString(hmacResult)
	fmt.Printf("HMAC: %s\n", hmacHex)

	// 4. 计算 PBKDF2
	fmt.Println("\n计算 PBKDF2...")
	pbkdf2Rounds := problem.Pbkdf2.Rounds

	fmt.Printf("PBKDF2 参数: hash=%s, rounds=%d\n", problem.Pbkdf2.Hash, pbkdf2Rounds)

	// PBKDF2-SHA256 输出长度为 32 字节
	pbkdf2Result := utils.EncodePBKDF2(problem.Password, salt, pbkdf2Rounds, 32)
	pbkdf2Hex := hex.EncodeToString(pbkdf2Result)
	fmt.Printf("PBKDF2: %s\n", pbkdf2Hex)

	// 5. 计算 scrypt
	fmt.Println("\n计算 scrypt...")
	scryptN := problem.Scrypt.N
	scryptP := problem.Scrypt.P
	scryptR := problem.Scrypt.R
	scryptBuflen := problem.Scrypt.Buflen

	fmt.Printf("scrypt 参数: N=%d, p=%d, r=%d, buflen=%d\n", scryptN, scryptP, scryptR, scryptBuflen)
	fmt.Printf("scrypt 示例控制值: %s\n", problem.Scrypt.Example)

	scryptResult := utils.EncodeToScrypt(problem.Password, salt, scryptN, scryptR, scryptP, scryptBuflen)
	scryptHex := hex.EncodeToString(scryptResult)
	fmt.Printf("scrypt: %s\n", scryptHex)

	// 6. 提交答案
	solution := Solution{
		SHA256: sha256Hex,
		HMAC:   hmacHex,
		PBKDF2: pbkdf2Hex,
		Scrypt: scryptHex,
	}

	fmt.Println("\n准备提交答案...")
	err = submitSolution(token, solution)
	if err != nil {
		log.Fatalf("提交答案失败: %v", err)
	}

	fmt.Println("\n✅ 挑战完成！")
}

func submitSolution(token string, solution Solution) error {
	url := fmt.Sprintf("https://hackattic.com/challenges/password_hashing/solve?access_token=%s", token)

	jsonData, err := json.Marshal(solution)
	if err != nil {
		return fmt.Errorf("序列化答案失败: %w", err)
	}

	fmt.Printf("提交的答案:\n%s\n", string(jsonData))

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("提交请求失败: %w", err)
	}
	defer resp.Body.Close()

	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)

	fmt.Printf("\n响应状态: %s\n", resp.Status)
	fmt.Printf("响应内容: %s\n", string(body[:n]))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("提交失败，状态码: %d", resp.StatusCode)
	}

	return nil
}
