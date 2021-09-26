package tbot

import (
	"crypto/md5"
	"encoding/hex"
)

func sha1string(str string) string {
	s := md5.New() // #nosec G401 not used for cryptographic purposes
	_, _ = s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}
