// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package main

import (
	"context"
	"fmt"
	"github.com/mdhender/ottomap/pkg/reports/dao"
	"github.com/mdhender/ottomap/pkg/simba"
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
		agent, err := simba.NewAgent("data/simba.mdb", context.Background())
		if err != nil {
			log.Fatal(err)
		}

		s, err := server.New(
			server.WithHost("localhost"),
			server.WithPort("3030"),
			server.WithRoot("."),
			server.WithTemplates("../templates"),
			server.WithPublic("../public"),
			server.WithPolicyAgent(agent),
			server.WithReportsStore(reports.NewStore()),
		)
		if err != nil {
			log.Fatal(err)
		}

		s.Routes()
		s.ShowMeSomeRoutes()

		log.Printf("serve: listening on %s\n", s.BaseURL())
		return http.ListenAndServe(s.Addr, s.Router())
	},
}
