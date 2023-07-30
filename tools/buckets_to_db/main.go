package main

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/db"
	"budgeting/internal/pkg/model"
	"database/sql"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var bucketspostdate *regexp.Regexp = regexp.MustCompile(`^(2(?:0|1)[\d]{2})-(0[1-9]|1[0-2])-(0[1-9]|1[0-9]|2[0-9]|3[0-1])`)

// Tool to import a Buckets database file to our format

func printUsage() {
	log.Print("Usages:")
	log.Print("Re-init file:")
	log.Print("migrate <bucketsfile> <outputfile>")

	os.Exit(1)
}

func main() {
	if len(os.Args) != 3 {
		log.Print("ERROR: Incorrect arguments provided")
		printUsage()
	}

	buckets := os.Args[1]
	outfile := os.Args[2]

	var sdb *db.SQLite = &db.SQLite{}

	log.Printf("Open: %s", outfile)
	if err := sdb.Open(outfile); err != nil {
		log.Fatalf("Error opening DB: %s", err.Error())
	}

	if err := sdb.Init(); err != nil {
		log.Fatalf("Error init-ing DB file: %s", err.Error())
	}

	bdb, err := sql.Open("sqlite3", buckets)
	if err != nil {
		log.Fatalf("Failed to open Buckets db file: %s", err.Error())
	}

	// Envelope Groups

	log.Printf("Migrate Envelope Groups")

	var eg_map map[int]model.PKEY = make(map[int]model.PKEY)

	sortcounter := 100

	rows, err := bdb.Query("SELECT * FROM bucket_group ORDER BY ranking ASC")
	if err != nil {
		log.Fatalf("Failed to query bucket_group: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var eg struct {
			ID      int
			Created string
			Name    string
			Sort    string
			Notes   string
		}
		if err := rows.Scan(
			&eg.ID,
			&eg.Created,
			&eg.Name,
			&eg.Sort,
			&eg.Notes,
		); err != nil {
			log.Fatalf("Failed to scan next bucket_group: %s", err.Error())
		}

		neg := model.EnvelopeGroup{
			Name: eg.Name,
			Sort: sortcounter,
		}

		err := sdb.NewEnvelopeGroup(&neg)
		if err != nil {
			log.Fatalf("Failed to add new EnvelopeGroup: %s", err.Error())
		}

		sortcounter += 100

		eg_map[eg.ID] = neg.ID

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after bucket_group rows: %s", err.Error())
	}

	// Accounts

	log.Printf("Migrate Accounts")

	var a_map map[int]model.PKEY = make(map[int]model.PKEY)
	var a_bal_map map[model.PKEY]int = make(map[model.PKEY]int)

	sortcounter = 100

	rows, err = bdb.Query("SELECT * FROM account")
	if err != nil {
		log.Fatalf("Failed to query account: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var a struct {
			ID        int
			Created   string
			Name      string
			Balance   int
			Currency  string
			ImportBal sql.NullString
			Closed    bool
			Notes     string
			Offbudget bool
			Kind      string
		}
		if err := rows.Scan(
			&a.ID,
			&a.Created,
			&a.Name,
			&a.Balance,
			&a.Currency,
			&a.ImportBal,
			&a.Closed,
			&a.Notes,
			&a.Offbudget,
			&a.Kind,
		); err != nil {
			log.Fatalf("Failed to scan next account: %s", err.Error())
		}

		na := model.Account{
			Hidden: a.Closed,
		}

		// Specific edits
		switch a.Kind {
		case "offbudget":
			na.Offbudget = true
		case "debt":
			na.Debt = true
		}

		parts := strings.Split(a.Name, ":")
		if len(parts) == 2 {
			na.Institution = parts[0]
			na.Name = parts[1]
		} else {
			na.Name = a.Name
		}

		// TODO: Account class

		err := sdb.NewAccount(&na)
		if err != nil {
			log.Fatalf("Failed to add new Account: %s", err.Error())
		}

		sortcounter += 100

		a_map[a.ID] = na.ID
		a_bal_map[na.ID] = a.Balance

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after account rows: %s", err.Error())
	}

	// Envelopes

	log.Printf("Migrate Envelopes")

	var e_map map[int]model.PKEY = make(map[int]model.PKEY)

	sortcounter = 100

	rows, err = bdb.Query("SELECT * FROM bucket WHERE debt_account_id IS NULL ORDER BY ranking ASC")
	if err != nil {
		log.Fatalf("Failed to query bucket: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var e struct {
			ID            int
			Created       string
			Name          string
			Notes         string
			Balance       int
			Kicked        bool
			GroupID       sql.NullInt32
			Sort          string
			GoalType      string
			Goal          sql.NullInt32
			EndDate       sql.NullString
			Deposit       sql.NullInt32
			Color         string
			DebtAccountID sql.NullInt32
		}
		if err := rows.Scan(
			&e.ID,
			&e.Created,
			&e.Name,
			&e.Notes,
			&e.Balance,
			&e.Kicked,
			&e.GroupID,
			&e.Sort,
			&e.GoalType,
			&e.Goal,
			&e.EndDate,
			&e.Deposit,
			&e.Color,
			&e.DebtAccountID,
		); err != nil {
			log.Fatalf("Failed to scan next bucket: %s", err.Error())
		}

		ne := model.Envelope{
			DebtAccount: sql.NullInt32{Valid: false},
			Hidden:      e.Kicked,
			Name:        e.Name,
			Notes:       e.Notes,
			Sort:        sortcounter,
		}

		// Specific edits
		if e.GroupID.Valid {
			ne.GroupID = eg_map[int(e.GroupID.Int32)]
		} else {
			ne.GroupID = 1
		}

		switch e.GoalType {
		case "deposit":
			ne.Goal = model.GT_RECUR
			ne.GoalAmt = int(e.Deposit.Int32)
		case "goal-date":
			// This is how I use the goal-date type
			ne.Goal = model.GT_TGT
			ne.GoalTgt = int(e.Goal.Int32)
		case "goal-deposit":
			ne.Goal = model.GT_RECTIL
			ne.GoalAmt = int(e.Deposit.Int32)
			ne.GoalTgt = int(e.Goal.Int32)
		}

		err := sdb.NewEnvelope(&ne)
		if err != nil {
			log.Fatalf("Failed to add new Envelope: %s", err.Error())
		}

		sortcounter += 100

		e_map[e.ID] = ne.ID

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after bucket rows: %s", err.Error())
	}

	// Sepcial handling for debt envelopes

	log.Printf("Migrate Debt Envelopes")

	rows, err = bdb.Query("SELECT id, debt_account_id FROM bucket WHERE debt_account_id IS NOT NULL ORDER BY ranking ASC")
	if err != nil {
		log.Fatalf("Failed to query debt bucket: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var e struct {
			ID            int
			DebtAccountID int
		}
		if err := rows.Scan(
			&e.ID,
			&e.DebtAccountID,
		); err != nil {
			log.Fatalf("Failed to scan next debt bucket: %s", err.Error())
		}

		// Get debt envelope for this account -> TODO reduce duplication here

		ne, err := sdb.GetDebtEnvelopeFor(a_map[e.DebtAccountID])
		if err != nil {
			log.Fatalf("Failed to query debt envelope for account: %s", err.Error())
		}

		e_map[e.ID] = ne.ID

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after debt bucket rows: %s", err.Error())
	}

	// Account transactions

	toInsA := make([]model.AccountTransaction, 0)
	toInsE := make([]model.EnvelopeTransaction, 0)

	// AcctTrans LEFT JOIN BucketTrans.account_trans_id --> Account Transactions

	log.Printf("Migrate Account Transactions")

	rows, err = bdb.Query("SELECT AT.account_id, AT.posted, AT.amount, AT.memo, AT.general_cat, AT.cleared, BT.bucket_id FROM account_transaction AT LEFT JOIN bucket_transaction BT ON BT.account_trans_id = AT.id")
	if err != nil {
		log.Fatalf("Failed to query account transactions: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var t struct {
			AccountID int
			Posted    string
			Amount    int
			Memo      string
			Category  string
			Cleared   bool
			BucketID  sql.NullInt32
		}
		if err := rows.Scan(
			&t.AccountID,
			&t.Posted,
			&t.Amount,
			&t.Memo,
			&t.Category,
			&t.Cleared,
			&t.BucketID,
		); err != nil {
			log.Fatalf("Failed to scan next account transaction: %s", err.Error())
		}

		nt := model.AccountTransaction{
			AccountID: a_map[t.AccountID],
			Amount:    t.Amount,
			Cleared:   t.Cleared,
			Memo:      t.Memo,
		}

		if t.BucketID.Valid {
			nt.EnvelopeID.Valid = true
			nt.EnvelopeID.Int32 = int32(e_map[int(t.BucketID.Int32)])
		}

		switch t.Category {
		case "income":
			nt.Typ = model.TT_INCOME
		case "transfer":
			nt.Typ = model.TT_TRANSFER
		}

		match := bucketspostdate.FindStringSubmatch(t.Posted)
		if match == nil {
			log.Fatalf("Failed to match postdate with regex: %s", err.Error())
		}
		postdate, err := strconv.Atoi(match[1] + match[2] + match[3])
		if err != nil {
			log.Fatalf("Failed to convert matched postdate to int: %s", err.Error())
		}

		nt.PostDate = bcdate.BCDate(postdate)

		toInsA = append(toInsA, nt)

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after account transaction rows: %s", err.Error())
	}

	// BucketTrans.linked_trans_id IS NOT NULL --> Debt envelope transactions

	log.Printf("Migrate Debt Transactions")

	rows, err = bdb.Query("SELECT bucket_id, amount, posted FROM bucket_transaction WHERE linked_trans_id IS NOT NULL")
	if err != nil {
		log.Fatalf("Failed to query debt transactions: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var t struct {
			BucketID int
			Amount   int
			Posted   string
		}
		if err := rows.Scan(
			&t.BucketID,
			&t.Amount,
			&t.Posted,
		); err != nil {
			log.Fatalf("Failed to scan next debt transaction: %s", err.Error())
		}

		nt := model.EnvelopeTransaction{
			EnvelopeID: e_map[t.BucketID],
			Amount:     t.Amount,
		}

		match := bucketspostdate.FindStringSubmatch(t.Posted)
		if match == nil {
			log.Fatalf("Failed to match postdate with regex: %s", err.Error())
		}
		postdate, err := strconv.Atoi(match[1] + match[2] + match[3])
		if err != nil {
			log.Fatalf("Failed to convert matched postdate to int: %s", err.Error())
		}

		nt.PostDate = bcdate.BCDate(postdate)

		toInsE = append(toInsE, nt)

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after debt transaction rows: %s", err.Error())
	}

	// BucketTrans WHERE account_trans_id IS NULL AND linked_trans_id IS NULL --> Envelope only

	log.Printf("Migrate Envelope Transactions")

	rows, err = bdb.Query("SELECT bucket_id, amount, posted FROM bucket_transaction WHERE account_trans_id IS NULL AND linked_trans_id IS NULL")
	if err != nil {
		log.Fatalf("Failed to query envelope transactions: %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		var t struct {
			BucketID int
			Amount   int
			Posted   string
		}
		if err := rows.Scan(
			&t.BucketID,
			&t.Amount,
			&t.Posted,
		); err != nil {
			log.Fatalf("Failed to scan next envelope transaction: %s", err.Error())
		}

		nt := model.EnvelopeTransaction{
			EnvelopeID: e_map[t.BucketID],
			Amount:     t.Amount,
		}

		match := bucketspostdate.FindStringSubmatch(t.Posted)
		if match == nil {
			log.Fatalf("Failed to match postdate with regex: %s", err.Error())
		}
		postdate, err := strconv.Atoi(match[1] + match[2] + match[3])
		if err != nil {
			log.Fatalf("Failed to convert matched postdate to int: %s", err.Error())
		}

		nt.PostDate = bcdate.BCDate(postdate)

		toInsE = append(toInsE, nt)

	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Failed after envelope transaction rows: %s", err.Error())
	}

	// Do the batch inserstions to save some processing time

	log.Printf("Insert %d Account Transactions", len(toInsA))

	if err = sdb.Batch_NewAccountTransaction(toInsA); err != nil {
		log.Fatalf("Failed to insert all the Account Transactions: %s", err.Error())
	}

	log.Printf("Insert %d Envelope Transactions", len(toInsE))
	if err = sdb.Batch_NewEnvelopeTransaction(toInsE); err != nil {
		log.Fatalf("Failed to insert all the Envelope Transactions: %s", err.Error())
	}

	// Due to the ridiculous way Buckets stores the starting balance, we now have to calculate it
	// Get buckets balance, subtract from our calculated last balance and set starting bal

	log.Printf("Set starting balances for accounts")

	for aid, abal := range a_bal_map {
		asum, err := sdb.GetAccountSummary(bcdate.CurrentMonth(), aid)
		if err != nil {
			log.Fatalf("Failed to get current/latest account summary: %s", err.Error())
		}

		log.Printf("Set starting bal: %d = %d", aid, abal-asum.Bal)
		sdb.SetStartingBalance(aid, abal-asum.Bal)

	}

}
