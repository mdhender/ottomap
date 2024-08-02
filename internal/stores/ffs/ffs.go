// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package ffs

//go:generate sqlc generate

import (
	"context"
	"crypto/sha256"
	"database/sql"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/internal/stores/ffs/sqlc"
	"log"
	_ "modernc.org/sqlite"
	"os"
	"path/filepath"
	"regexp"
)

var (
	//go:embed schema.sql
	schema string
)

type Store struct {
	path    string        // path to the store and data files
	file    string        // path to the store file
	mdb     *sql.DB       // the in-memory database
	queries *sqlc.Queries // the sqlc database query functions
	ctx     context.Context
}

func New(options ...Option) (*Store, error) {
	s := &Store{
		ctx: context.Background(),
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, err
		}
	}

	if s.path == "" {
		return nil, fmt.Errorf("path not set")
	} else if sb, err := os.Stat(s.path); err != nil {
		return nil, err
	} else if !sb.IsDir() {
		return nil, fmt.Errorf("%s: not a directory", s.path)
	}
	s.file = filepath.Join(s.path, "store.db")
	log.Printf("ffs: store: %s\n", s.file)
	_ = os.Remove(s.file)

	if mdb, err := sql.Open("sqlite", s.file); err != nil {
		return nil, err
	} else {
		s.mdb = mdb
	}
	// todo: uncomment this when the schema is fixed and we are saving the store to disk
	// defer func() {
	//	if s.mdb != nil {
	//		_ = s.mdb.Close()
	//	}
	// }()

	// create the schema
	if err := s.createSchema(); err != nil {
		return nil, errors.Join(ErrCreateSchema, err)
	}

	// confirm that the database has foreign keys enabled
	if rslt, err := s.mdb.Exec("PRAGMA" + " foreign_keys = ON"); err != nil {
		log.Printf("error: foreign keys are disabled\n")
		return nil, ErrForeignKeysDisabled
	} else if rslt == nil {
		log.Printf("error: foreign keys pragma failed\n")
		return nil, ErrPragmaReturnedNil
	}

	// compile the regular expressions that we'll use when processing the files
	rxMagicKey, err := regexp.Compile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	if err != nil {
		return nil, err
	}
	rxTurnReports, err := regexp.Compile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.report\.txt`)
	if err != nil {
		return nil, err
	}
	rxTurnMap, err := regexp.Compile(`^([0-9]{4}-[0-9]{2})\.([0-9]{4})\.wxx`)
	if err != nil {
		return nil, err
	}
	_, _, _ = rxMagicKey, rxTurnReports, rxTurnMap

	// create the sqlc interface to our database
	s.queries = sqlc.New(s.mdb)

	// find all paths in the root directory that contain clan data
	entries, err := os.ReadDir(s.path)
	if err != nil {
		log.Printf("ffs: readRoot: %v\n", err)
		return nil, err
	}
	for _, entry := range entries {
		// is the entry a directory and is it a valid magic key?
		if !entry.IsDir() || !rxMagicKey.MatchString(entry.Name()) {
			continue
		}
		log.Printf("ffs: %q: found key\n", entry.Name())
		// does the entry contain a clan file
		var clan struct {
			Id   string
			Clan string
		}
		keyPath := filepath.Join(s.path, entry.Name())
		if data, err := os.ReadFile(filepath.Join(keyPath, "clan.json")); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		} else if err = json.Unmarshal(data, &clan); err != nil {
			log.Printf("warn: %q: %v\n", entry.Name(), err)
			continue
		} else if clan.Id != entry.Name() {
			log.Printf("warn: %q: clan.json: id mismatch\n", entry.Name())
			continue
		}

		// create a fake user for the clan with hashed password for authentication and session management
		hash := sha256.Sum256([]byte(entry.Name()))
		hashStr := hex.EncodeToString(hash[:])
		uid, err := s.queries.CreateUser(s.ctx, sqlc.CreateUserParams{
			Clan:           clan.Clan,
			Handle:         clan.Clan,
			HashedPassword: hashStr,
			MagicKey:       clan.Id,
			Path:           keyPath,
		})
		if err != nil {
			log.Printf("ffs: %q: %v\n", clan.Id, err)
			continue
		}
		log.Printf("ffs: user %d: key %q\n", uid, clan.Id)

		//sm.sessions[hashStr] = session_t{
		//	clan:    clan.Clan,
		//	id:      entry.Name(),
		//	key:     hashStr,
		//	expires: time.Now().Add(sm.ttl),
		//}
		//log.Printf("session: load %q -> %q\n", entry.Name(), hashStr)
	}

	return s, nil
}

//func (s *Store) GetClans(id string) ([]string, error)                              {}
//func (s *Store) GetClanDetails(id, clan string) (ClanDetail_t, error)          {}
//func (s *Store) GetTurnListing(id string) ([]Turn_t, error)                    {}
//func (s *Store) GetTurnDetails(id string, turnId string) (TurnDetail_t, error) {}
//func (s *Store) GetTurnReportDetails(id string, turnId, clanId string) (TurnReportDetails_t, error) {
//}

type Option func(*Store) error

func WithPath(path string) Option {
	return func(s *Store) error {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		} else if sb, err := os.Stat(absPath); err != nil {
			return err
		} else if !sb.IsDir() {
			return fmt.Errorf("%s: not a directory", absPath)
		}
		s.path = absPath
		return nil
	}
}

func (s *Store) createSchema() error {
	if _, err := s.mdb.Exec(schema); err != nil {
		return err
	}
	return nil
}

func (s *Store) Close() error {
	if s.mdb != nil {
		return s.mdb.Close()
	}
	return nil
}
