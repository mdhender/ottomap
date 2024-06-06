// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package htmx implements the server for the HTMX client.
package htmx

import (
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"log"
	"math/rand/v2"
	"time"
)

type HTMX struct {
	db    *sqlc.DB
	paths struct {
		public    string // path to public files
		templates string // path to template files
	}
	rum    *ReportUploadMutex
	stores struct {
		db   *sqlc.DB
		auth *AuthStore
	}
}

type ReportUploadMutex struct {
	fileDataChan chan *FileUploadMetadata
	// semaphore to limit the number of concurrent upload requests
	sem chan struct{}
}

type FileUploadMetadata struct {
	Id       string // uuid
	Clan     string
	Name     string // can't trust
	Length   int
	Checksum string // SHA-256 checksum
	Data     string
}

func New(db *sqlc.DB) (*HTMX, error) {
	h := &HTMX{
		db: db,
	}
	h.stores.db = db
	h.stores.auth = &AuthStore{}

	if path, err := db.ReadMetadataPublic(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("abspath"))
	} else {
		h.paths.public = path
	}

	if path, err := db.ReadMetadataTemplates(); err != nil {
		return nil, errors.Join(err, fmt.Errorf("abspath"))
	} else {
		h.paths.templates = path
	}

	// create the upload channel along with a semaphore to limit the number of concurrent requests
	const maxConcurrentUploadRequests = 2
	h.rum = &ReportUploadMutex{
		fileDataChan: make(chan *FileUploadMetadata, maxConcurrentUploadRequests),
		sem:          make(chan struct{}, maxConcurrentUploadRequests),
	}

	// fake a consumer for the moment
	// Consume the file data from the channel in a separate goroutine
	go func() {
		for fum := range h.rum.fileDataChan {
			started := time.Now()
			log.Printf("rum: clan %q: file %q: length %d (%d)\n", fum.Clan, fum.Name, fum.Length, len(fum.Data))
			// Simulate a load by introducing a random delay between 3 and 5 seconds
			delay := time.Duration(rand.IntN(2)+3) * time.Second
			time.Sleep(delay)
			log.Printf("rum: clan %q: file %q: processed in %v\n", fum.Clan, fum.Name, time.Now().Sub(started))
		}
	}()

	return h, nil
}
