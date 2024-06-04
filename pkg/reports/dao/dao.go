// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package reports

import (
	domain "github.com/mdhender/ottomap/pkg/reports/domain"
)

// Store is a mock implementation of ReportListingRepository.
type Store struct {
	reports []domain.Report
}

func NewStore() *Store {
	return &Store{
		reports: []domain.Report{
			{"900-06.0991", "900-06", "0991", "Pending"},
			{"900-05.0991", "900-05", "0991", "Complete"},
			{"900-04.0991", "900-04", "0991", "Complete"},
			{"900-03.0991", "900-03", "0991", "Complete"},
			{"900-02.0991", "900-02", "0991", "Complete"},
			{"900-01.0991", "900-01", "0991", "Complete"},
			{"899-12.0991", "899-12", "0991", "Complete"},
		},
	}
}

func (s *Store) AllReports(authorized func(r domain.Report) bool) (domain.Listing, error) {
	var list domain.Listing
	for _, report := range s.reports {
		if authorized(report) {
			list = append(list, report)
		}
	}
	return list, nil
}
