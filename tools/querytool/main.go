package main

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/model"
	"flag"
	"log"
	"os"
)

// Tool to query the DB without starting a webserver

func printUsage() {
	log.Print("Usages:")
	log.Print("Re-init file:")
	log.Print("querytool <dbfile> init")
	log.Print("Dump file contents:")
	log.Print("querytool <dbfile> dump")
	log.Print("Account:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) a [flags...]")
	log.Print("Envelope Group:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) e_grp [flags...]")
	log.Print("Envelope:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) e [flags...]")
	log.Print("Account Transaction:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) a_t [flags...]")
	log.Print("Envelope Transaction:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) e_t [flags...]")
	log.Print("Account Summaries:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) a_chk [flags...]")
	log.Print("Envelope Summaries:")
	log.Print("querytool <dbfile> (sel|ins|upd|del) e_chk [flags...]")

	os.Exit(1)
}

func main() {
	if len(os.Args) < 3 {
		log.Print("ERROR: Not enough arguments provided")
		printUsage()
	}

	dbname := os.Args[1]
	op := os.Args[2]

	var sdb db.DB = db.NewSQLite()

	log.Printf("Open: %s", dbname)
	if err := sdb.Open(dbname); err != nil {
		log.Fatalf("Error opening DB: %s", err.Error())
	}

	switch op {

	case "init":
		log.Printf("Reinit: %s", dbname)

		if err := sdb.Init(); err != nil {
			log.Fatalf("Error init-ing file: %s", err.Error())
		}

	case "dump":
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

	case "sel":
		handleDBOP(sdb, op, os.Args[3:])
	case "ins":
		handleDBOP(sdb, op, os.Args[3:])
	case "upd":
		handleDBOP(sdb, op, os.Args[3:])
	case "del":
		handleDBOP(sdb, op, os.Args[3:])

	default:
		log.Printf("ERROR: Unrecognized operation: %s", op)
		printUsage()

	}

}

func handleDBOP(sdb db.DB, op string, args []string) {

	switch args[0] {
	case "a":

		handleAccount(sdb, op, args[1:])

	case "e_grp":

		handleEnvelopeGroup(sdb, op, args[1:])

	case "e":

		switch op {
		case "sel":
		case "ins":
		case "upd":
		case "del":
		default:
		}

	case "a_t":

		switch op {
		case "sel":
		case "ins":
		case "upd":
		case "del":
		default:
		}

	case "e_t":

		switch op {
		case "sel":
		case "ins":
		case "upd":
		case "del":
		default:
		}

	case "a_chk":

		switch op {
		case "sel":
		case "ins":
		case "upd":
		case "del":
		default:
		}

	case "e_chk":

		switch op {
		case "sel":
		case "ins":
		case "upd":
		case "del":
		default:
		}

	default:
		log.Fatalf("Unrecognized item type")
	}

}

func handleAccount(sdb db.DB, op string, args []string) {

	fs := flag.NewFlagSet("Account", flag.ContinueOnError)
	id := fs.Int(
		"id",
		0,
		"ID          -- sel|   |upd|del")
	hide := fs.Bool(
		"hide",
		false,
		"Hidden?     --    |ins|upd|   ")
	offb := fs.Bool(
		"offb",
		false,
		"Offbudget?  --    |ins|upd|   ")
	debt := fs.Bool(
		"debt",
		false,
		"Debt Acct?  --    |ins|upd|   ")
	inst := fs.String(
		"inst",
		"",
		"Institution --    |ins|upd|   ")
	name := fs.String(
		"name",
		"",
		"Name        --    |ins|upd|   ")
	class := fs.Int(
		"class",
		0,
		"Class       --    |ins|upd|   ")
	sbal := fs.Int(
		"sbal",
		0,
		"Start Bal   --    |ins|upd|   ")

	fs.Parse(args)

	a := model.Account{
		ID:          model.PKEY(*id),
		Hidden:      *hide,
		Offbudget:   *offb,
		Debt:        *debt,
		Institution: *inst,
		Name:        *name,
		Class:       model.AccountClass(*class),
	}

	switch op {
	case "sel":
		// Need ID, ignore others
		if *id == 0 {
			log.Print("Error: To select, --id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		a, err := sdb.GetAccount(a.ID)
		if err != nil {
			log.Fatalf("Error getting account: %s", err.Error())
		}

		log.Print("Account result:")
		log.Printf("%s", a)

	case "ins":
		// ID will be overwritten
		err := sdb.NewAccount(&a)
		if err != nil {
			log.Fatalf("Error inserting account: %s", err.Error())
		}

		log.Print("Account:")
		log.Printf("%s", a)

		fs.Visit(func(f *flag.Flag) {
			if f.Name == "sbal" {
				err := sdb.SetStartingBalance(a.ID, *sbal)
				if err != nil {
					log.Fatalf("Error setting starting balance: %s", err.Error())
				}

				log.Printf("Set starting balance of %d to %d", a.ID, *sbal)

			}
		})

	case "upd":
		// Need ID to query, then overlay given
		if *id == 0 {
			log.Print("Error: To update, --id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		aDB, err := sdb.GetAccount(a.ID)
		if err != nil {
			log.Fatalf("Error getting account: %s", err.Error())
		}

		fs.Visit(func(f *flag.Flag) {
			switch f.Name {
			case "hide":
				aDB.Hidden = a.Hidden
			case "offb":
				aDB.Offbudget = a.Offbudget
			case "debt":
				aDB.Debt = a.Debt
			case "inst":
				aDB.Institution = a.Institution
			case "name":
				aDB.Name = a.Name
			case "class":
				aDB.Class = a.Class
			case "sbal":
				err := sdb.SetStartingBalance(a.ID, *sbal)
				if err != nil {
					log.Fatalf("Error setting starting balance: %s", err.Error())
				}

				log.Printf("Set starting balance of %d to %d", a.ID, *sbal)
			}
		})

		if err := sdb.UpdateAccount(aDB); err != nil {
			log.Fatalf("Error updating account: %s", err.Error())
		}

		log.Print("Updated account:")
		log.Printf("%s", aDB)

	case "del":
		// Need ID, ignore others
		if *id == 0 {
			log.Print("Error: To delete, --id is required")
			fs.PrintDefaults()
			os.Exit(1)
		}

		if err := sdb.DeleteAccount(a.ID); err != nil {
			log.Fatalf("Error deleting account: %s", err.Error())
		}

		log.Print("Deleted account")

	default:
	}

}

func handleEnvelopeGroup(sdb db.DB, op string, args []string) {
	fs := flag.NewFlagSet("EnvelopeGroup", flag.ContinueOnError)

	id := fs.Int(
		"id",
		0,
		"ID   -- sel|   |upd|del")
	name := fs.String(
		"name",
		"",
		"Name --    |ins|upd|   ")
	sort := fs.Int(
		"sort",
		0,
		"Sort --    |ins|upd|   ")

	fs.Parse(args)

	_ = model.EnvelopeGroup{
		ID:   model.PKEY(*id),
		Name: *name,
		Sort: *sort,
	}

	switch op {
	case "sel":
		// Need ID, ignore others
	case "ins":
		// Need all but ID
	case "upd":
		// Need ID to query, then overlay given
	case "del":
		// Need ID, ignore others
	default:
	}
}
