package toolkit

import (
	"crypto/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_-+"

type Tools struct{}

func (t *Tools) RandomString(size int) string {
	b := make([]byte, size)
	rand.Read(b)
	for i := 0; i < size; i++ {
		b[i] = charset[b[i]%byte(len(charset))]
	}
	return string(b)
}
