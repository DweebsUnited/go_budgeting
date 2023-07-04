package main

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/db"
	"flag"
	"log"
)

// TODO: Tool to query the DB without starting a webserver

func main() {
	dbname := flag.String("db", "", "DB file to open / create")
	toInit := flag.Bool("i", false, "Re-init the DB")
	toStat := flag.Bool("v", false, "Dump DB statistics")

	flag.Parse()

	var sdb db.DB
	if *toInit {
		log.Printf("Force create: %s", *dbname)
		sdb = db.NewSQLite(*dbname)
	} else {
		log.Printf("Open/Create: %s", *dbname)
		sdb = db.OpenSQLite(*dbname)
	}

	if *toStat {

		log.Print("Accounts in DB:")

		accts, err := sdb.GetAccounts()

		if err != nil {
			log.Fatalf("Error getting Accounts: %s", err.Error())
		}

		for _, acct := range accts {
			log.Printf("%s", acct)

			s, err := sdb.GetAccountSummary(bcdate.CurrentMonth(), acct.ID)
			if err != nil {
				log.Fatalf("Error getting Account summary: %s", err.Error())
			}
			log.Printf("\t%s", s)
		}

		log.Print("Envelope Groups in DB:")

		egs, err := sdb.GetEnvelopeGroups()

		for _, eg := range egs {
			log.Printf("%s", eg)

			es, err := sdb.GetEnvelopesInGroup(eg.ID)
			if err != nil {
				log.Fatalf("Error getting Envelopes in Group: %s", err.Error())
			}

			for _, e := range es {
				log.Printf("\t%s", e)

				s, err := sdb.GetEnvelopeSummary(bcdate.CurrentMonth(), e.ID)
				if err != nil {
					log.Fatalf("Error getting Envelope Summary: %s", err.Error())
				}

				log.Printf("\t\t%s", s)
			}
		}

		if err != nil {
			log.Fatalf("Error getting Envelope Groups: %s", err.Error())
		}

	}

}
