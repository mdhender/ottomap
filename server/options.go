// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"fmt"
	"net"
	"path/filepath"
)

type Options []Option
type Option func(*Server) error

func WithCookie(name string) Option {
	return func(s *Server) error {
		s.sessions.cookies.name = name
		s.sessions.cookies.secure = true
		return nil
	}
}

func WithCSS(path string) Option {
	return func(s *Server) error {
		if s.app.paths.public == "" {
			return fmt.Errorf("must set public before css")
		}
		s.app.paths.css = filepath.Join(s.app.paths.public, path)
		return nil
	}
}

func WithHost(host string) Option {
	return func(s *Server) error {
		s.app.host = host
		s.Addr = net.JoinHostPort(s.app.host, s.app.port)
		return nil
	}
}

func WithPort(port string) Option {
	return func(s *Server) error {
		s.app.port = port
		s.Addr = net.JoinHostPort(s.app.host, s.app.port)
		return nil
	}
}

func WithPublic(path string) Option {
	return func(s *Server) error {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before public")
		}
		s.app.paths.public = filepath.Join(s.app.paths.root, path)
		s.app.paths.css = filepath.Join(s.app.paths.public, "css")
		return nil
	}
}

func WithRoot(path string) Option {
	return func(s *Server) error {
		s.app.paths.root = path
		s.app.paths.public = filepath.Join(s.app.paths.root, "public")
		s.app.paths.css = filepath.Join(s.app.paths.public, "css")
		s.app.paths.templates = filepath.Join(s.app.paths.root, "templates")
		return nil
	}
}

func WithSessions(path string) Option {
	return func(s *Server) error {
		s.sessions.path = path
		return nil
	}
}

func WithSigningKey(secret string) Option {
	return func(s *Server) (err error) {
		if len(secret) == 0 {
			return fmt.Errorf("signing key is empty")
		}
		s.auth.secret = hashit(secret + hashit(secret+"ottomap"))
		return err
	}
}

func WithTemplates(path string) Option {
	return func(s *Server) error {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before templates")
		}
		s.app.paths.templates = filepath.Join(s.app.paths.root, path)
		return nil
	}
}

func WithUsers(path string) Option {
	return func(s *Server) error {
		s.users.path = path
		return nil
	}
}
