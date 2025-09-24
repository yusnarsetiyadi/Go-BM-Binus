package encoding

import (
	"encoding/base64"
	"errors"
	"net/url"
)

func Encode(str string) (data string) {
	data = reverse(str)
	data = base64.StdEncoding.EncodeToString([]byte(data))
	data = url.PathEscape(data)
	return url.QueryEscape(data)
}

func Decode(str string) (data string, err error) {
	if data, err = url.QueryUnescape(str); err != nil {
		return "", errors.New("failed to decode URL")
	}

	if data, err = url.PathUnescape(data); err != nil {
		return "", errors.New("failed to decode URL")
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", errors.New("failed to decode base64")
	}
	data = string(decodedBytes)

	return reverse(data), nil
}

// function, which takes a string as
// argument and return the reverse of string.
func reverse(s string) string {
	rns := []rune(s) // convert to rune
	for i, j := 0, len(rns)-1; i < j; i, j = i+1, j-1 {
		// swap the letters of the string,
		// like first with last and so on.
		rns[i], rns[j] = rns[j], rns[i]
	}

	// return the reversed string.
	return string(rns)
}
