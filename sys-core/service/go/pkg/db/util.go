package db

import (
	"crypto/md5"
)

func MD5(str string) []byte {
	h := md5.New()
	h.Write([]byte(str))
	return h.Sum(nil)
}
