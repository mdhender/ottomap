// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"os"
)

func hashit(s string) string {
	hh := sha1.New()
	hh.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hh.Sum(nil))
}

func isdir(path string) error {
	if sb, err := os.Stat(path); err != nil {
		return err
	} else if !sb.IsDir() {
		return fmt.Errorf("%s: not a directory", path)
	}
	return nil
}

func isfile(path string) error {
	if sb, err := os.Stat(path); err != nil {
		return err
	} else if sb.IsDir() {
		return fmt.Errorf("%s: not a file", path)
	} else if !sb.Mode().IsRegular() {
		return fmt.Errorf("%s: not a file", path)
	}
	return nil
}
