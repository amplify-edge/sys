package delivery

import (
	"time"
)

func timestampNow() int64 {
	return time.Now().UTC().Unix()
}
