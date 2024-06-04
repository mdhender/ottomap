// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

// Listing is a list of reports that a User is allowed to view.
type Listing []Report

func (l Listing) Len() int {
	return len(l)
}

func (l Listing) Less(i, j int) bool {
	return l[i].Less(l[j])
}

func (l Listing) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

// Report is the metadata for a report.
type Report struct {
	Id     string // report id (e.g. 0991-02.0991)
	Turn   string // display value turn id formatted as YYY-MM (e.g. 901-02)
	Clan   string // clan id (e.g. 0991)
	Status string // status of report (e.g. "pending")
}

func (r Report) Less(other Report) bool {
	return r.Id < other.Id
}
