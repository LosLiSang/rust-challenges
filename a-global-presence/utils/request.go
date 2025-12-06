package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type TokenResponse struct {
	PresenceToken string `json:"presence_token"`
}

func RequestToken(access_token string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	res, err := client.Get("https://hackattic.com/challenges/a_global_presence/problem?access_token=" + access_token)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		if body, err := io.ReadAll(res.Body); err == nil {
			return "", fmt.Errorf("API 返回错误状态码 %d: %s", res.StatusCode, string(body))
		}
		return "", fmt.Errorf("API 返回错误状态码 %d", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	var tokenResp TokenResponse
	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", err
	}
	return tokenResp.PresenceToken, nil
}

func RequestPresence(presence_token string, proxy string) (string, error) {

	var transport *http.Transport
	if proxy == "" {
		transport = &http.Transport{}
	} else {
		url, err := url.Parse("http://" + proxy)
		if err != nil {
			return "", err
		}

		transport = &http.Transport{
			Proxy: http.ProxyURL(url),
		}
	}
	client := &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
	}

	res, err := client.Get("https://hackattic.com/_/presence/" + presence_token)
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		if body, err := io.ReadAll(res.Body); err == nil {
			return "", fmt.Errorf("API 返回错误状态码 %d: %s", res.StatusCode, string(body))
		}
		return "", fmt.Errorf("API 返回错误状态码 %d", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func SubmitSolution(access_token string) (string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	marshledJson, err := json.Marshal(map[string]string{})
	if err != nil {
		return "", err
	}
	res, err := client.Post("https://hackattic.com/challenges/a_global_presence/solve?access_token="+access_token, "application/json", bytes.NewBuffer(marshledJson))
	if err != nil {
		return "", err
	}
	if res.StatusCode != http.StatusOK {
		if body, err := io.ReadAll(res.Body); err == nil {
			return "", fmt.Errorf("API 返回错误状态码 %d: %s", res.StatusCode, string(body))
		}
		return "", fmt.Errorf("API 返回错误状态码 %d", res.StatusCode)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	
	return string(body), nil
}
