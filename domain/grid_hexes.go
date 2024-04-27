// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package domain

import "fmt"

func (h *GridHex) String() string {
	if h == nil || h.Grid == "" {
		return "N/A"
	}
	return fmt.Sprintf("%s %02d%02d", h.Grid, h.Column, h.Row)
}
