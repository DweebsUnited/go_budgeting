package model

// TODO: All structure definitions should go here
// This is the M of MVC

type Summary struct {
	Float    int
	Income   int
	Expenses int
	Gain     int
	Bank     int
}

type Account struct {
	id   int
	name string
	typ  string

	balance       int
	bal_uncleared int
	bal_cleared   int

	amt_in  int
	amt_out int
	amt_net int
}

type envelopeGroup struct {
	id   int
	name string
}

type envelope struct {
	id   int
	name string
}
