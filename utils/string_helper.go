package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"
	"unicode"

	uuid "github.com/satori/go.uuid"
)

var _random_once sync.Once

func GenerateKey() ([]byte, error) {
	b := make([]byte, 64) //nolint:gomnd
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func RandomString(size int) string {

	_random_once.Do(func() {
		rand.Seed(time.Now().Unix())
	})

	b := make([]byte, size)
	//	bigInt := big.NewInt(int64(length))
	for i := 0; i < size; i++ {
		b[i] = letterBytes[rand.Intn(size)]
	}
	return string(b)
}

func ToSnakeString(s string) string {
	var res []rune
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 && unicode.IsLower(rune(s[i-1])) {
				res = append(res, '_')
			}
			res = append(res, unicode.ToLower(r))
		} else {
			res = append(res, r)
		}
	}
	return string(res)
}

func Md5Sum(strToSign string) string {
	ret := md5.Sum([]byte(strToSign))
	return fmt.Sprintf("%x", ret[:])
}

func Sha1Sum(strToSign string) string {
	ret := sha1.Sum([]byte(strToSign))
	return fmt.Sprintf("%x", ret[:])
}
func NewRequestId() string {
	uuid := uuid.NewV4()
	return strings.ToUpper(uuid.String())
}
