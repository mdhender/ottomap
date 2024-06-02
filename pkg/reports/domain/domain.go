// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

type ReportListingRepository interface {
	AllReports(authorize func(rpt Report) bool) (ReportListing, error)
}

// ReportListing is a listing of reports that a User is allowed to view.
type ReportListing []Report

func (rl ReportListing) Len() int {
	return len(rl)
}

func (rl ReportListing) Less(i, j int) bool {
	return rl[i].Less(rl[j])
}

func (rl ReportListing) Swap(i, j int) {
	rl[i], rl[j] = rl[j], rl[i]
}

// Report is the metadata for a report.
type Report struct {
	Id     string // report id (e.g. 0991-02.0991)
	Turn   string // turn id formatted as YYY-MM (e.g. 901-02)
	Clan   string // clan id (e.g. 0991)
	Status string // status of report (e.g. "pending")
}

func (r Report) Less(other Report) bool {
	return r.Id < other.Id
}
