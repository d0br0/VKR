package main

import (
	"net/url"
)

// Конвертируем запрос для использование в качестве части URL
func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
