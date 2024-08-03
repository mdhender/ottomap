// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

func (s *Store) GetClans(id string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

type ClanDetail_t struct {
	Id      string
	Clan    string
	Maps    []string
	Reports []string
}

type Turn_t struct {
	Id string
}

type TurnDetail_t struct {
	Id    string
	Clans []string
	Maps  []string
}

type TurnReportDetails_t struct {
	Id    string
	Clan  string
	Map   string // set only if there is a single map
	Units []UnitDetails_t
}

type UnitDetails_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}

func (s *Store) GetClanDetails(id, clan string) (ClanDetail_t, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) GetTurnListing(id string) ([]Turn_t, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) GetTurnDetails(id string, turnId string) (TurnDetail_t, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) GetTurnReportDetails(id string, turnId, clanId string) (TurnReportDetails_t, error) {
	//TODO implement me
	panic("implement me")
}

func (s *Store) createSchema() error {
	if _, err := s.mdb.Exec(schema); err != nil {
		return err
	}
	return nil
}
