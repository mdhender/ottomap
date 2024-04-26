// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

type Movements struct {
	Steps  []*Step
	Failed struct {
		Direction string
		Edge      string
		Terrain   string
		RawText   string
	}
	Found []string
}

type Step struct {
	Direction  string
	Terrain    string
	Edges      [6]string
	Settlement string
	RawText    string
}
