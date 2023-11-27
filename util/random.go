package util

import (
	"fmt"
	"math/rand"
	"strings"
)

const lowerCaseAlphabet = "abcdefghijklmnopqrstuvwxyz"

func init() {
	// rand.Seed(time.Now().UnixNano())
}

func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(lowerCaseAlphabet)
	for i := 0; i < n; i++ {
		r := lowerCaseAlphabet[rand.Intn(k)]
		sb.WriteByte(r)
	}
	return sb.String()
}

func RandomOwner() string {
	return RandomString(6)
}

func RandomAmount() int64 {
	return RandomInt(0, 69420)
}

func RandomCurrency() string {
	currencies := []string{USD, CAD, EUR, GBP, MXN}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
