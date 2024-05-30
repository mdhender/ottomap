// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mdhender/ottomap/cerrs"
	"github.com/mdhender/ottomap/reports"
	"log"
	"os"
	"sort"
	"strings"
)

// Config is the application configuration.
type Config struct {
	Path       string `json:"path,omitempty"`   // path to the application configuration file
	OutputPath string `json:"output,omitempty"` // path to create output files in
	Inputs     struct {
		TurnId              string `json:"-"` // turn to process (yyyy-mm format)
		Year                int    `json:"-"` // year to process
		Month               int    `json:"-"` // month to process
		ClanId              string `json:"-"` // clan to process
		GridOriginId        string `json:"-"` // grid id of the origin
		ShowIgnoredReports  bool   `json:"-"` // show ignored reports
		ShowSkippedSections bool   `json:"-"` // show skipped sections
		ShowSteps           bool   `json:"-"` // show all steps in the parser
	} `json:"-"`
	Reports reports.Reports `json:"reports,omitempty"` // list of report files we have loaded
}

// Load loads the configuration file.
// It is the same as Read but returns errors.
func Load(path string) (*Config, error) {
	cfg := &Config{}
	if data, err := os.ReadFile(path); err != nil {
		return nil, err
	} else if err = json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	cfg.Path = path

	if strings.TrimSpace(cfg.OutputPath) != cfg.OutputPath {
		log.Printf("config: output: %q\n", cfg.OutputPath)
		return nil, errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("leading or trailing spaces"))
	} else if sb, err := os.Stat(cfg.OutputPath); err != nil {
		log.Printf("config: output: %q\n", cfg.OutputPath)
		return nil, errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, err)
	} else if !sb.IsDir() {
		log.Printf("config: output: %q\n", cfg.OutputPath)
		return nil, errors.Join(cerrs.ErrInvalidPath, cerrs.ErrInvalidOutputPath, fmt.Errorf("output path is not a directory"))
	}

	log.Printf("config: loaded %s\n", cfg.Path)
	return cfg, nil
}

// Read reads the configuration file.
// It is of dubious value if the configuration file does not exist.
// And maybe if it does. There is no way to tell.
// If the file does exist, then reports from the configuration file
// are merged into this configuration's reports list if they're not already there.
func (c *Config) Read() {
	var oldConfig Config
	if sb, err := os.Stat(c.Path); err != nil && os.IsNotExist(err) {
		log.Printf("config: warning: %s does not exist\n", c.Path)
	} else if sb.IsDir() {
		log.Fatalf("config: error: %s is a folder\n", c.Path)
	} else if data, err := os.ReadFile(c.Path); err != nil {
		log.Fatalf("config: error: %v\n", err)
	} else if err = json.Unmarshal(data, &oldConfig); err != nil {
		log.Fatalf("config: error: %v\n", err)
	} else {
		if c.Path != oldConfig.Path {
			log.Printf("config: warning: config path changed!\n")
		}
		if c.OutputPath == "" && oldConfig.OutputPath != "" {
			c.OutputPath = oldConfig.OutputPath
			log.Printf("config: output now %s\n", c.OutputPath)
		}
		for _, rpt := range oldConfig.Reports {
			if c.Reports.Contains(rpt.Id) {
				continue
			}
			c.Reports = append(c.Reports, rpt)
		}
	}

	// always sort the reports
	sort.Sort(c.Reports)

	log.Printf("config: loaded: %s\n", c.Path)
	log.Printf("config: output: %s\n", c.OutputPath)
}

func (c *Config) Save() {
	// always sort the reports
	sort.Sort(c.Reports)

	if data, err := json.MarshalIndent(c, "", "\t"); err != nil {
		log.Printf("config: error: %v\n", err)
	} else if err = os.WriteFile(c.Path, data, 0644); err != nil {
		log.Printf("config: error: %v\n", err)
	}
	log.Printf("config: saved %s\n", c.Path)
}

func (c *Config) AddReport(rpt *reports.Report) {
	if c.Reports.Contains(rpt.Id) { // replace existing report
		for i, r := range c.Reports {
			if r.Id == rpt.Id {
				c.Reports[i] = rpt
				log.Printf("config: %s: replaced %s\n", rpt.Id, rpt.Path)
				return
			}
		}
		return
	}
	c.Reports = append(c.Reports, rpt)
	log.Printf("config: %s: added    %s\n", rpt.Id, rpt.Path)
}
