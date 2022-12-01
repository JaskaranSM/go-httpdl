package httpdl

import (
	"errors"
	"fmt"
	"math/rand"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandString(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func ResponseCodeNotSuccessFullError(statusCode int) error {
	return errors.New(fmt.Sprintf("Response status is not sucessfull, statusCode=%d", statusCode))
}
