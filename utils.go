package main

import (
	"crypto/sha1" // #nosec G505 not used for cryptographic purposes
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
)

func RemoveContents(dir string) error {
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

func Sha1(str string) string {
	s := sha1.New() // #nosec G401 not used for cryptographic purposes
	_, _ = s.Write([]byte(str))
	return hex.EncodeToString(s.Sum(nil))
}

type Set map[string]bool

func (c Set) Add(str string) {
	c[str] = true
}

func (c Set) GetAsOrderedList() []string {
	var ret []string
	for k := range c {
		ret = append(ret, k)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i] < ret[j]
	})
	return ret
}
