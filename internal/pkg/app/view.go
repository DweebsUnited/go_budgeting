package app

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/middleware/querymonth"
	"budgeting/internal/pkg/shiftpath"
	"fmt"
	"net/http"
	"text/template"
)

// TODO: Handler for View endpoints
// Call Controllers, then render the outputs
type ViewHandler struct {
	sdb db.DB
}

func NewViewHandler(sdb db.DB) http.Handler {

	return &ViewHandler{sdb}

}

func (h *ViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	head, tail := shiftpath.ShiftPath(r.URL.Path)

	switch head {
	case "":
		// Default to envelopes view
		http.Redirect(w, r, "/envelopes", http.StatusSeeOther)

	// Normal pages
	case "accounts":
		h.ServeHTTP_accounts(w, r)
	case "transactions":
		h.ServeHTTP_transactions(w, r)
	case "envelopes":
		h.ServeHTTP_envelopes(w, r)

	// Nested snippets
	case "view":

		head, tail = shiftpath.ShiftPath(tail)

		switch head {
		// Summary bar up top
		case "summary":
			h.ServeHTTP_snip_summary(w, r)

		// Account details
		case "account":
			h.ServeHTTP_snip_account(w, r, tail)

		// envelope details
		case "envelope":
			h.ServeHTTP_snip_envelope(w, r, tail)

		case "transaction":
			h.ServeHTTP_snip_transaction(w, r, tail)

		default:
			http.NotFound(w, r)
		}

	// Anything else, 404
	default:
		http.NotFound(w, r)
	}
}

func (h *ViewHandler) ServeHTTP_accounts(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/accounts"))

	// TODO: Render and return account list
}

func (h *ViewHandler) ServeHTTP_transactions(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/transactions"))

	// TODO: Render and return transaction list
}

func (h *ViewHandler) ServeHTTP_envelopes(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("/envelopes"))

	// TODO: Render and return envelope list
}

func (h *ViewHandler) ServeHTTP_snip_summary(w http.ResponseWriter, r *http.Request) {
	// TODO: Render and return summary bar

	month := querymonth.GetQM(r)

	summ, err := h.sdb.GetOverallSummary(bcdate.BCDate(month))

	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	// TODO: Load all templates on startup
	// This is faster for debugging though..
	tmpl, err := template.New("summary").ParseFiles("web/template/summary.tmpl")
	if err != nil {
		panic(fmt.Errorf("failed to parse template -- %w", err))
	}
	err = tmpl.Execute(w, summ)
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}
}

func (h *ViewHandler) ServeHTTP_snip_account(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/view/account"))

	// TODO: Render and return account detail and transactions
}

func (h *ViewHandler) ServeHTTP_snip_envelope(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/view/envelope"))

	// TODO: Render and return envelope transactions
}

func (h *ViewHandler) ServeHTTP_snip_transaction(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/view/transaction"))

	// TODO: Render and return entry form to add or update a transaction
}
