// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

// Listing is a list of turns that a User is allowed to view.
type Listing []Turn

func (l Listing) Len() int {
	return len(l)
}

func (l Listing) Less(i, j int) bool {
	return l[i].Less(l[j])
}

func (l Listing) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Turn is the metadata for a report.
type Turn struct {
	Id    string // turn id (e.g. 0991-02)
	Turn  string // display value for turn id formatted as YYY-MM (e.g. 901-02)
	Year  int    // year of turn (e.g. 901)
	Month int    // month of turn (e.g. 02)
	URL   string // url to turn (e.g. /turns/0901-02)
}

func (t Turn) Less(other Turn) bool {
	return t.Id < other.Id
}
