package app

import (
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/shiftpath"
	"net/http"
)

// TODO: Handler for API endpoints
// Call Controllers, then dump the result to JSON
type APIHandler struct {
	sdb db.DB
}

func NewAPIHandler(sdb db.DB) http.Handler {
	return &APIHandler{sdb}
}

func (h *APIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	head, tail := shiftpath.ShiftPath(r.URL.Path)

	switch head {
	case "summary":
		h.ServeHTTP_summary(w, r)
	case "accounts":
		h.ServeHTTP_accounts(w, r)
	case "account":
		h.ServeHTTP_account(w, r, tail)
	case "transactions":
		h.ServeHTTP_transactions(w, r)
	case "transaction":
		h.ServeHTTP_transaction(w, r, tail)
	case "envelopes":
		h.ServeHTTP_envelopes(w, r)
	case "envelope":
		h.ServeHTTP_envelope(w, r, tail)
	case "sanity":
		h.ServeHTTP_sanity(w, r)

	// Anything else, 404
	default:
		http.NotFound(w, r)
	}

}

func (h *APIHandler) ServeHTTP_summary(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api/summary"))

	// TODO: Return summary info
}

func (h *APIHandler) ServeHTTP_accounts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api/accounts"))

	// TODO: Return account list
}

func (h *APIHandler) ServeHTTP_account(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/api/account"))

	// TODO: GET takes account id, returns transactions
	// TODO: POST creates a new account
	// TODO: DELETE deletes an account and all associated data (eek!)
	// TODO: PATCH updates an accounts info -- does not allow type changes
}

func (h *APIHandler) ServeHTTP_transactions(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api/transactions"))

	// TODO: Return full transaction list
}

func (h *APIHandler) ServeHTTP_transaction(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/api/transaction"))

	// TODO: POST creates a new transaction
	// TODO: DELETE deletes a transaction
	// TODO: PATCH updates a transaction by removing old, and creating new
}

func (h *APIHandler) ServeHTTP_envelopes(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api/envelopes"))

	// TODO: Return envelope list
}

func (h *APIHandler) ServeHTTP_envelope(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/api/envelope"))

	// TODO: Return envelope transaction list
}

func (h *APIHandler) ServeHTTP_sanity(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/api/sanity"))

	// TODO: Run sanity checks on the database to ensure all temp values are correct
}
