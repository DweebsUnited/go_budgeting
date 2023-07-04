package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"database/sql"
	"errors"
	"fmt"
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
	accts := make([]model.Account, 0)

	rows, err := s.db.Query("SELECT * FROM a")
	if err != nil {
		return nil, fmt.Errorf("GetAccounts -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var acct model.Account
		if err := rows.Scan(
			&acct.ID,
			&acct.Hidden,
			&acct.Offbudget,
			&acct.Debt,
			&acct.Institution,
			&acct.Name,
			&acct.Class,
		); err != nil {
			return nil, fmt.Errorf("GetAccounts -- %w", err)
		}
		accts = append(accts, acct)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAccounts -- %w", err)
	}
	return accts, nil
}
func (s *SQLite) NewAccount(a *model.Account, startingBal int) error {
	return errors.New("not implemented")
}
func (s *SQLite) UpdateAccount(model.Account) error { return errors.New("not implemented") }
func (s *SQLite) DeleteAccount(id model.PKEY) error { return errors.New("not implemented") }

func (s *SQLite) SetStartingBalance(id model.PKEY, balance int) error {
	return errors.New("not implemented")
}

func (s *SQLite) GetEnvelopeGroups() ([]model.EnvelopeGroup, error) {
	egs := make([]model.EnvelopeGroup, 0)

	rows, err := s.db.Query("SELECT * FROM e_grp")
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var eg model.EnvelopeGroup
		if err := rows.Scan(
			&eg.ID,
			&eg.Name,
			&eg.Sort,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
		}
		egs = append(egs, eg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
	}
	return egs, nil
}
func (s *SQLite) NewEnvelopeGroup(*model.EnvelopeGroup) error   { return errors.New("not implemented") }
func (s *SQLite) UpdateEnvelopeGroup(model.EnvelopeGroup) error { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelopeGroup(id model.PKEY) error       { return errors.New("not implemented") }

func (s *SQLite) GetEnvelopes() ([]model.Envelope, error) {
	return make([]model.Envelope, 0), errors.New("not implemented")
}
func (s *SQLite) GetEnvelopesInGroup(id model.PKEY) ([]model.Envelope, error) {
	es := make([]model.Envelope, 0)

	rows, err := s.db.Query("SELECT * FROM e WHERE groupID = ?", id)
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopesInGroup -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var e model.Envelope
		if err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.Hidden,
			&e.DebtAccount,
			&e.Name,
			&e.Notes,
			&e.Goal,
			&e.GoalAmt,
			&e.GoalTgt,
			&e.Sort,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopesInGroup -- %w", err)
		}
		es = append(es, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopesInGroup -- %w", err)
	}
	return es, nil
}
func (s *SQLite) NewEnvelope(*model.Envelope) error   { return errors.New("not implemented") }
func (s *SQLite) UpdateEnvelope(model.Envelope) error { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelope(id model.PKEY) error  { return errors.New("not implemented") }

func (s *SQLite) GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAllAccountTransactions(id model.PKEY) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAllEnvelopeTransactions(id model.PKEY) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetAccountTransactions(month bcdate.BCDate, id model.PKEY) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), errors.New("not implemented")
}
func (s *SQLite) GetEnvelopeTransactions(month bcdate.BCDate, id model.PKEY) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), errors.New("not implemented")
}

func (s *SQLite) GetAccountSummary(month bcdate.BCDate, id model.PKEY) (model.AccountSummary, error) {
	summ := model.AccountSummary{}
	row := s.db.QueryRow("SELECT * FROM a_chk WHERE month <= ? AND accountID = ? ORDER BY month DESC LIMIT 1", month, id)
	if err := row.Scan(
		&summ.AccountID,
		&summ.Month,
		&summ.Bal,
		&summ.In,
		&summ.Out,
		&summ.Cleared,
	); err != nil {
		return summ, fmt.Errorf("GetAccountSummary -- %w", err)
	}

	return summ, nil
}
func (s *SQLite) GetEnvelopeSummary(month bcdate.BCDate, id model.PKEY) (model.EnvelopeSummary, error) {
	summ := model.EnvelopeSummary{}
	row := s.db.QueryRow("SELECT * FROM e_chk WHERE month <= ? AND envelopeID = ? ORDER BY month DESC LIMIT 1", month, id)
	if err := row.Scan(
		&summ.EnvelopeID,
		&summ.Month,
		&summ.Bal,
		&summ.In,
		&summ.Out,
	); err != nil {
		return summ, fmt.Errorf("GetEnvelopeSummary -- %w", err)
	}

	return summ, nil
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
func (s *SQLite) DeleteAccountTransaction(id model.PKEY) error  { return errors.New("not implemented") }
func (s *SQLite) DeleteEnvelopeTransaction(id model.PKEY) error { return errors.New("not implemented") }
