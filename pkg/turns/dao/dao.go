// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package turns

import (
	domain "github.com/mdhender/ottomap/pkg/turns/domain"
)

// Store is a mock implementation of TurnListingRepository.
type Store struct {
	turns []domain.Turn
}

func NewStore() *Store {
	return &Store{
		turns: []domain.Turn{
			{Id: "0900-06", Turn: "900-06", Year: 900, Month: 6, URL: "/turns/0900-06"},
			{Id: "0900-05", Turn: "900-05", Year: 900, Month: 5, URL: "/turns/0900-05"},
			{Id: "0900-04", Turn: "900-04", Year: 900, Month: 4, URL: "/turns/0900-04"},
			{Id: "0900-03", Turn: "900-03", Year: 900, Month: 3, URL: "/turns/0900-03"},
			{Id: "0900-02", Turn: "900-02", Year: 900, Month: 2, URL: "/turns/0900-02"},
			{Id: "0900-01", Turn: "900-01", Year: 900, Month: 1, URL: "/turns/0900-01"},
			{Id: "0899-12", Turn: "899-12", Year: 899, Month: 12, URL: "/turns/0899-12"},
		},
	}
}

func (s *Store) AllTurns(authorized func(t domain.Turn) bool) (domain.Listing, error) {
	var list domain.Listing
	for _, turn := range s.turns {
		if authorized(turn) {
			list = append(list, turn)
		}
	}
	return list, nil
}
