package generator

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Generator struct{}

func NewGenerator() *Generator {
	return &Generator{}
}

func (g *Generator) GenerateShortCode(length int) string {
	if length <= 0 {
		length = 6
	}

	var sb strings.Builder
	alphabetLen := big.NewInt(int64(len(alphabet)))

	for i := 0; i < length; i++ {
		n, _ := rand.Int(rand.Reader, alphabetLen)
		sb.WriteByte(alphabet[n.Int64()])
	}

	return sb.String()
}

func (g *Generator) ValidateCustomCode(code string) bool {
	if len(code) < 3 || len(code) > 20 {
		return false
	}

	for _, c := range code {
		if !strings.ContainsRune(alphabet, c) {
			return false
		}
	}

	return true
}
