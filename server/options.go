// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"github.com/mdhender/ottomap/app"
	"github.com/mdhender/ottomap/htmx"
	"net"
)

type Options []Option
type Option func(*Server) error

func WithApp(app *app.App) Option {
	return func(s *Server) (err error) {
		s.mux, err = app.Routes()
		return err
	}
}

func WithHost(host string) Option {
	return func(s *Server) error {
		s.host = host
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}

func WithHTMX(app *htmx.HTMX) Option {
	return func(s *Server) (err error) {
		s.mux, err = app.Routes()
		return err
	}
}

func WithPort(port string) Option {
	return func(s *Server) error {
		s.port = port
		s.Addr = net.JoinHostPort(s.host, s.port)
		return nil
	}
}
