// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package server

import (
	"github.com/mdhender/ottomap/sessions"
	"github.com/mdhender/ottomap/way"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en"><head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width">

	<title>ottomap - a mapping tool</title>

	<link rel="stylesheet" href="https://unpkg.com/missing.css@1.1.1">
	<link href="https://fonts.bunny.net/css?family=source-sans-3:400,700|m-plus-code-latin:400,700" rel="stylesheet">

	<style>
		:root {
			--main-font: "Source Sans 3", -apple-system, system-ui, sans-serif;
		}
		dfn > code {
			font-style: normal;
			text-decoration: 1px dashed var(--muted-fg) underline;
		}
		code a {
			font-family: inherit;
		}
	</style>
</head>
<body>
	<header class="navbar">
    <nav>
        <ul role="list">
            <li>
                <a href="/">ottomap</a>
            </li>
        </ul>
    </nav>
</header>

<main class="airy">
    <big-screen class="dense">
        <h1>Turn Reports to Maps</h1>
        <p>
            Ottomap is a tool that reads turn reports and creates map files.
        </p>
        <tool-bar>
            <strong><button aria-controls="signup-form" aria-expanded="false" type="button">Sign up</button></strong>
            <a class="<button>" href="#">Learn more</a>
        </tool-bar>
    </big-screen>

    <div>
        <form hidden="" id="signup-form" class="box dense absolute">
            <h4>Sign up for an account</h4>
            <div class="table rows">
                <p>
                    <label for="email-in">Email</label>
                    <input type="email" name="email" id="email-in" placeholder="you@example.com">
                </p>
                <p>
                    <label for="update-freq">Update frequency</label>
                    <radio-buttons id="update-freq">
                        <input type="radio" name="upd-freq-in" id="upd-all" checked="">
                        <label for="upd-all">All updates</label>
    
                        <input type="radio" name="upd-freq-in" id="upd-important">
                        <label for="upd-important">Most important</label>
    
                        <input type="radio" name="upd-freq-in" id="upd-weekly">
                        <label for="upd-weekly">Weekly digest</label>
                    </radio-buttons>
                </p>
            <p><button>Sign Up</button></p>
    </div></form>

    <p>
        <b class="lede">
            Ottomap is a simple tool to read turn reports and create map files.
        </b>
        Lorem ipsum dolor sit amet consectetur adipisicing elit.
        Aliquam odit animi iure autem magni molestiae architecto, earum, quaerat quisquam, totam at sequi eum. 
        Rerum ipsam consequatur autem eaque et velit?
    </p>

    <ul role="list" class="f-switch dense">
        <li class="box">
            <h2 class="<h4>">Beautiful by default</h2>
            <p>Just drop the stylesheet in and use semantic HTML.</p>
        </li>
        <li class="box">
            <h2 class="<h4>">Accessible structures</h2>
            <p>Mark up tabs and other components and we'll handle the styling.</p>
        </li>
        <li class="box">
            <h2 class="<h4>">Useful components</h2>
            <p>...such as these cards!</p>
        </li>
    </ul>
</div></main>

<footer>
    <nav class="f-switch">
        <div>
            <h4>Ottomap</h4>
            <ul role="list">
                <li><a href="#">About Us</a></li>
                <li><a href="#">Contact</a></li>
            </ul>
        </div>

        <div>
            <h4>Links</h4>
            <ul role="list">
                <li><a href="https://tribenet.wiki/">Tribenet Wiki</a></li>
                <li><a href="https://worldographer.com/">Worldographer</a></li>
            </ul>
        </div>

        <div>
            <h4>Legal</h4>
            <ul role="list">
                <li><a href="#">Privacy Statement</a></li>
                <li><a href="#">End User License Agreement</a></li>
            </ul>
        </div>
    </nav>

    <p><small>© 2024 Fictitious Vaporware Industries</small></p>
</footer>
</body>
</html>`))
	}
}

func (s *Server) getLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := sessions.User(r.Context())
		if user.IsAuthenticated {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		_, _ = w.Write([]byte(`<!DOCTYPE html>
<html lang="en"><head>
	<meta charset="utf-8">
	<meta name="viewport" content="width=device-width">

	<title>ottomap</title>

	<link rel="stylesheet" href="https://unpkg.com/missing.css@1.1.1">
	<link href="https://fonts.bunny.net/css?family=source-sans-3:400,700|m-plus-code-latin:400,700" rel="stylesheet">

	<style>
		:root {
			--main-font: "Source Sans 3", -apple-system, system-ui, sans-serif;
		}
		dfn > code {
			font-style: normal;
			text-decoration: 1px dashed var(--muted-fg) underline;
		}
		code a {
			font-family: inherit;
		}
	</style>
</head>
<body>
	<main>
<p>Login:</p><form class="box rows">
    <p>
    <label for="handle">My handle</label>
    <input type="text" id="handle" value="password">
    </p><p>
    <label for="secret">My secret</label>
    <input type="text" id="secret" value="password">
</p></form>

</main>
</body></html>`))
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// if there's a session cookie, delete it and the session
		if id, ok := s.sessions.manager.GetCookie(r); ok {
			// delete the session
			s.sessions.manager.DeleteSession(id)
			// delete the session cookie
			s.sessions.manager.DeleteCookie(w)
		}

		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<p>You have been logged out and your session has been invalidated. Please close your browser and re-open to log in again."))
	}
}

// returns a handler that will serve a static file if one exists, otherwise return not found.
func (s *Server) handleStaticFiles(prefix, root string, debug bool) http.Handler {
	log.Println("[static] initializing")
	defer log.Println("[static] initialized")

	log.Printf("[static] strip: %q\n", prefix)
	log.Printf("[static]  root: %q\n", root)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if r.Method != "GET" || !sessions.User(ctx).IsAuthenticated {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))
		if debug {
			log.Printf("[static] %q\n", file)
		}

		stat, err := os.Stat(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// only serve regular files, never directories or directory listings.
		if stat.IsDir() || !stat.Mode().IsRegular() {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		// pretty sure that we have a regular file at this point.
		rdr, err := os.Open(file)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		defer func(r io.ReadCloser) {
			_ = r.Close()
		}(rdr)

		// let Go serve the file. it does magic things like content-type, etc.
		http.ServeContent(w, r, file, stat.ModTime(), rdr)
	})
}

func (s *Server) handleVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(s.version.String()))
	}
}

func (s *Server) apiGetLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := way.Param(r.Context(), "name")
		secret := way.Param(r.Context(), "secret")

		// authenticate the user or return an error
		user, ok := s.users.store.Authenticate(name, secret)
		if !ok {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}

		// create a new session or return an error
		sessionId, ok := s.sessions.manager.CreateSession(user.Id)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// create the session cookie or return an error
		ok = s.sessions.manager.AddCookie(w, sessionId)
		if !ok {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		// add the user to the request context or return an error
		ctx := s.sessions.manager.AddUser(r.Context(), user)
		if ctx == nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r.WithContext(ctx), "/", http.StatusFound)
	}
}
