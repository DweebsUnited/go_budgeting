package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
)

// TODO: Interface wrapping various DB drivers with our Models
type DB interface {
	GetAccounts() ([]model.Account, error)
	NewAccount(a *model.Account, startingBal int) error
	UpdateAccount(model.Account) error
	DeleteAccount(id uint) error

	SetStartingBalance(id uint, balance int) error

	GetEnvelopeGroups() ([]model.EnvelopeGroup, error)
	NewEnvelopeGroup(*model.EnvelopeGroup) error
	UpdateEnvelopeGroup(model.EnvelopeGroup) error
	DeleteEnvelopeGroup(id uint) error

	GetEnvelopes() ([]model.Envelope, error)
	GetEnvelopesInGroup(model.EnvelopeGroup) ([]model.Envelope, error)
	NewEnvelope(*model.Envelope) error
	UpdateEnvelope(model.Envelope) error
	DeleteEnvelope(id uint) error

	GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error)
	GetAllAccountTransactions(id uint) ([]model.AccountTransaction, error)
	GetAllEnvelopeTransactions(id uint) ([]model.EnvelopeTransaction, error)
	GetAccountTransactions(month bcdate.BCDate, id uint) ([]model.AccountTransaction, error)
	GetEnvelopeTransactions(month bcdate.BCDate, id uint) ([]model.EnvelopeTransaction, error)

	GetAccountSummary(month bcdate.BCDate, id uint) (model.AccountSummary, error)
	GetEnvelopeSummary(month bcdate.BCDate, id uint) (model.EnvelopeSummary, error)
	GetOverallSummary(month bcdate.BCDate) (model.Summary, error)

	UpdateTransaction(model.AccountTransaction) error
	NewAccountTransaction(*model.AccountTransaction) error
	NewEnvelopeTransaction(*model.EnvelopeTransaction) error
	DeleteAccountTransaction(id uint) error
	DeleteEnvelopeTransaction(id uint) error
}
