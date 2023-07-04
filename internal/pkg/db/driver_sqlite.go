package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"database/sql"
	"errors"
	"io/ioutil"
	"log"

	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Glue between our DB and the SQLite driver
type SQLite struct {
	db *sql.DB
}

func OpenSQLite(dbname string) DB {
	if _, err := os.Stat(dbname); err == nil {
		db, err := sql.Open("sqlite3", dbname)
		if err != nil {
			log.Fatalf("Failed to open db: %s", err.Error())
		}
		return &SQLite{db}
	} else if os.IsNotExist(err) {
		return NewSQLite(dbname)
	} else {
		log.Fatalf("Could not stat file: %s", err.Error())
	}
	return nil
}

func NewSQLite(dbname string) DB {
	db, err := sql.Open("sqlite3", dbname)
	if err != nil {
		log.Fatalf("Failed to open db: %s", err.Error())
	}

	query, err := ioutil.ReadFile("init/sqlite3.sql")
	if err != nil {
		log.Fatalf("Error running DB setup command: %s", err.Error())
	}
	if _, err := db.Exec(string(query)); err != nil {
		panic(err)
	}

	return &SQLite{db}
}

func (s *SQLite) GetAccounts() ([]model.Account, error) {
	return make([]model.Account, 0), errors.New("not implemented")
}
func (s *SQLite) NewAccount(a *model.Account, startingBal int) error {
	return errors.New("not implemented")
}
func (s *SQLite) UpdateAccount(model.Account) error { return errors.New("not implemented") }
func (s *SQLite) DeleteAccount(id uint) error       { return errors.New("not implemented") }

func (s *SQLite) SetStartingBalance(id uint, balance int) error { return errors.New("not implemented") }

func (s *SQLite) GetEnvelopeGroups() ([]model.EnvelopeGroup, error) {
	return make([]model.EnvelopeGroup, 0), errors.New("not implemented")
}
func (s *SQLite) NewEnvelopeGroup(*model.EnvelopeGroup) error   { return errors.New("not implemented") }
func (s *SQLite) UpdateEnvelopeGroup(model.EnvelopeGroup) error { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelopeGroup(id uint) error             { return errors.New("not implemented") }

func (s *SQLite) GetEnvelopes() ([]model.Envelope, error) {
	return make([]model.Envelope, 0), errors.New("not implemented")
}
func (s *SQLite) GetEnvelopesInGroup(model.EnvelopeGroup) ([]model.Envelope, error) {
	return make([]model.Envelope, 0), errors.New("not implemented")
}
func (s *SQLite) NewEnvelope(*model.Envelope) error   { return errors.New("not implemented") }
func (s *SQLite) UpdateEnvelope(model.Envelope) error { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelope(id uint) error        { return errors.New("not implemented") }

func (s *SQLite) GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAllAccountTransactions(id uint) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAllEnvelopeTransactions(id uint) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAccountTransactions(month bcdate.BCDate, id uint) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetEnvelopeTransactions(month bcdate.BCDate, id uint) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), errors.New("not implemented")
}

func (s *SQLite) GetAccountSummary(month bcdate.BCDate, id uint) (model.AccountSummary, error) {
	return model.AccountSummary{}, errors.New("not implemented")
}
func (s *SQLite) GetEnvelopeSummary(month bcdate.BCDate, id uint) (model.EnvelopeSummary, error) {
	return model.EnvelopeSummary{}, errors.New("not implemented")
}
func (s *SQLite) GetOverallSummary(month bcdate.BCDate) (model.Summary, error) {
	return model.Summary{}, errors.New("not implemented")
}

func (s *SQLite) UpdateTransaction(model.AccountTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) NewAccountTransaction(*model.AccountTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) NewEnvelopeTransaction(*model.EnvelopeTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) DeleteAccountTransaction(id uint) error  { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelopeTransaction(id uint) error { return errors.New("not implemented") }
