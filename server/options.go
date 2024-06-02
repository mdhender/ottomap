// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"fmt"
	reports "github.com/mdhender/ottomap/pkg/reports/dao"
	"github.com/mdhender/ottomap/pkg/simba"
	"net"
	"path/filepath"
)

type Options []Option
type Option func(*Server) error

func WithCSS(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.public == "" {
			return fmt.Errorf("must set public before css")
		}
		if s.app.paths.css, err = filepath.Abs(filepath.Join(s.app.paths.public, path)); err != nil {
			return err
		}
		return nil
	}
}

func WithHost(host string) Option {
	return func(s *Server) error {
		s.host = host
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPolicyAgent(a *simba.Agent) Option {
	return func(s *Server) error {
		s.app.policies = a
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPublic(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before public")
		}
		if s.app.paths.public, err = filepath.Abs(filepath.Join(s.app.paths.root, path)); err != nil {
			return err
		}
		if s.app.paths.css, err = filepath.Abs(filepath.Join(s.app.paths.public, "css")); err != nil {
			return err
		}
		return nil
	}
}

func WithReportsStore(rs *reports.Store) Option {
	return func(s *Server) error {
		s.app.stores.reports = rs
		return nil
	}
}

func WithRoot(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.root, err = filepath.Abs(path); err != nil {
			return err
		}
		if s.app.paths.public, err = filepath.Abs(filepath.Join(s.app.paths.root, "public")); err != nil {
			return err
		}
		if s.app.paths.css, err = filepath.Abs(filepath.Join(s.app.paths.public, "css")); err != nil {
			return err
		}
		if s.app.paths.templates, err = filepath.Abs(filepath.Join(s.app.paths.root, "templates")); err != nil {
			return err
		}
		return nil
	}
}

func WithTemplates(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before templates")
		}
		if s.app.paths.templates, err = filepath.Abs(filepath.Join(s.app.paths.root, path)); err != nil {
			return err
		}
		return nil
	}
}
