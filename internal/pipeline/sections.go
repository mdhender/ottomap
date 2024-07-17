// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package pipeline

type FileInfo_t struct {
	Path     string // path to file
	TurnInfo *TurnInfo_t
}

type SectionContext_t struct {
	FileInfo *FileInfo_t
}

type Token_t struct{}

type TurnInfo_t struct {
	Id    string // turn id as YYYY-MM
	Year  int
	Month int
}

type UnitInfo_t struct {
	Id     string
	Parent *UnitInfo_t
}

// the section pipeline accepts tokens (defined at some level) and updates the section context.
// when we switch to a new input, we must check that it is for the following turn.
// if it is not, we must terminate the pipeline.
// when we switch to a new section, what are we doing?

// NewSectionPipeline returns a pipeline (a channel) that accepts tokens from the parser and does something with them?
func NewSectionPipeline(ctx *SectionContext_t) chan *Token_t {
	// TODO: add a section pipeline
	return nil
}
