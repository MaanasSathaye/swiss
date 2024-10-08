package requests

import (
	"bytes"
	"fmt"
	"math/rand"
	"net/http"
)

func NewConstantRequest(url string) *http.Request {
	var (
		err error
		req *http.Request
	)
	b := make([]byte, 1024*1024)

	if req, err = http.NewRequest("GET", url, bytes.NewReader(b)); err != nil {
		fmt.Println(err, "unable to generate constant request")
	}
	req.Header.Set("User-Agent", "MyConstantAgent")
	req.Header.Set("Content-Type", "application/json")
	return req
}

func NewVariableRequest(url string) *http.Request {
	var (
		err error
		req *http.Request
	)

	// minimum request size to avoid excessively small requests
	minSize := 1024             // 1 KB
	maxSize := 30 * 1024 * 1024 // 30 MB

	randomSize := rand.Intn(maxSize-minSize+1) + minSize

	b := make([]byte, randomSize)

	if req, err = http.NewRequest("GET", url, bytes.NewReader(b)); err != nil {
		fmt.Println(err, "unable to generate constant request")
	}
	req.Header.Set("User-Agent", "MyVariableAgent")
	req.Header.Set("Content-Type", "application/json")
	return req
}
