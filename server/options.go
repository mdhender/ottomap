// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"github.com/mdhender/ottomap/app"
	"net"
	"path/filepath"
)

type Options []Option
type Option func(*Server) error

func WithApp(app *app.App) Option {
	return func(s *Server) error {
		s.mux = app.Routes()
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

func WithPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithPublic(path string) Option {
	return func(s *Server) (err error) {
		if s.public, err = filepath.Abs(path); err != nil {
			return err
		}
		return nil
	}
}
