package utils

import "math/rand"

func GenerateRandomString(length int) string {
	var charSet string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	for i := range bytes {
		bytes[i] = charSet[rand.Intn(len(charSet))]
	}
	return string(bytes)
}