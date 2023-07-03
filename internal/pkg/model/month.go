package model

import "budgeting/internal/pkg/bcdate"

// This is an aggregation model, a way to get a whole month at a time

type Month_Account struct {
	A Account
	S AccountSummary
	T []AccountTransaction
}

type Month_Envelopes struct {
	E Envelope
	S EnvelopeSummary
	T []EnvelopeTransaction
}

type Month struct {
	Month     bcdate.BCDate
	Accounts  []Month_Account
	Groups    []EnvelopeGroup
	Envelopes []Month_Envelopes
}
