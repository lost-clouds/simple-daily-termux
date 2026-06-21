package idgen

import (
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"time"
)

func New() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		panic("idgen: crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b) + "-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
