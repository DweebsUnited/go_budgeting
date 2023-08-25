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
	"strconv"
)

// Handler for View endpoints
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
		http.Redirect(w, r, "/analysis", http.StatusSeeOther)

		// Normal pages
	case "envelopes":
		h.ServeHTTP_envelopes(w, r)
	case "envelope":
		h.ServeHTTP_envelope(w, r, tail)
	case "accounts":
		h.ServeHTTP_accounts(w, r)
	case "account":
		h.ServeHTTP_account(w, r, tail)
	case "transactions":
		h.ServeHTTP_transactions(w, r)
	case "analysis":
		h.ServeHTTP_analysis(w, r)

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

	month := bcdate.BCDate(querymonth.GetQM(r))

	type as struct {
		A model.Account
		S model.AccountSummary
	}

	aids := make([]model.PKEY, 0)

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
		aids = append(aids, acct.ID)
	}

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "accounts.html", struct {
		URL  string
		QM   bcdate.BCDate
		S    model.Summary
		AS   map[model.PKEY]as
		AIDs []model.PKEY
	}{
		URL:  "/accounts",
		QM:   month,
		S:    summ,
		AS:   acctSumm,
		AIDs: aids,
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

}

func (h *ViewHandler) ServeHTTP_transactions(w http.ResponseWriter, r *http.Request) {
	// Render and return transaction list

	month := bcdate.BCDate(querymonth.GetQM(r))

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	as := make(map[model.PKEY]string)
	es := make(map[model.PKEY]string)

	if accts, err := h.sdb.GetAccounts(); err == nil {
		for _, acct := range accts {
			as[acct.ID] = acct.Name
		}
	} else {
		panic(fmt.Errorf("failed to get account list -- %w", err))
	}

	if envs, err := h.sdb.GetEnvelopes(); err == nil {
		for _, env := range envs {
			es[env.ID] = env.Name
		}
	} else {
		panic(fmt.Errorf("failed to get envelope list -- %w", err))
	}

	atList, err := h.sdb.GetAllTransactions(month)
	if err != nil {
		panic(fmt.Errorf("failed to get transaction list -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "transactions.html", struct {
		URL string
		QM  bcdate.BCDate
		S   model.Summary
		AS  map[model.PKEY]string
		ES  map[model.PKEY]string
		ATs []model.AccountTransaction
	}{
		URL: "/transactions",
		QM:  month,
		S:   summ,
		AS:  as,
		ES:  es,
		ATs: atList,
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

}

func (h *ViewHandler) ServeHTTP_envelopes(w http.ResponseWriter, r *http.Request) {
	// Render and return envelope list

	month := bcdate.BCDate(querymonth.GetQM(r))

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	type esum struct {
		E model.Envelope
		S model.EnvelopeSummary
	}
	type ege struct {
		G  model.EnvelopeGroup
		GS model.EnvelopeSummary
		GE model.Envelope
		Es []esum
	}

	eges := make(map[model.PKEY]ege)
	egids := make([]model.PKEY, 0)

	egs, err := h.sdb.GetEnvelopeGroups()
	if err != nil {
		panic(fmt.Errorf("failed to get envelope groups -- %w", err))
	}

	for _, eg := range egs {
		summ := model.EnvelopeSummary{}

		es, err := h.sdb.GetEnvelopesInGroup(eg.ID)
		if err != nil {
			panic(fmt.Errorf("failed to get envelopes in group -- %w", err))
		}

		ess := make([]esum, 0)

		ggoal := 0

		for _, e := range es {
			sum, err := h.sdb.GetEnvelopeSummary(month, e.ID)
			if err != nil {
				panic(fmt.Errorf("failed to get envelope summary -- %w", err))
			}

			ess = append(ess, esum{e, sum})

			summ.Bal += sum.Bal
			summ.In += sum.In
			summ.Out += sum.Out

			ggoal += e.GoalWant(sum.Bal)
		}

		eges[eg.ID] = ege{
			G:  eg,
			GS: summ,
			GE: model.Envelope{Goal: model.GT_TGT, GoalTgt: ggoal},
			Es: ess,
		}
		egids = append(egids, eg.ID)
	}

	err = h.tmpl.ExecuteTemplate(w, "envelopes.html", struct {
		URL  string
		QM   bcdate.BCDate
		S    model.Summary
		EGs  []model.PKEY
		EGEs map[model.PKEY]ege
	}{
		URL:  "/envelopes",
		QM:   month,
		S:    summ,
		EGs:  egids,
		EGEs: eges,
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

}

func (h *ViewHandler) ServeHTTP_snip_summary(w http.ResponseWriter, r *http.Request) {
	// Render and return summary bar

	month := bcdate.BCDate(querymonth.GetQM(r))

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "summary.html", summ)
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

}

func (h *ViewHandler) ServeHTTP_account(w http.ResponseWriter, r *http.Request, tail string) {
	// TODO: Render and return account detail and transactions

	id, _ := shiftpath.ShiftPath(tail)
	if len(id) == 0 {
		panic(fmt.Errorf("no account id provided"))
	}

	iid, err := strconv.Atoi(id)
	if err != nil {
		panic(fmt.Errorf("provided id is not integer -- %w", err))
	}

	month := bcdate.BCDate(querymonth.GetQM(r))

	envs, err := h.sdb.GetEnvelopes()
	if err != nil {
		panic(fmt.Errorf("failed to get envelope list -- %w", err))
	}

	envList := make(map[model.PKEY]string, len(envs))

	for _, env := range envs {
		envList[env.ID] = env.Name
	}

	acct, err := h.sdb.GetAccount(model.PKEY(iid))
	if err != nil {
		panic(fmt.Errorf("failed to get account list -- %w", err))
	}

	accts, err := h.sdb.GetAccountSummary(month, acct.ID)
	if err != nil {
		panic(fmt.Errorf("failed to get account summary -- %w", err))
	}

	trans, err := h.sdb.GetAccountTransactions(month, acct.ID)
	if err != nil {
		panic(fmt.Errorf("failed to get account transactions -- %w", err))
	}

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	err = h.tmpl.ExecuteTemplate(w, "account.html", struct {
		URL string
		QM  bcdate.BCDate
		S   model.Summary
		ES  map[model.PKEY]string
		A   model.Account
		AS  model.AccountSummary
		AT  []model.AccountTransaction
	}{
		URL: "/account/" + id,
		QM:  month,
		S:   summ,
		ES:  envList,
		A:   acct,
		AS:  accts,
		AT:  trans,
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}
}

func (h *ViewHandler) ServeHTTP_envelope(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/envelope/" + tail))

	// TODO: Render and return envelope transactions
}

func (h *ViewHandler) ServeHTTP_snip_transaction(w http.ResponseWriter, r *http.Request, tail string) {
	w.Write([]byte("/view/transaction"))

	// TODO: Render and return entry form to add or update a transaction
}

func (h *ViewHandler) ServeHTTP_analysis(w http.ResponseWriter, r *http.Request) {
	// Render and return envelope list

	month := bcdate.BCDate(querymonth.GetQM(r))

	summ, err := h.sdb.GetOverallSummary(month)
	if err != nil {
		panic(fmt.Errorf("failed to get overall summary from DB -- %w", err))
	}

	type Gauge struct {
		Value float32
		Limit float32
	}
	type Gauges struct {
		Spend Gauge
		Bills Gauge
		Gain  Gauge
	}

	gs := Gauges{}

	egs, err := h.sdb.GetEnvelopeGroups()
	if err != nil {
		panic(fmt.Errorf("failed to get envelope groups -- %w", err))
	}

	for _, eg := range egs {
		summ := model.EnvelopeSummary{}

		es, err := h.sdb.GetEnvelopesInGroup(eg.ID)
		if err != nil {
			panic(fmt.Errorf("failed to get envelopes in group -- %w", err))
		}

		ggoal := 0

		for _, e := range es {
			sum, err := h.sdb.GetEnvelopeSummary(month, e.ID)
			if err != nil {
				panic(fmt.Errorf("failed to get envelope summary -- %w", err))
			}

			summ.Bal += sum.Bal
			summ.In += sum.In
			summ.Out += sum.Out

			ggoal += e.GoalWant(sum.Bal)
		}

		switch eg.Name {
		case "Monthly Bills":
			gs.Bills.Value = float32(summ.Bal) / 100.0
			gs.Bills.Limit = float32(ggoal) / 100.0
		case "Spending Money":
			gs.Spend.Value = float32(summ.Bal) / 100.0
			gs.Spend.Limit = float32(ggoal) / 100.0
		}
	}

	gs.Gain.Value = float32(summ.Gain()) / 100.0
	gs.Gain.Limit = float32(summ.Income) / 100.0

	err = h.tmpl.ExecuteTemplate(w, "analysis.html", struct {
		URL string
		QM  bcdate.BCDate
		S   model.Summary
		G   Gauges
	}{
		URL: "/analysis",
		QM:  month,
		S:   summ,
		G:   gs,
	})
	if err != nil {
		panic(fmt.Errorf("failed to execute template -- %w", err))
	}

}
