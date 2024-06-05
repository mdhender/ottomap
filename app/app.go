// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package app implements the ottomap web application
package app

import (
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/pkg/simba"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"github.com/mdhender/semver"
	"log"
	"os"
)

type App struct {
	baseURL string
	paths   struct {
		public    string
		templates string
	}
	db       *sqlc.DB
	debug    bool
	dateFmt  string
	policies *simba.Agent
	version  semver.Version
}

func New(options ...Option) (*App, error) {
	a := &App{}

	for _, opt := range options {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	var err error
	a.paths.public, err = a.db.ReadMetadataPublic()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("reading metadata public"))
	}
	a.paths.templates, err = a.db.ReadMetadataTemplates()
	if err != nil {
		return nil, errors.Join(err, fmt.Errorf("reading metadata templates"))
	}

	if err = isdir(a.paths.public); err != nil {
		return nil, err
	} else if err = isdir(a.paths.templates); err != nil {
		return nil, err
	}
	log.Printf("app: public    is %s\n", a.paths.public)
	log.Printf("app: templates is %s\n", a.paths.templates)

	return a, nil
}

func isdir(path string) error {
	if sb, err := os.Stat(path); err != nil {
		return err
	} else if !sb.IsDir() {
		return fmt.Errorf("%s: not a directory", path)
	}
	return nil
}

type Options []Option
type Option func(*App) error

func WithPolicyAgent(agent *simba.Agent) Option {
	return func(a *App) error {
		a.policies = agent
		return nil
	}
}

func WithStore(store *sqlc.DB) Option {
	return func(a *App) error {
		a.db = store
		return nil
	}
}

func WithVersion(v semver.Version) Option {
	return func(a *App) (err error) {
		a.version = v
		return nil
	}
}
