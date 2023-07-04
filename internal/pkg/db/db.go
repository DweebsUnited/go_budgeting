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
	DeleteAccount(id model.PKEY) error

	SetStartingBalance(id model.PKEY, balance int) error

	GetEnvelopeGroups() ([]model.EnvelopeGroup, error)
	NewEnvelopeGroup(*model.EnvelopeGroup) error
	UpdateEnvelopeGroup(model.EnvelopeGroup) error
	DeleteEnvelopeGroup(id model.PKEY) error

	GetEnvelopes() ([]model.Envelope, error)
	GetEnvelopesInGroup(id model.PKEY) ([]model.Envelope, error)
	NewEnvelope(*model.Envelope) error
	UpdateEnvelope(model.Envelope) error
	DeleteEnvelope(id model.PKEY) error

	GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error)
	GetAllAccountTransactions(id model.PKEY) ([]model.AccountTransaction, error)
	GetAllEnvelopeTransactions(id model.PKEY) ([]model.EnvelopeTransaction, error)
	GetAccountTransactions(month bcdate.BCDate, id model.PKEY) ([]model.AccountTransaction, error)
	GetEnvelopeTransactions(month bcdate.BCDate, id model.PKEY) ([]model.EnvelopeTransaction, error)

	GetAccountSummary(month bcdate.BCDate, id model.PKEY) (model.AccountSummary, error)
	GetEnvelopeSummary(month bcdate.BCDate, id model.PKEY) (model.EnvelopeSummary, error)
	GetOverallSummary(month bcdate.BCDate) (model.Summary, error)

	UpdateTransaction(model.AccountTransaction) error
	NewAccountTransaction(*model.AccountTransaction) error
	NewEnvelopeTransaction(*model.EnvelopeTransaction) error
	DeleteAccountTransaction(id model.PKEY) error
	DeleteEnvelopeTransaction(id model.PKEY) error
}
