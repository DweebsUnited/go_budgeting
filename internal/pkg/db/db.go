package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
)

// Interface wrapping various DB drivers with our Models

type DB interface {
	Open(string) error
	Init() error
	Run(fname string) error

	GetAccounts() ([]model.Account, error)
	GetAccount(id model.PKEY) (model.Account, error)
	NewAccount(a *model.Account) error
	UpdateAccount(model.Account) error
	DeleteAccount(id model.PKEY) error

	GetStartingBalance(id model.PKEY) (int, error)
	SetStartingBalance(id model.PKEY, balance int) error

	GetEnvelopeGroups() ([]model.EnvelopeGroup, error)
	GetEnvelopeGroup(id model.PKEY) (model.EnvelopeGroup, error)
	NewEnvelopeGroup(*model.EnvelopeGroup) error
	UpdateEnvelopeGroup(model.EnvelopeGroup) error
	DeleteEnvelopeGroup(id model.PKEY) error

	GetEnvelopesInGroup(id model.PKEY) ([]model.Envelope, error)
	GetDebtEnvelopeFor(id model.PKEY) (model.Envelope, error)

	GetEnvelopes() ([]model.Envelope, error)
	GetEnvelope(id model.PKEY) (model.Envelope, error)
	NewEnvelope(*model.Envelope) error
	UpdateEnvelope(model.Envelope) error
	DeleteEnvelope(id model.PKEY) error

	GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error)

	GetAllAccountTransactions(id model.PKEY) ([]model.AccountTransaction, error)
	GetAccountTransactions(month bcdate.BCDate, id model.PKEY) ([]model.AccountTransaction, error)
	NewAccountTransaction(*model.AccountTransaction) error
	UpdateAccountTransaction(model.AccountTransaction) error
	DeleteAccountTransaction(id model.PKEY) error

	GetAllEnvelopeTransactions(id model.PKEY) ([]model.EnvelopeTransaction, error)
	GetEnvelopeTransactions(month bcdate.BCDate, id model.PKEY) ([]model.EnvelopeTransaction, error)
	NewEnvelopeTransaction(*model.EnvelopeTransaction) error
	UpdateEnvelopeTransaction(model.EnvelopeTransaction) error
	DeleteEnvelopeTransaction(id model.PKEY) error

	GetAccountSummary(month bcdate.BCDate, id model.PKEY) (model.AccountSummary, error)
	GetEnvelopeSummary(month bcdate.BCDate, id model.PKEY) (model.EnvelopeSummary, error)
	GetOverallSummary(month bcdate.BCDate) (model.Summary, error)
}
