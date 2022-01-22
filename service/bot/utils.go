package bot

import (
	"crypto/sha1" // #nosec G505 not used for cryptographic purposes
	"encoding/hex"
	"os"
	"path/filepath"
)

// sha1string creates the HEX representation of the SHA1 of a string
func sha1string(str string) string {
	s := sha1.New() // #nosec G401 not used for cryptographic purposes
	_, _ = s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

// removeContents delete all contents in a given directory, recursively
func removeContents(dir string) error {
	d, err := os.Open(dir) // #nosec: G304 - The path is from a configuration, not from the user
	if err != nil {
		return err
	}
	defer func() {
		_ = d.Close()
	}()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}
	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}
	return nil
}
