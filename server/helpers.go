// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"crypto/sha1"
	"encoding/base64"
)

func hashit(s string) string {
	hh := sha1.New()
	hh.Write([]byte(s))
	return base64.URLEncoding.EncodeToString(hh.Sum(nil))
}
