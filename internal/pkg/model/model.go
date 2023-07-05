package model

import (
	"budgeting/internal/pkg/bcdate"
)

// All structure definitions should go here
// This is the M of MVC

type PKEY uint

type AccountClass uint16

const (
	AT_CHECKING AccountClass = iota
	AT_SAVINGS
	AT_INVESTMENT
	AT_LOAN
	AT_CREDITCARD
)

type Account struct {
	ID        PKEY
	Hidden    bool
	Offbudget bool
	Debt      bool

	Institution string
	Name        string
	Class       AccountClass
}

type TransactionType uint16

const (
	TT_NORM TransactionType = iota
	TT_INCOME
	TT_TRANSFER
	TT_ADJUST
)

type AccountTransaction struct {
	ID         PKEY
	AccountID  PKEY
	EnvelopeID PKEY

	Typ TransactionType

	PostDate bcdate.BCDate

	Amount  int
	Cleared bool
	Memo    string
}

type AccountSummary struct {
	AccountID PKEY
	Month     bcdate.BCDate

	Bal       int
	In        int
	Out       int
	Uncleared int
}

type EnvelopeGroup struct {
	ID   PKEY
	Name string
	Sort int
}

type GoalType uint16

const (
	GT_NONE = iota
	GT_RECUR
	GT_TGT
	GT_RECTIL
)

type Envelope struct {
	ID      PKEY
	GroupID PKEY

	DebtAccount PKEY
	Hidden      bool

	Name  string
	Notes string

	Goal    GoalType
	GoalAmt int
	GoalTgt int

	Sort int
}

type EnvelopeTransaction struct {
	ID         PKEY
	EnvelopeID PKEY
	PostDate   bcdate.BCDate
	Amount     int
}

type EnvelopeSummary struct {
	EnvelopeID PKEY
	Month      bcdate.BCDate
	Bal        int
	In         int
	Out        int
}

type Summary struct {
	Month    bcdate.BCDate
	Float    int
	Income   int
	Expenses int
	Banked   int
	NetWorth int
	Delta    int
}
