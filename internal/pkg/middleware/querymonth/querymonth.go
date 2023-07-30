package querymonth

import (
	"budgeting/internal/pkg/bcdate"
	"context"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type queryMonthKeyType int

const queryMonthKey queryMonthKeyType = 0

type QueryMonth struct {
	next http.Handler
	re   *regexp.Regexp
}

func NewQueryMonth(next http.Handler) http.Handler {
	return &QueryMonth{next, regexp.MustCompile(`^(2(?:0|1)[\d]{2})-(0[1-9]|1[0-2])$`)}
}

func (h *QueryMonth) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if qm, ok := r.URL.Query()["qm"]; ok {
		if len(qm) > 0 {
			match := h.re.FindStringSubmatch(qm[0])
			if match != nil {
				month, err := strconv.Atoi(match[1] + match[2] + "00")
				if err == nil {
					log.Printf("Found querymonth: %s-%s =  %d", match[1], match[2], month)
					h.next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), queryMonthKey, int(month))))
					return
				}
			}
		}
	}

	// If valid was found, we returned above
	// Construct today
	qm := int(bcdate.CurrentMonth())
	log.Printf("Missing or mal-parsed date, use today: %d", qm)
	h.next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), queryMonthKey, qm)))

}

func GetQM(r *http.Request) int {
	if v := r.Context().Value(queryMonthKey); v == nil {
		// This will be picked up by the logger
		panic("QM not set on request context")
	} else {
		return v.(int)
	}
}
