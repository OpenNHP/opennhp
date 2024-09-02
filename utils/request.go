package utils

import (
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	// ContentType ContentType
	ContentType string = "application/x-www-form-urlencoded"
)

// Get access web service
func Get(url string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func Request(url, method string, str string, header map[string]string) (string, error) {
	payload := strings.NewReader(str)

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return "", err
	}

	// req.Header.Add("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Add(k, v)
	}

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
