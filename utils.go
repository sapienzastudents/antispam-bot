package main

import (
	"os"
	"path/filepath"
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
