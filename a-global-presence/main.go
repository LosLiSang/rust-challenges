package main

import (
	"a-global-presence/utils"
	"bufio"
	"log"
	"os"
	"time"
)

func main() {
	accsess_token := "aa8e14a982d0994e"
	token, err := utils.RequestToken(accsess_token)
	if err != nil {
		log.Panicf("%v", err)
	}
	log.Printf("Presence Token: %s", token)

	proxyFilePath := "proxy.txt"
	proxyFile, err := os.Open(proxyFilePath)
	if err != nil {
		log.Panicf("无法打开代理文件: %v", err)
	}
	defer proxyFile.Close()

	var proxyList []string
	scanner := bufio.NewScanner(proxyFile)
	for scanner.Scan() {
		proxyList = append(proxyList, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Panicf("读取代理文件时出错: %v", err)
	}
	proxyList = append(proxyList, "")
	for _, proxy := range proxyList {
		go func(proxy string) {
			countryCode, err := utils.RequestPresence(token, proxy)
			if err != nil {
				log.Printf("使用代理 %s 请求失败: %v\n", proxy, err)
			}else {
				log.Printf("使用代理 %s 成功获取国家代码: %s\n", proxy, countryCode)
			}
		}(proxy)

	}
	time.Sleep(time.Second * 20)
	res, err := utils.SubmitSolution(accsess_token)
	log.Printf("提交结果 %v", res)
}
