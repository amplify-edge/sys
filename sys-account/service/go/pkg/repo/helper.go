package repo

import (
	"time"
)

func timestampNow() int64 {
	return time.Now().UTC().Unix()
}

