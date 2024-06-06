// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package htmx

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
	"github.com/mdhender/ottomap/pkg/sqlc"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func (h *HTMX) Routes() (*http.ServeMux, error) {
	mux := http.NewServeMux() // default mux, no routes

	mux.HandleFunc("GET /", getClanHomePage(h.paths.templates, h.stores.auth, h.stores.db))

	//mux.HandleFunc("POST /clans/{clanId}/reports/upload", postClanReports(h.paths.templates, h.db, h.rum))
	mux.HandleFunc("POST /clans/{clanId}/reports/upload", postClanReportsUpload(h.paths.templates, h.db))
	mux.HandleFunc("GET /reports/queued", getReportsQueued(h.paths.templates, h.stores.auth, h.db))
	mux.HandleFunc("GET /reports/queued/{queueId}", getReportsQueuedDetail(h.paths.templates, h.stores.auth, h.db))
	mux.HandleFunc("DELETE /reports/queued/{queueId}", deleteReportQueuedDetail(h.stores.auth, nil))
	mux.HandleFunc("GET /reports/upload/{queueId}", getReportsUpload(h.paths.templates, h.db))

	// walk the public directory and add routes to serve static files
	validExtensions := map[string]bool{
		".css":    true,
		".html":   true,
		".ico":    true,
		".jpg":    true,
		".js":     true,
		".png":    true,
		".robots": true,
		".svg":    true,
	}
	if err := filepath.WalkDir(h.paths.public, func(path string, d os.DirEntry, err error) error {
		// don't serve unknown file types or dotfiles
		if err != nil || d.IsDir() || !validExtensions[filepath.Ext(path)] || strings.HasPrefix(filepath.Base(path), ".") {
			return nil
		}
		route := "GET " + strings.TrimPrefix(path, h.paths.public)
		//log.Printf("htmx: public: path  %q\n", path)
		log.Printf("htmx: public: route %q\n", route)
		mux.Handle(route, getPublicFiles("", h.paths.public))
		return nil
	}); err != nil {
		return nil, err
	}
	return mux, nil
}

func getClanHomePage(templatesPath string, auth AuthRepo, db interface {
	AllClanReportMetadata(cid string) ([]reports.Metadata, error)
}) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "clan_homepage.gohtml"),
		filepath.Join(templatesPath, "upload_ui.gohtml"),
	}

	type ReportDetail struct {
		Name    string
		Status  string
		Created string
		URL     string
	}

	type UploadUI struct {
		Status    string
		UploadURL string
	}

	type Payload struct {
		Page struct {
			Title string
		}
		Clan      string
		NoReports bool
		Reports   []ReportDetail
		UploadUI  UploadUIFragmentPayload
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if !auth.CurrentUser(r).IsAuthenticated() {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		clanId := auth.CurrentUser(r).ClanId()
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload Payload

		payload.Page.Title = fmt.Sprintf("Clan %s", payload.Clan)
		payload.Clan = clanId
		payload.UploadUI.Status = "waiting"
		payload.UploadUI.UploadURL = "/reports/upload"

		// fetch metadata on all clan reports
		allReportMetadata, err := db.AllClanReportMetadata(clanId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("%s: %s: clan %q: AllClanReportMetadata: %v", r.Method, r.URL.Path, clanId, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
			log.Printf("%s: %s: clan %q: allReports: no rows", r.Method, r.URL.Path, clanId)
		}
		sort.Slice(allReportMetadata, func(i, j int) bool {
			return allReportMetadata[i].Id > allReportMetadata[j].Id
		})
		for _, md := range allReportMetadata {
			payload.Reports = append(payload.Reports, ReportDetail{
				Name:    md.Name,
				Status:  md.Status,
				Created: md.Created.Format("2006-01-02"),
				URL:     fmt.Sprintf("/reports/%s", md.Id),
			})
		}
		payload.NoReports = len(payload.Reports) == 0

		render(w, r, payload, templateFiles...)
	}
}

func getReportsQueued(templatesPath string, auth AuthRepo, db interface {
	ReadQueuedReports(cid string) ([]sqlc.DBQueuedReport, error)
}) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "reports_queued_page.gohtml"),
	}

	type QueuedReport struct {
		Id      string
		Clan    string
		Status  string
		Created string
		Updated string
		URL     string
	}
	type Payload struct {
		Page struct {
			Title string
		}
		Clan            string
		NoQueuedReports bool
		Queue           []QueuedReport
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if !auth.CurrentUser(r).IsAuthenticated() {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		clanId := auth.CurrentUser(r).ClanId()
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		var payload Payload
		payload.Page.Title = "Report Queue"
		payload.Clan = clanId

		rows, err := db.ReadQueuedReports(clanId)
		if err != nil {
			log.Printf("%s: %s: clanId %q: readQueuedReports %v\n", r.Method, r.URL.Path, clanId, err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Updated.After(rows[j].Updated)
		})
		for _, row := range rows {
			payload.Queue = append(payload.Queue, QueuedReport{
				Id:      row.Id,
				Clan:    row.Clan,
				Status:  row.Status,
				Created: row.Created.Format(time.DateTime),
				Updated: row.Updated.Format(time.DateTime),
				URL:     fmt.Sprintf("/reports/queued/%s", row.Id),
			})
		}
		payload.NoQueuedReports = len(payload.Queue) == 0

		render(w, r, payload, templateFiles...)
	}
}

func deleteReportQueuedDetail(auth AuthRepo, db interface {
	DeleteQueuedReport(cid, qid string) error
}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if !auth.CurrentUser(r).IsAuthenticated() {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		clanId := auth.CurrentUser(r).ClanId()
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		queueId := r.PathValue("queueId")
		if queueId == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
	}
}

func getReportsQueuedDetail(templatesPath string, auth AuthRepo, db interface {
	ReadQueuedReport(cid, qid string) (sqlc.DBQueuedReportDetail, error)
}) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "reports_queued_detail.gohtml"),
	}

	type Payload struct {
		Page struct {
			Title string
		}
		Clan         string
		QueuedReport struct {
			Id       string
			Clan     string
			Name     string
			Status   string
			Checksum string
			Created  string
			Updated  string
			URL      string
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		if !auth.CurrentUser(r).IsAuthenticated() {
			log.Printf("%s: %s: user: not authenticated\n", r.Method, r.URL.Path)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		clanId := auth.CurrentUser(r).ClanId()
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		queueId := r.PathValue("queueId")
		if queueId == "" {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var payload Payload
		payload.Page.Title = "Queued Report"
		payload.Clan = clanId

		row, err := db.ReadQueuedReport(clanId, queueId)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				log.Printf("%s: %s: clanId %q: readQueuedReport %v\n", r.Method, r.URL.Path, clanId, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		payload.QueuedReport.Id = queueId
		payload.QueuedReport.Clan = row.Clan
		payload.QueuedReport.Name = row.Name
		payload.QueuedReport.Checksum = row.Checksum
		payload.QueuedReport.Status = row.Status
		payload.QueuedReport.Created = row.Created.Format(time.DateTime)
		payload.QueuedReport.Updated = row.Updated.Format(time.DateTime)
		payload.QueuedReport.URL = fmt.Sprintf("/reports/queued/%s", row.Id)

		render(w, r, payload, templateFiles...)
	}
}

func render(w http.ResponseWriter, r *http.Request, payload any, templates ...string) {
	// parse the template file, logging any errors
	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// execute the template with our payload, saving the response to a buffer so that we can capture errors in a nice way.
	buf := &bytes.Buffer{}
	if err = tmpl.Execute(buf, payload); err != nil {
		log.Printf("%s: %s: template: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

func renderFragment(w http.ResponseWriter, r *http.Request, name string, payload any, templates ...string) {
	// parse the template file, logging any errors
	tmpl, err := template.ParseFiles(templates...)
	if err != nil {
		log.Printf("%s: %s: fragment: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// execute the template with our payload, saving the response to a buffer so that we can capture errors in a nice way.
	buf := &bytes.Buffer{}
	if err = tmpl.ExecuteTemplate(buf, name, payload); err != nil {
		log.Printf("%s: %s: fragment: %v", r.Method, r.URL.Path, err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}

func getLandingPage(templatesPath string) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "layout.gohtml"),
	}

	type Payload struct {
		Page struct {
			Title string
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload Payload
		payload.Page.Title = "OttoMap"

		render(w, r, payload, templateFiles...)
	}
}

// returns a handler that will serve a static file if one exists, otherwise return not found.
func getPublicFiles(prefix, root string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		file := filepath.Join(root, filepath.Clean(strings.TrimPrefix(r.URL.Path, prefix)))

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
	}
}

func computeChecksum(fileData []byte) string {
	hash := sha256.Sum256(fileData)
	return fmt.Sprintf("%x", hash)
}

func postClanReports(templatesPath string, repo interface {
	CountQueuedByChecksum(cksum string) int
	CountQueuedInProgressReports(cid string) int
	QueueReport(qid, cid, name, cksum, data string) error
}, rum *ReportUploadMutex) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		clanId := r.PathValue("clanId")
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// how to render a fragment?

		// yeah, this check has a race condition. it's okay.
		queued := repo.CountQueuedInProgressReports(clanId)
		log.Printf("%s: %s: queued %8d\n", r.Method, r.URL.Path, queued)
		if queued > 3 {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		// Parse the multipart form data
		err := r.ParseMultipartForm(10 << 20) // 10MB limit
		if err != nil {
			log.Printf("%s: %s: clanId %q: parseForm: %v", r.Method, r.URL.Path, clanId, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Get the file from the request
		file, header, err := r.FormFile("turn-report")
		if err != nil {
			log.Printf("%s: %s: clanId %q: formFile: %v", r.Method, r.URL.Path, clanId, err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer file.Close()

		// read the file data into a []byte
		fileData, err := io.ReadAll(file)
		if err != nil {
			log.Printf("%s: %s: clanId %q: readall: %v", r.Method, r.URL.Path, clanId, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// quick test that it is a turn report file
		if len(fileData) < 128 || !bytes.HasPrefix(fileData, []byte("Tribe 0")) {
			log.Printf("%s: %s: clanId %q: not a turn report\n", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		// try not to upload if we know that the file has already been uploaded
		cksum := computeChecksum(fileData)
		queued = repo.CountQueuedByChecksum(cksum)
		if queued != 0 {
			log.Printf("%s: %s: cksum %8d\n", r.Method, r.URL.Path, queued)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		fum := &FileUploadMetadata{
			Id:       uuid.New().String(),
			Clan:     clanId,
			Name:     header.Filename,
			Length:   len(fileData),
			Checksum: cksum,
			Data:     string(fileData),
		}
		log.Printf("%s: %s: clanId %q: file %q: %d bytes\n", r.Method, r.URL.Path, clanId, fum.Name, fum.Length)

		if err := repo.QueueReport(fum.Id, fum.Clan, fum.Name, fum.Checksum, fum.Data); err != nil {
			log.Printf("%s: %s: clanId %q: file %q: %v\n", r.Method, r.URL.Path, clanId, fum.Name, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// old code - switching to asynchronous processing after upload
		//// acquire a semaphore slot and upload the file, reporting any errors
		//channelAvailable, started := false, time.Now()
		//select {
		//case rum.sem <- struct{}{}:
		//	log.Printf("%s: %s: clanId %q: file %q: pushing\n", r.Method, r.URL.Path, clanId, fum.Name)
		//	// send the file data to the channel without blocking
		//	select {
		//	case rum.fileDataChan <- fum:
		//		// file data sent successfully
		//		channelAvailable = true
		//	default:
		//		// channel is full, report the error
		//		channelAvailable = false
		//		log.Printf("%s: %s: clanId %q: file %q: channel is full\n", r.Method, r.URL.Path, clanId, header.Filename)
		//	}
		//	<-rum.sem // release the semaphore slot
		//	log.Printf("%s: %s: clanId %q: file %q: pushed in %v\n", r.Method, r.URL.Path, clanId, fum.Name, time.Now().Sub(started))
		//default:
		//	// no semaphore slots available, handle the error
		//	channelAvailable = false
		//	log.Printf("%s: %s: clanId %q: file %q: no sem slots\n", r.Method, r.URL.Path, clanId, header.Filename)
		//}
		//
		//if !channelAvailable {
		//	http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		//	return
		//}

		http.Redirect(w, r, fmt.Sprintf("/reports/%s", fum.Id), http.StatusFound)
	}
}

type UploadUIFragmentPayload struct {
	Clan        string
	Status      string
	Message     string
	PctComplete int
	UploadURL   string
	StatusURL   string
}

func postClanReportsUpload(templatesPath string, repo interface {
	CountQueuedByChecksum(cksum string) int
	CountQueuedInProgressReports(cid string) int
	QueueReport(qid, cid, name, cksum, data string) error
}) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "upload_ui.gohtml"),
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)
		startTime := time.Now()

		clanId := r.PathValue("clanId")
		if clanId == "" {
			log.Printf("%s: %s: clanId %q", r.Method, r.URL.Path, clanId)
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		var payload UploadUIFragmentPayload

		// yeah, this check has a race condition. it's okay.
		queued := repo.CountQueuedInProgressReports(clanId)
		log.Printf("%s: %s: queued %8d\n", r.Method, r.URL.Path, queued)
		if queued > 3 {
			payload.Status = "too-many-reports"
			payload.Message = fmt.Sprintf("There are currently %d reports queued for processing. Please try again later.", queued)
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}

		log.Printf("%s: %s: clanId %q: content type %q", r.Method, r.URL.Path, clanId, r.Header.Get("Content-Type"))
		//switch ct := r.Header.Get("Content-Type"); ct {
		//case "application/x-www-form-urlencoded": // parse the urlencoded form data
		//	log.Printf("%s: %s: clanId %q: content type %q: not acceptable\n", r.Method, r.URL.Path, clanId, ct)
		//	payload.Status = "bad-request"
		//	payload.Message = fmt.Sprintf("There was a problem parsing your request. The data was not accepted because of an error in the upload parameters.")
		//	renderFragment(w, r, "upload_ui", payload, templateFiles...)
		//	return
		//case "multipart/form-data": // parse the multipart form data
		//	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		//		log.Printf("%s: %s: clanId %q: parseMPForm: %v", r.Method, r.URL.Path, clanId, err)
		//		payload.Status = "bad-request"
		//		payload.Message = fmt.Sprintf("There was a problem parsing your request. This is unexpected.")
		//		renderFragment(w, r, "upload_ui", payload, templateFiles...)
		//		return
		//	}
		//default:
		//	log.Printf("%s: %s: clanId %q: content type %q: unknown", r.Method, r.URL.Path, clanId, ct)
		//	payload.Status = "bad-request"
		//	payload.Message = fmt.Sprintf("There was a problem parsing your request. The data was not accepted because the server doesn't understand the content type sent by your browser.")
		//	renderFragment(w, r, "upload_ui", payload, templateFiles...)
		//	return
		//}

		file, header, err := r.FormFile("turn-report")
		if err != nil {
			log.Printf("%s: %s: clanId %q: formFile: %v", r.Method, r.URL.Path, clanId, err)
			if errors.Is(err, http.ErrMissingFile) {
				payload.Status = "bad-request"
				payload.Message = fmt.Sprintf("There was a problem parsing your request. The file was not accepted because it was missing from the request. This is probably an issue with the client; it shouldn't allow the Upload button to be pressed if no file has been selected.")
				renderFragment(w, r, "upload_ui", payload, templateFiles...)
				return
			}
			payload.Status = "bad-request"
			payload.Message = fmt.Sprintf("There was a problem parsing your request.")
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}
		defer func() {
			_ = file.Close()
		}()

		// read the file data into a []byte
		fileData, err := io.ReadAll(file)
		if err != nil {
			log.Printf("%s: %s: clanId %q: readall: %v", r.Method, r.URL.Path, clanId, err)
			payload.Status = "bad-request"
			payload.Message = fmt.Sprintf("There was a problem parsing your request.")
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}

		// quick test that it is a turn report file
		if len(fileData) < 128 || !bytes.HasPrefix(fileData, []byte("Tribe 0")) {
			log.Printf("%s: %s: clanId %q: not a turn report\n", r.Method, r.URL.Path, clanId)
			payload.Status = "bad-request"
			payload.Message = fmt.Sprintf("The file was uploaded but it did not contain a valid turn report.")
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}

		// try not to upload if we know that the file has already been uploaded
		cksum := computeChecksum(fileData)
		queued = repo.CountQueuedByChecksum(cksum)
		if queued != 0 {
			log.Printf("%s: %s: cksum %8d\n", r.Method, r.URL.Path, queued)
			payload.Status = "bad-request"
			payload.Message = fmt.Sprintf("The file is a duplicate of a previous turn report.")
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}

		queueId := uuid.New().String()

		fum := &FileUploadMetadata{
			Id:       queueId,
			Clan:     clanId,
			Name:     header.Filename,
			Length:   len(fileData),
			Checksum: cksum,
			Data:     string(fileData),
		}
		log.Printf("%s: %s: clanId %q: file %q: %d bytes\n", r.Method, r.URL.Path, clanId, fum.Name, fum.Length)

		if err := repo.QueueReport(fum.Id, fum.Clan, fum.Name, fum.Checksum, fum.Data); err != nil {
			log.Printf("%s: %s: clanId %q: file %q: %v\n", r.Method, r.URL.Path, clanId, fum.Name, err)
			payload.Status = "bad-request"
			payload.Message = fmt.Sprintf("We were unable to queue your report for processing due to database issues Please report this to the site administrator.")
			renderFragment(w, r, "upload_ui", payload, templateFiles...)
			return
		}

		//// simulate a load by introducing a random delay between 3 and 5 seconds
		//delay := time.Duration(rand.IntN(2)+3) * time.Second
		//time.Sleep(delay)

		payload.Status = "queued"
		payload.Message = "Your report has been queued for processing."
		payload.PctComplete = 0
		payload.UploadURL = fmt.Sprintf("/clan/%s/reports/upload", clanId)
		payload.StatusURL = fmt.Sprintf("/reports/upload/%s", queueId)

		renderFragment(w, r, "upload_ui", payload, templateFiles...)

		log.Printf("%s: %s: finished: %v\n", r.Method, r.URL.Path, time.Now().Sub(startTime))
	}
}

func getReportsUpload(templatesPath string, repo *sqlc.DB) http.HandlerFunc {
	templateFiles := []string{
		filepath.Join(templatesPath, "upload_ui.gohtml"),
	}

	counter := 0
	return func(w http.ResponseWriter, r *http.Request) {
		queueId := r.PathValue("queueId")
		if queueId == "" {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}

		log.Printf("%s: %s: entered\n", r.Method, r.URL.Path)

		var payload UploadUIFragmentPayload

		counter = counter + 23
		if counter < 10 {
			payload.Status = "queued"
			payload.Message = "Your report has been queued for processing."
			payload.PctComplete = counter
		} else if counter < 100 {
			payload.Status = "parsing"
			payload.Message = "Your report is being parsed."
			payload.PctComplete = counter
		} else {
			counter = 0
			payload.Status = "complete"
			payload.Message = "Your report has been uploaded. Please refresh the page to see the results."
			payload.PctComplete = counter
		}
		payload.StatusURL = fmt.Sprintf("/reports/upload/%s", queueId)

		renderFragment(w, r, "upload_ui", payload, templateFiles...)
	}
}
