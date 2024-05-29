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

func WithSessions(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before sessions")
		}
		if s.sessions.path, err = filepath.Abs(filepath.Join(s.app.paths.root, path)); err != nil {
			return err
		}
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

func WithUsers(path string) Option {
	return func(s *Server) (err error) {
		if s.app.paths.root == "" {
			return fmt.Errorf("must set root before users")
		}
		if s.users.path, err = filepath.Abs(filepath.Join(s.app.paths.root, path)); err != nil {
			return err
		}
		return nil
	}
}
