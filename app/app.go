// Copyright (c) 2024 Michael D Henderson. All rights reserved.

// Package app implements the ottomap web application
package app

import (
	"fmt"
	reports "github.com/mdhender/ottomap/pkg/reports/dao"
	"github.com/mdhender/ottomap/pkg/simba"
	"github.com/mdhender/semver"
	"log"
	"os"
	"path/filepath"
)

type App struct {
	baseURL string
	paths   struct {
		root      string
		templates string
	}
	debug    bool
	dateFmt  string
	policies *simba.Agent
	stores   struct {
		// todo: maybe might should be the interfaces
		reports *reports.Store
	}
	version semver.Version
}

func New(options ...Option) (*App, error) {
	a := &App{}
	a.paths.root = "."
	a.paths.templates = "templates"

	for _, opt := range options {
		if err := opt(a); err != nil {
			return nil, err
		}
	}

	if err := isdir(a.paths.root); err != nil {
		return nil, err
	} else if err = isdir(a.paths.templates); err != nil {
		return nil, err
	}
	log.Printf("app: root      is %s\n", a.paths.root)
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

func WithReportsStore(rs *reports.Store) Option {
	return func(a *App) error {
		a.stores.reports = rs
		return nil
	}
}

func WithRoot(path string) Option {
	return func(a *App) (err error) {
		if a.paths.root, err = filepath.Abs(path); err != nil {
			return err
		}
		if a.paths.templates, err = filepath.Abs(filepath.Join(a.paths.root, "templates")); err != nil {
			return err
		}
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(a *App) (err error) {
		if a.paths.root == "" {
			return fmt.Errorf("must set root before templates")
		}
		if a.paths.templates, err = filepath.Abs(filepath.Join(a.paths.root, path)); err != nil {
			return err
		}
		return nil
	}
}

func WithVersion(v semver.Version) Option {
	return func(a *App) (err error) {
		a.version = v
		return nil
	}
}