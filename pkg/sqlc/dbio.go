// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package sqlc

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/mdhender/ottomap/cerrs"
	reports "github.com/mdhender/ottomap/pkg/reports/domain"
	turns "github.com/mdhender/ottomap/pkg/turns/domain"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
	"path/filepath"
	"strings"
	"time"

	"log"
	"os"
)

var (
	//go:embed schema.sql
	schemaDDL string
)

// CreateDatabase creates and initializes a new database.
// It is an error if the database already exists.
func CreateDatabase(dbName string) error {
	// verify that database does not exist.
	if _, err := os.Stat(dbName); !os.IsNotExist(err) {
		return cerrs.ErrDatabaseExists
	}
	// create the database.
	mdb, err := sql.Open("sqlite", dbName)
	if err != nil {
		return err
	}
	defer func() {
		if mdb != nil {
			_ = mdb.Close()
		}
	}()

	// confirm that the database has foreign keys enabled
	var rslt sql.Result
	checkPragma := "PRAGMA" + " foreign_keys = ON"
	if rslt, err = mdb.Exec(checkPragma); err != nil {
		log.Printf("sqlc: error: foreign keys are disabled\n")
		return cerrs.ErrForeignKeysDisabled
	} else if rslt == nil {
		log.Printf("sqlc: error: foreign keys pragma failed\n")
		return cerrs.ErrPragmaReturnedNil
	}

	// create the schema. this also runs any data initialization in the schema file.
	if _, err = mdb.Exec(schemaDDL); err != nil {
		log.Printf("sqlc: failed to initialize schema\n")
		return errors.Join(cerrs.ErrCreateSchema, err)
	}

	return nil
}

type DB struct {
	DB      *sql.DB
	Ctx     context.Context
	Queries *Queries
}

func OpenDatabase(dbName string, ctx context.Context) (*DB, error) {
	// verify that database exists.
	if _, err := os.Stat(dbName); err != nil {
		if os.IsNotExist(err) {
			return nil, cerrs.ErrDatabaseExists
		}
		return nil, err
	}

	db := &DB{Ctx: ctx}

	if mdb, err := sql.Open("sqlite", dbName); err != nil {
		return nil, err
	} else if db.DB = mdb; db.DB == nil {
		return nil, fmt.Errorf("db.db is nil")
	} else if db.Queries = New(mdb); db.Queries == nil {
		return nil, fmt.Errorf("db.q is nil")
	} else {
		// confirm that the database has foreign keys enabled
		var rslt sql.Result
		checkPragma := "PRAGMA" + " foreign_keys = ON"
		if rslt, err = mdb.Exec(checkPragma); err != nil {
			log.Printf("sqlc: error: foreign keys are disabled\n")
			return nil, cerrs.ErrForeignKeysDisabled
		} else if rslt == nil {
			log.Printf("sqlc: error: foreign keys pragma failed\n")
			return nil, cerrs.ErrPragmaReturnedNil
		}
	}

	return db, nil
}

func (db *DB) CloseDatabase() {
	if db.DB != nil {
		db.DB.Close()
		db.DB = nil
	}
}

func (db *DB) AuthenticateUserEmail(email, plaintextSecret string) (uid string, err error) {
	handle, err := db.Queries.ReadUserByEmail(db.Ctx, email)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read user authentication data"), err)
		}
		return "", nil
	}
	return db.AuthenticateUserHandle(handle, plaintextSecret)
}

func (db *DB) AuthenticateUserHandle(handle, plaintextSecret string) (uid string, err error) {
	user, err := db.Queries.ReadUserAuthData(db.Ctx, handle)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read user authentication data"), err)
		}
		return "", nil
	}

	// check if two passwords match using bcrypt's CompareHashAndPassword
	// which return nil on success and an error on failure. (from gregorygaines.com)
	//hashedSecretBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
	if err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(plaintextSecret)); err != nil {
		return "", nil
	}

	return user.Uid, nil
}

func (db *DB) AllClanReportMetadata(cid string) ([]reports.Metadata, error) {
	log.Printf("sqlc: AllClanReportMetadata(%s)\n", cid)
	rpts, err := db.Queries.ReadAllClanReports(db.Ctx, cid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(fmt.Errorf("read all clan reports"), err)
		}
	}

	var list []reports.Metadata
	for _, rpt := range rpts {
		list = append(list, reports.Metadata{
			Id:      rpt.Rid,
			TurnId:  rpt.Tid,
			Clan:    rpt.Cid,
			Status:  "N/A",
			Created: rpt.Crdttm,
		})
	}

	return list, nil
}

func (db *DB) AllClanReports(cid string) ([]reports.Report, error) {
	log.Printf("sqlc: AllClanReports(%s)\n", cid)
	rpts, err := db.Queries.ReadAllClanReports(db.Ctx, cid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(fmt.Errorf("read all clan reports"), err)
		}
	}

	var list []reports.Report
	for _, rpt := range rpts {
		list = append(list, reports.Report{
			Id:      rpt.Rid,
			Turn:    rpt.Tid,
			Clan:    rpt.Cid,
			Status:  "N/A",
			Created: rpt.Crdttm,
		})
	}

	return list, nil
}

func (db *DB) AllTurnMetadata() ([]turns.Metadata, error) {
	log.Printf("sqlc: AllTurnMetadata()\n")
	rows, err := db.Queries.ReadAllTurns(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(fmt.Errorf("read all turns"), err)
		}
	}

	var list []turns.Metadata
	for _, row := range rows {
		list = append(list, turns.Metadata{
			Id:      row.Tid,
			Name:    row.Turn,
			Year:    int(row.Year),
			Month:   int(row.Month),
			Created: row.Crdttm,
		})
	}

	return list, nil
}

func (db *DB) AllTurns() ([]turns.Turn, error) {
	log.Printf("sqlc: AllTurns()\n")
	rows, err := db.Queries.ReadAllTurns(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(fmt.Errorf("read all turns"), err)
		}
		return nil, sql.ErrNoRows
	}
	var list []turns.Turn
	for _, row := range rows {
		list = append(list, turns.Turn{
			Id:      row.Tid,
			Turn:    row.Turn,
			Year:    int(row.Year),
			Month:   int(row.Month),
			Created: row.Crdttm,
		})
	}
	return list, nil
}

func (db *DB) CountQueuedByChecksum(cksum string) int {
	n, err := db.Queries.CountQueuedByChecksum(db.Ctx, cksum)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("sqlc: countQueuedByChecksum: error: %v\n", err)
			return 999_999_999
		}
		n = 0
	}
	return int(n)
}

func (db *DB) CountQueuedInProgressReports(cid string) int {
	n, err := db.Queries.CountQueuedInProgressReports(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			log.Printf("sqlc: countQueuedInProgressReports: error: %v\n", err)
			return 999_999_999
		}
		n = 0
	}
	return int(n)
}

func (db *DB) CreateClan(uid, cid string) error {
	ccp := CreateClanParams{
		Uid: uid,
		Cid: cid,
	}
	if err := db.Queries.CreateClan(db.Ctx, ccp); err != nil {
		return errors.Join(fmt.Errorf("create clan"), err)
	}
	return nil

}

func (db *DB) CreateRole(rlid string) error {
	if len(rlid) == 0 {
		return fmt.Errorf("invalid role")
	} else if rlid != strings.TrimSpace(rlid) {
		return fmt.Errorf("invalid role")
	} else if rlid != strings.ToLower(rlid) {
		return fmt.Errorf("invalid role")
	}
	err := db.Queries.CreateRole(db.Ctx, rlid)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) CreateSession(uid string) (sid string, err error) {
	sid = uuid.New().String()
	exp := time.Now().UTC().Add(2 * 7 * 24 * time.Hour)
	if err := db.Queries.CreateSession(db.Ctx, CreateSessionParams{
		Uid:       uid,
		Sid:       sid,
		ExpiresAt: exp,
	}); err != nil {
		return "", errors.Join(fmt.Errorf("create session"), err)
	}
	return sid, nil
}

func (db *DB) CreateUser(handle, email, secret string) (string, error) {
	if len(handle) == 0 {
		return "", fmt.Errorf("invalid handle")
	} else if handle != strings.TrimSpace(handle) {
		return "", fmt.Errorf("invalid handle")
	} else if handle != strings.ToLower(handle) {
		return "", fmt.Errorf("invalid handle")
	}
	if len(email) == 0 {
		return "", fmt.Errorf("email is too short")
	} else if email != strings.TrimSpace(email) {
		return "", fmt.Errorf("email contains spaces")
	} else if email != strings.ToLower(email) {
		return "", fmt.Errorf("email is not lowercase")
	}
	if len(secret) < 7 {
		return "", fmt.Errorf("secret is too short")
	} else if email != strings.TrimSpace(email) {
		return "", fmt.Errorf("secret contains spaces")
	}

	// use bcrypt to hash the secret (from gregorygaines.com)
	hashedSecretBytes, err := bcrypt.GenerateFromPassword([]byte(secret), bcrypt.MinCost)
	if err != nil {
		return "", errors.Join(fmt.Errorf("create user"), err)
	}
	hashedSecret := string(hashedSecretBytes)
	log.Printf("createUser:\n\tsecret %q\n\thashed %q\n", secret, hashedSecret)

	uid := uuid.New().String()
	cup := CreateUserParams{
		Uid:            uid,
		Username:       handle,
		Email:          email,
		HashedPassword: hashedSecret,
	}
	if err := db.Queries.CreateUser(db.Ctx, cup); err != nil {
		return "", errors.Join(fmt.Errorf("create user"), err)
	}

	return uid, nil
}

func (db *DB) CreateUserRole(uid, rlid string) error {
	curp := CreateUserRoleParams{
		Uid:   uid,
		Rlid:  rlid,
		Value: "true",
	}
	if err := db.Queries.CreateUserRole(db.Ctx, curp); err != nil {
		return errors.Join(fmt.Errorf("create user role"), err)
	}
	return nil
}

func (db *DB) DeleteSession(sid string) error {
	return db.Queries.DeleteSession(db.Ctx, sid)
}

func (db *DB) QueueReport(qid, cid, name, cksum, data string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback()
	}()
	qtx := db.Queries.WithTx(tx)

	cqrp := CreateQueuedReportParams{
		Qid:    qid,
		Cid:    cid,
		Status: "uploading",
	}
	if err := qtx.CreateQueuedReport(db.Ctx, cqrp); err != nil {
		log.Printf("sqlc: queueReport: report: %v\n", err)
		return err
	}

	cqrdp := CreateQueuedReportDataParams{
		Qid:   qid,
		Name:  name,
		Cksum: cksum,
		Lines: data,
	}
	if err := qtx.CreateQueuedReportData(db.Ctx, cqrdp); err != nil {
		log.Printf("sqlc: queueReport: data: %v\n", err)
		return err
	}

	if err := tx.Commit(); err != nil {
		log.Printf("sqlc: queueReport: commit: %v\n", err)
		return err
	}

	return nil
}

func (db *DB) ReadInputOutputPaths() (input, output string, err error) {
	paths, err := db.Queries.ReadMetadataInputOutputPaths(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.Join(fmt.Errorf("read metadata public"), err)
		}
		return "", "", sql.ErrNoRows
	}
	return paths.InputPath, paths.OutputPath, nil
}

type DBImportedFile struct {
	Name     string
	Checksum string
	Status   string
	Created  time.Time
}

func (db *DB) ReadImportedFiles() ([]DBImportedFile, error) {
	rows, err := db.Queries.ReadAllInputMetadata(db.Ctx)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(fmt.Errorf("read imported files"), err)
	}
	var list []DBImportedFile
	for _, row := range rows {
		list = append(list, DBImportedFile{
			Name:     row.Name,
			Checksum: row.Cksum,
			Status:   row.Status,
			Created:  row.Crdttm,
		})
	}
	return list, nil
}

func (db *DB) ReadMetadataPublic() (path string, err error) {
	path, err = db.Queries.ReadMetadataPublic(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read metadata public"), err)
		}
		return "", sql.ErrNoRows
	}
	return path, nil
}

func (db *DB) ReadMetadataTemplates() (path string, err error) {
	path, err = db.Queries.ReadMetadataTemplates(db.Ctx)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read metadata templates"), err)
		}
		return "", sql.ErrNoRows
	}
	return path, nil
}

type DBPendingInputDetail struct {
	Id   int
	Path string
	Name string
}

func (db *DB) ReadPendingInput() (DBPendingInputDetail, bool) {
	rows, err := db.Queries.ReadPendingInputMetadata(db.Ctx)
	if err != nil || len(rows) == 0 {
		return DBPendingInputDetail{}, false
	}
	return DBPendingInputDetail{
		Id:   int(rows[0].ID),
		Path: rows[0].Path,
		Name: rows[0].Name,
	}, true
}

type DBQueuedReportDetail struct {
	Id       string
	Clan     string
	Name     string
	Status   string
	Checksum string
	Created  time.Time
	Updated  time.Time
}

func (db *DB) ReadQueuedReport(cid, qid string) (DBQueuedReportDetail, error) {
	rqrp := ReadQueuedReportParams{
		Cid: cid,
		Qid: qid,
	}
	row, err := db.Queries.ReadQueuedReport(db.Ctx, rqrp)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return DBQueuedReportDetail{}, errors.Join(fmt.Errorf("read queued report"), err)
		}
		return DBQueuedReportDetail{}, sql.ErrNoRows
	}

	rpt := DBQueuedReportDetail{
		Id:      qid,
		Clan:    row.Cid,
		Status:  row.Status,
		Created: row.Crdttm,
		Updated: row.Updttm,
	}
	if row.Cksum.Valid {
		rpt.Checksum = row.Cksum.String
	}
	if row.Name.Valid {
		rpt.Name = row.Name.String
	}

	return rpt, nil
}

type DBQueuedReport struct {
	Id      string
	Clan    string
	Status  string
	Created time.Time
	Updated time.Time
}

func (db *DB) ReadQueuedReports(cid string) ([]DBQueuedReport, error) {
	rows, err := db.Queries.ReadQueuedReports(db.Ctx, cid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(fmt.Errorf("read queued reports"), err)
		}
	}

	var list []DBQueuedReport
	for _, row := range rows {
		list = append(list, DBQueuedReport{
			Id:      row.Qid,
			Clan:    row.Cid,
			Status:  row.Status,
			Created: row.Crdttm,
			Updated: row.Updttm,
		})
	}

	return list, nil
}

func (db *DB) ReadSession(sid string) (uid string, exp time.Time, err error) {
	session, err := db.Queries.ReadSession(db.Ctx, sid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", time.Time{}, errors.Join(fmt.Errorf("read session"), err)
		}
		return "", time.Time{}, sql.ErrNoRows
	}
	return session.Uid, session.ExpiresAt, nil
}

func (db *DB) ReadUser(uid string) (handle, email string, err error) {
	user, err := db.Queries.ReadUser(db.Ctx, uid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.Join(fmt.Errorf("read user"), err)
		}
		return "", "", sql.ErrNoRows
	}
	return user.Username, user.Email, nil
}

func (db *DB) ReadUserAuthData(handle string) (uid, hashedSecret string, err error) {
	user, err := db.Queries.ReadUserAuthData(db.Ctx, handle)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", "", errors.Join(fmt.Errorf("read user auth data"), err)
		}
		return "", "", sql.ErrNoRows
	}
	return user.Uid, user.HashedPassword, nil
}

func (db *DB) ReadUserClan(uid string) (clan string, err error) {
	cid, err := db.Queries.ReadUserClan(db.Ctx, uid)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read user clan"), err)
		}
		return "", sql.ErrNoRows
	}
	return cid, nil
}

func (db *DB) ReadUserRole(uid, role string) (value string, err error) {
	value, err = db.Queries.ReadUserRole(db.Ctx, ReadUserRoleParams{
		Uid:  uid,
		Rlid: role,
	})
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return "", errors.Join(fmt.Errorf("read user role"), err)
		}
		return "", sql.ErrNoRows
	}
	return value, nil
}

func (db *DB) UpdateInputOutputPaths(input, output string) error {
	if path, err := filepath.Abs(input); err != nil {
		return err
	} else if path != input {
		return errors.New("input path is not absolute")
	}
	if path, err := filepath.Abs(output); err != nil {
		return err
	} else if path != output {
		return errors.New("output path is not absolute")
	}

	umiop := UpdateMetadataInputOutputPathsParams{
		InputPath:  input,
		OutputPath: output,
	}
	return db.Queries.UpdateMetadataInputOutputPaths(db.Ctx, umiop)
}

func (db *DB) UpdateInputStatus(id int, from, to string) error {
	return db.Queries.UpdateInputStatus(db.Ctx, UpdateInputStatusParams{
		ID:       int64(id),
		Status:   from,
		Status_2: to,
	})
}

func (db *DB) UpdateMetadataPublicPath(public string) error {
	if path, err := filepath.Abs(public); err != nil {
		return err
	} else if path != public {
		return errors.New("public path is not absolute")
	}
	return db.Queries.UpdateMetadataPublic(db.Ctx, public)
}

func (db *DB) UpdateMetadataTemplatesPath(templates string) error {
	if path, err := filepath.Abs(templates); err != nil {
		return err
	} else if path != templates {
		return errors.New("public path is not absolute")
	}
	return db.Queries.UpdateMetadataTemplates(db.Ctx, templates)
}

func (db *DB) OpenSession() (*DBSession, error) {
	tx, err := db.DB.BeginTx(db.Ctx, nil)
	if err != nil {
		return nil, err
	}
	return &DBSession{
		tx:      tx,
		Queries: db.Queries.WithTx(tx),
	}, nil
}

type DBSession struct {
	tx      *sql.Tx
	Queries *Queries
}

func (dbs *DBSession) Close(err error) error {
	if err != nil {
		_ = dbs.Rollback()
		dbs.tx = nil
		return err
	}
	_ = dbs.Commit()
	dbs.tx = nil
	return err
}

func (dbs *DBSession) Abort(err error) error {
	_ = dbs.Rollback()
	dbs.tx = nil
	return err
}

func (dbs *DBSession) Commit() error {
	if dbs.tx == nil {
		return nil
	}
	return dbs.tx.Commit()
}

func (dbs *DBSession) Rollback() error {
	if dbs.tx == nil {
		return nil
	}
	return dbs.tx.Rollback()
}
