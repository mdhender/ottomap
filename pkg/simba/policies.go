// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package simba

import (
	"github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/turns/domain"
	"log"
	"net/http"
	"time"
)

// RequestIsAuthenticated returns true if the request contains is a valid session that has not expired.
func (a *Agent) RequestIsAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("simba-session")
	if err != nil {
		log.Printf("agent: requestIsAuthenticated: no cookie\n")
		return false
	}
	return a.SessionIsAuthenticated(cookie.Value)
}

// SessionIsAuthenticated returns true if the session is valid and has not expired.
func (a *Agent) SessionIsAuthenticated(sid string) bool {
	sid, exp, err := a.db.ReadSession(sid)
	if err != nil {
		log.Printf("agent: sessionIsAuthenticated: %v\n", err)
		return false
	} else if sid == "" {
		log.Printf("agent: sessionIsAuthenticated: session: id not found\n")
		return false
	} else if !time.Now().Before(exp) {
		log.Printf("agent: sessionIsAuthenticated: session: expired\n")
		return false
	}
	return true
}

func (a *Agent) UserCanViewReport(uid string, rpt reports.Report) bool {
	clan, err := a.db.ReadUserClan(uid)
	if err != nil {
		log.Printf("agent: userCanViewReport: %v\n", err)
		return false
	} else if clan == "" {
		return false
	}
	return clan == rpt.Clan || a.UserIsAdministrator(uid)
}

// UserIsAdministrator returns true if the user is an administrator.
func (a *Agent) UserIsAdministrator(uid string) bool {
	value, err := a.db.ReadUserRole(uid, "administrator")
	if err != nil {
		log.Printf("agent: userIsAdminstrator: %v\n", err)
		return false
	} else if value == "" {
		return false
	}
	return value == "true"
}

func (a *Agent) UserReportsFilter(uid string) func(reports.Report) bool {
	clan, err := a.db.ReadUserClan(uid)
	if err != nil {
		log.Printf("agent: userReportsFilter: %v\n", err)
		return func(reports.Report) bool { return false }
	} else if clan == "" {
		return func(reports.Report) bool { return false }
	}

	return func(rpt reports.Report) bool {
		if rpt.Clan == clan {
			return true
		}
		return false
	}
}

func (a *Agent) UserTurnsFilter(uid string) func(turns.Turn) bool {
	_, _, err := a.db.ReadUser(uid)
	if err != nil {
		log.Printf("agent: userTurnsFilter: %v\n", err)
		return func(turns.Turn) bool { return false }
	}

	return func(turns.Turn) bool { return true }
}
