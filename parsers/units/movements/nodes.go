// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package movements

type Movements struct {
	Moves  []*Movement
	Failed struct {
		Direction string
		Edge      string
		Terrain   string
		Text      string
	}
	Found []string
}

type Movement struct {
	Direction string
	Result    string
	Found     []string
	Raw       string
}
