package utils

import "encoding/base64"

func DecodeString(s string) ([]byte, error) {
	res, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func EncodingString(s []byte) string {
	return base64.StdEncoding.EncodeToString(s)
}
