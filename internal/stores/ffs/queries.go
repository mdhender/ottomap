// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

func (s *Store) GetClans(id string) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

type Clan_t struct {
	Id    string   // id of the player's clan
	Turns []Turn_t // list of turns that the clan has uploaded reports for
}

type Turn_t struct {
	Id      string
	Reports []Report_t // list of reports that the clan has uploaded for this turn
}

type Report_t struct {
	Id    string
	Clan  string   // id of the clan that owns the report
	Units []Unit_t // list of units included in this report
	Map   string   // set when there is a map file
}

type Unit_t struct {
	Id          string
	CurrentHex  string
	PreviousHex string
}

func (s *Store) GetClan(uid int64) (Clan_t, error) {
	var c Clan_t

	user, err := s.queries.GetUser(s.ctx, uid)
	if err != nil {
		return c, err
	}
	c.Id = user.Clan

	return c, nil
}

func (s *Store) createSchema() error {
	if _, err := s.mdb.Exec(schema); err != nil {
		return err
	}
	return nil
}
