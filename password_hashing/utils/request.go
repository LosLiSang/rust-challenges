package utils

import (
	"encoding/json"
	"net/http"
)

type ProblemResponse struct {
	Password string `json:"password"`
	Salt     string `json:"salt"`
	Pbkdf2   struct {
		Hash   string `json:"hash"`
		Rounds int `json:"rounds"`
	} `json:"pbkdf2"`
	Scrypt struct {
		N       int `json:"N"`
		P       int `json:"p"`
		R       int `json:"r"`
		Buflen  int `json:"buflen"`
		Example string `json:"_control"` // example scrypt calculated for password="rosebud", salt="pepper", N=128, p=8, n=4
	} `json:"scrypt"`
}

var client = http.Client{}

func GetProblemSet(accessToken string) (ProblemResponse, error) {
	url := "https://hackattic.com/challenges/password_hashing/problem?access_token=" + accessToken
	res, err := client.Get(url)
	var problemResponse ProblemResponse
	if err != nil {
		return problemResponse, err
	}
	defer res.Body.Close()
	buf := make([]byte, res.ContentLength)
	res.Body.Read(buf)
	err = json.Unmarshal(buf, &problemResponse)
	if err != nil {
		return problemResponse, err
	}
	return problemResponse, nil
}
