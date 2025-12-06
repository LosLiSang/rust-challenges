package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ProblemResponse 是 API 返回的问题结构
type ProblemResponse struct {
	RDB          string `json:"rdb"`
	Requirements struct {
		CheckTypeOf string `json:"check_type_of"`
	} `json:"requirements"`
}

// Solution 是提交给 API 的解答结构
type Solution struct {
	DbCount       int    `json:"db_count"`
	EmojiKeyValue string `json:"emoji_key_value"`
	ExpiryMillis  int64  `json:"expiry_millis"`
	// 动态字段用于 check_type_of 的值
	Extra map[string]interface{} `json:"-"`
}

// MarshalJSON 自定义 JSON 序列化，用于合并额外字段
func (s Solution) MarshalJSON() ([]byte, error) {
	type Alias Solution
	aux := struct {
		*Alias
	}{
		Alias: (*Alias)(&s),
	}
	b, err := json.Marshal(aux)
	if err != nil {
		return nil, err
	}

	// 将额外字段添加到 JSON 中
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return nil, err
	}

	for k, v := range s.Extra {
		m[k] = v
	}

	return json.Marshal(m)
}

// DownloadProblem 从 API 下载问题（获取 RDB 快照）
func DownloadProblem(accessToken string) (*ProblemResponse, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("https://hackattic.com/challenges/the_redis_one/problem?access_token=%s", accessToken)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应体失败: %w", err)
	}

	var problem ProblemResponse
	err = json.Unmarshal(body, &problem)
	if err != nil {
		return nil, fmt.Errorf("解析 JSON 失败: %w", err)
	}

	return &problem, nil
}

// UploadSolution 提交解答到 API
func UploadSolution(accessToken string, solution *Solution) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf("https://hackattic.com/challenges/the_redis_one/solve?access_token=%s", accessToken)

	// 序列化解答
	solutionJSON, err := json.Marshal(solution)
	if err != nil {
		return "", fmt.Errorf("序列化解答失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(solutionJSON))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("提交失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应体失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	return string(body), nil
}
