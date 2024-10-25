package requests

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
)

func NewConstantRequest(ctx context.Context, url string) (*http.Request, error) {
	var (
		err error
		req *http.Request
	)
	b := make([]byte, 1024*1024)

	if req, err = http.NewRequestWithContext(ctx, "GET", url, bytes.NewReader(b)); err != nil {
		fmt.Println(err, "unable to generate constant request")
	}
	req.Header.Set("User-Agent", "MyConstantAgent")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func NewVariableRequest(ctx context.Context, url string) (*http.Request, error) {
	var (
		err error
		req *http.Request
	)

	// minimum request size to avoid excessively small requests
	minSize := 1024             // 1 KB
	maxSize := 15 * 1024 * 1024 // 15 MB

	randomSize := rand.Intn(maxSize-minSize+1) + minSize

	b := make([]byte, randomSize)

	if req, err = http.NewRequestWithContext(ctx, "GET", url, bytes.NewReader(b)); err != nil {
		fmt.Println(err, "unable to generate constant request")
	}
	req.Header.Set("User-Agent", "MyVariableAgent")
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
