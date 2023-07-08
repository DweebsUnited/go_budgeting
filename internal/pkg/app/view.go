package app

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/middleware/querymonth"
	"budgeting/internal/pkg/model"
	"budgeting/internal/pkg/shiftpath"
	"fmt"
	"html/template"
	"net/http"
)

// TODO: Handler for View endpoints
// Call Controllers, then render the outputs
type ViewHandler struct {
	sdb  db.DB
	tmpl *template.Template
}

func NewViewHandler(sdb db.DB) http.Handler {

	tmpl, err := template.New("View").
		Funcs(map[string]any{
			"FmtVal": model.FormatVal,
		}).
		ParseGlob("web/template/*.html")
	if err != nil {
		panic(fmt.Errorf("failed to parse templates -- %w", err))
	}

	return &ViewHandler{sdb, tmpl}

}

func (h *ViewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	head, tail := shiftpath.ShiftPath(r.URL.Path)

	// TODO: Remove this refresh on each request
	// This is faster for debugging when the templates are constantly changing
	tmpl, err := template.New("View").
		Funcs(map[string]any{
			"FmtVal": model.FormatVal,
		}).
		ParseGlob("web/template/*.html")
	if err != nil {
		panic(fmt.Errorf("failed to parse templates -- %w", err))
	}
	h.tmpl = tmpl

	switch head {
	case "":
		// Default to envelopes view
		http.Redirect(w, r, "/envelopes", http.StatusSeeOther)

	// Normal pages
	case "accounts":
		h.ServeHTTP_accounts(w, r)
	case "account":
		h.ServeHTTP_account(w, r, tail)
	case "transactions":
		h.ServeHTTP_transactions(w, r)
	case "envelopes":
		h.ServeHTTP_envelopes(w, r)
	case "envelope":
		h.ServeHTTP_envelope(w, r, tail)

	// Nested snippets
	case "view":

		head, tail = shiftpath.ShiftPath(tail)

		switch head {
		// Summary bar up top
		case "summary":
			h.ServeHTTP_snip_summary(w, r)

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
	// Render and return account list

	type as struct {
		A model.Account
		S model.AccountSummary
	}
	type asv struct {
		QM bcdate.BCDate
		AS map[model.PKEY]as
		S  model.Summary
	}

	month := bcdate.BCDate(querymonth.GetQM(r))

	accts, err := h.sdb.GetAccounts()
	if err != nil {
		panic(fmt.Errorf("failed to get account list -- %w", err))
	}

	acctSumm := make(map[model.PKEY]as, len(accts))

	for _, acct := range accts {
		s, err := h.sdb.GetAccountSummary(month, acct.ID)
		if err != nil {
			panic(fmt.Errorf("failed to get account summary -- %w", err))
		}

		acctSumm[acct.ID] = as{acct, s}
	}

	summ, err := h.sdb.GetOverallSummary(bcdate.BCDate(month))
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "accounts.html", asv{QM: month, AS: acctSumm, S: summ})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

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
	// Render and return summary bar

	month := querymonth.GetQM(r)

	summ, err := h.sdb.GetOverallSummary(bcdate.BCDate(month))
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "summary.html", summ)
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}
}

func (h *ViewHandler) ServeHTTP_account(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/account/" + tail))

	// TODO: Render and return account detail and transactions
}

func (h *ViewHandler) ServeHTTP_envelope(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/envelope/" + tail))

	// TODO: Render and return envelope transactions
}

func (h *ViewHandler) ServeHTTP_snip_transaction(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/view/transaction"))

	// TODO: Render and return entry form to add or update a transaction
}
