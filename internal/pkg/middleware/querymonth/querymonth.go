package querymonth

import (
	"context"
	"net/http"
	"regexp"
	"time"
)

type queryMonthKeyType int

const queryMonthKey queryMonthKeyType = 0

type QueryMonth struct {
	next http.Handler
	re   *regexp.Regexp
}

func NewQueryMonth(next http.Handler) http.Handler {
	return &QueryMonth{next, regexp.MustCompile(`20[\d]{2}-(?:0[1-9]|1[0-2])`)}
}

func (h *QueryMonth) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if qm, ok := r.URL.Query()["qm"]; ok {
		if len(qm) > 0 && h.re.MatchString(qm[0]) {
			h.next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), queryMonthKey, qm[0])))
			return
		}
	}

	// If valid was found, we returned above
	// Construct today
	qm := time.Now().Format("2006-01")
	h.next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), queryMonthKey, qm)))

}

func GetQM(r *http.Request) string {
	if v := r.Context().Value(queryMonthKey); v == nil {
		// This will be picked up by the logger
		panic("QM not set on request context")
	} else {
		return v.(string)
	}
}
