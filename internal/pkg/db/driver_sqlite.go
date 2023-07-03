package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"database/sql"
)

// Glue between our DB and the SQLite driver
type SQLite struct {
	db *sql.DB
}

func NewSQLiteDB() DB {
	return nil
}

func ConnectToSQLiteDB() DB {
	return nil
}

func (s *SQLite) GetAccounts() ([]model.Account, error) {
	return make([]model.Account, 0), nil
}
func (s *SQLite) NewAccount(a *model.Account, startingBal int) error { return nil }
func (s *SQLite) UpdateAccount(model.Account) error                  { return nil }
func (s *SQLite) DeleteAccount(id uint) error                        { return nil }

func (s *SQLite) SetStartingBalance(id uint, balance int) error { return nil }

func (s *SQLite) GetEnvelopeGroups() ([]model.EnvelopeGroup, error) {
	return make([]model.EnvelopeGroup, 0), nil
}
func (s *SQLite) NewEnvelopeGroup(*model.EnvelopeGroup) error   { return nil }
func (s *SQLite) UpdateEnvelopeGroup(model.EnvelopeGroup) error { return nil }
func (s *SQLite) DeleteEnvelopeGroup(id uint) error             { return nil }

func (s *SQLite) GetEnvelopes() ([]model.Envelope, error) {
	return make([]model.Envelope, 0), nil
}
func (s *SQLite) GetEnvelopesInGroup(model.EnvelopeGroup) ([]model.Envelope, error) {
	return make([]model.Envelope, 0), nil
}
func (s *SQLite) NewEnvelope(*model.Envelope) error   { return nil }
func (s *SQLite) UpdateEnvelope(model.Envelope) error { return nil }
func (s *SQLite) DeleteEnvelope(id uint) error        { return nil }

func (s *SQLite) GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), nil
}
func (s *SQLite) GetAllAccountTransactions(id uint) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), nil
}
func (s *SQLite) GetAllEnvelopeTransactions(id uint) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), nil
}
func (s *SQLite) GetAccountTransactions(month bcdate.BCDate, id uint) ([]model.AccountTransaction, error) {
	return make([]model.AccountTransaction, 0), nil
}
func (s *SQLite) GetEnvelopeTransactions(month bcdate.BCDate, id uint) ([]model.EnvelopeTransaction, error) {
	return make([]model.EnvelopeTransaction, 0), nil
}

func (s *SQLite) GetAccountSummary(month bcdate.BCDate, id uint) (model.AccountSummary, error) {
	return model.AccountSummary{}, nil
}
func (s *SQLite) GetEnvelopeSummary(month bcdate.BCDate, id uint) (model.EnvelopeSummary, error) {
	return model.EnvelopeSummary{}, nil
}
func (s *SQLite) GetOverallSummary(month bcdate.BCDate) (model.Summary, error) {
	return model.Summary{}, nil
}

func (s *SQLite) UpdateTransaction(model.AccountTransaction) error        { return nil }
func (s *SQLite) NewAccountTransaction(*model.AccountTransaction) error   { return nil }
func (s *SQLite) NewEnvelopeTransaction(*model.EnvelopeTransaction) error { return nil }
func (s *SQLite) DeleteAccountTransaction(id uint) error                  { return nil }
func (s *SQLite) DeleteEnvelopeTransaction(id uint) error                 { return nil }
