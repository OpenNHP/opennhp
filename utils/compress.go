package utils

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"io"
)

func Compression(data string) ([]byte, error) {
	var buf bytes.Buffer
	gzwriter := gzip.NewWriter(&buf)

	_, err := gzwriter.Write([]byte(data))
	if err != nil {
		return nil, err
	}

	err = gzwriter.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Decompression(data string) (string, error) {
	dataBytes, _ := base64.StdEncoding.DecodeString(data)
	reader := bytes.NewReader(dataBytes)
	gzreader, err := gzip.NewReader(reader)
	if err != nil {
		return "", err
	}
	defer gzreader.Close()

	output, err := io.ReadAll(gzreader)
	if err != nil {
		return "", err
	}

	return string(output), nil
}
