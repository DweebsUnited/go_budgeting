package main

import (
	"budgeting/internal/pkg/app"
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/middleware/logger"
	"budgeting/internal/pkg/middleware/querymonth"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {

	// TODO: Set up DB
	sdb := db.NewSQLite()
	err := sdb.Open("bin/db.db")
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err.Error())
	}

	// Set up top level muxer
	mux := http.NewServeMux()
	if mux == nil {
		log.Fatal("Failed to create a mux")
	}

	mux.HandleFunc("/now", func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte(time.Now().Format(time.RFC3339)))
	})
	mux.Handle("/uptime", NewUptimeHandler(time.Now()))

	// These set up their own muxers
	mux.Handle("/api/", http.StripPrefix("/api", app.NewAPIHandler(sdb)))
	mux.Handle("/", app.NewViewHandler(sdb))

	// Nearly done, static resources
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir("D:/Projects/go_budgeting/web/static"))))

	log.Println("Listening...")

	log.Fatal(http.ListenAndServe(":8000",
		logger.NewLogger(
			querymonth.NewQueryMonth(
				mux))))

}

type UptimeHandler struct {
	start time.Time
}

func NewUptimeHandler(t time.Time) http.Handler {
	return &UptimeHandler{t}
}

func (h *UptimeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(
		fmt.Sprintf("Uptime: %s -- Current Query Month: %d",
			time.Since(h.start),
			querymonth.GetQM(r))))
}
