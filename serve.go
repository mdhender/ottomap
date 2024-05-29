// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"fmt"
	"github.com/mdhender/ottomap/server"
	"github.com/spf13/cobra"
	"log"
	"net/http"
)

var argsServe struct {
	signingKey string
}

var cmdServe = &cobra.Command{
	Use:   "serve",
	Short: "Start web server",
	Long:  `Run a web server.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if len(argsServe.signingKey) == 0 {
			return fmt.Errorf("missing signing key")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		log.Printf("serve: sigkey %q\n", argsServe.signingKey)

		s, err := server.New(
			server.WithHost("localhost"),
			server.WithPort("3030"),
			server.WithSigningKey(argsServe.signingKey),
			server.WithRoot("."),
			server.WithUsers("users.json"),
			server.WithSessions("sessions.json"),
			server.WithCookie("ottomap"),
			server.WithTemplates("../templates"),
			server.WithPublic("../public"),
		)
		if err != nil {
			log.Fatal(err)
		}

		s.Routes()
		s.ShowMeSomeRoutes()

		log.Printf("serve: listening on http://%s\n", s.Addr)
		return http.ListenAndServe(s.Addr, s.Router())
	},
}
