package tbot

import (
	"crypto/sha1" // #nosec G505 not used for cryptographic purposes
	"encoding/hex"
)

func sha1string(str string) string {
	s := sha1.New() // #nosec G401 not used for cryptographic purposes
	_, _ = s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}
