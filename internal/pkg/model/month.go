package model

import "budgeting/internal/pkg/bcdate"

// This is an aggregation model, a way to get a whole month at a time
// This is the primary application model

type Month_Account struct {
	A Account
	S AccountSummary
	T []AccountTransaction
}

type Month_Groups struct {
	G EnvelopeGroup
	E []Month_Envelopes
}

type Month_Envelopes struct {
	E Envelope
	S EnvelopeSummary
	T []EnvelopeTransaction
}

type Month struct {
	Month    bcdate.BCDate
	Accounts []Month_Account
	Groups   []Month_Groups
}
