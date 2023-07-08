package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"database/sql"
	"fmt"
	"io/ioutil"

	"os"

	_ "github.com/mattn/go-sqlite3"
)

// Glue between our DB and the SQLite driver
type SQLite struct {
	db *sql.DB
}

func NewSQLite() DB {
	return &SQLite{nil}
}

// TODO: Pass over all calls and queries to use NullXxx variables instead

func (s *SQLite) Open(dbname string) error {
	_, err := os.Stat(dbname)

	if err == nil || os.IsNotExist(err) {
		s.db, err = sql.Open("sqlite3", dbname)
		if err != nil {
			return fmt.Errorf("failed to open db file: %w", err)
		}

		if os.IsNotExist(err) {
			return s.Init()
		}

		_, err = s.db.Exec("PRAGMA foreign_keys = ON")
		if err != nil {
			return fmt.Errorf("failed to enable foreign key handling: %w", err)
		}

		return nil
	} else {
		return fmt.Errorf("failed to stat file: %w", err)
	}
}

func (s *SQLite) Init() error {
	if s.db == nil {
		return fmt.Errorf("cannot init DB before opening")
	}

	if _, err := s.db.Exec(string("PRAGMA foreign_keys = OFF")); err != nil {
		return fmt.Errorf("failed disabling foreign keys: %w", err)
	}

	query, err := ioutil.ReadFile("init/sqlite3.sql")
	if err != nil {
		return fmt.Errorf("failed reading DB setup file: %w", err)
	}

	if _, err := s.db.Exec(string(query)); err != nil {
		return fmt.Errorf("failed running DB setup command: %w", err)
	}

	if _, err := s.db.Exec(string("PRAGMA foreign_keys = ON")); err != nil {
		return fmt.Errorf("failed re-enabling foreign keys: %w", err)
	}

	return nil
}

func (s *SQLite) GetAccounts() ([]model.Account, error) {
	accts := make([]model.Account, 0)

	rows, err := s.db.Query("SELECT * FROM a ORDER BY institution ASC, name ASC")
	if err != nil {
		return nil, fmt.Errorf("GetAccounts.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var acct model.Account
		if err := rows.Scan(
			&acct.ID,
			&acct.Hidden,
			&acct.Offbudget,
			&acct.Debt,
			&acct.Institution,
			&acct.Name,
			&acct.Class,
		); err != nil {
			return nil, fmt.Errorf("GetAccounts.Scan -- %w", err)
		}
		accts = append(accts, acct)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAccounts.Err -- %w", err)
	}
	return accts, nil
}
func (s *SQLite) GetAccount(id model.PKEY) (model.Account, error) {
	a := model.Account{}
	row := s.db.QueryRow("SELECT * FROM a WHERE ID = ?", id)
	if err := row.Scan(
		&a.ID,
		&a.Hidden,
		&a.Offbudget,
		&a.Debt,
		&a.Institution,
		&a.Name,
		&a.Class,
	); err != nil {
		return a, fmt.Errorf("GetAccount.Scan.a -- %w", err)
	}

	return a, nil
}
func (s *SQLite) NewAccount(a *model.Account) error {
	var id int

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("NewAccount.Begin-- %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow("INSERT INTO a (hidden,offbudget,debt,institution,name,class) VALUES (?,?,?,?,?,?) RETURNING ID", a.Hidden, a.Offbudget, a.Debt, a.Institution, a.Name, a.Class)
	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("NewAccount.Insert.a.Scan -- %w", err)
	}
	a.ID = model.PKEY(id)

	if a.Debt {
		if err := s.newDebtEnvelope(tx, a.ID, a.DebtEnvelopeName()); err != nil {
			return fmt.Errorf("NewAccount.newDebtEnvelope -- %s", err.Error())
		}
	}

	_, err = tx.Exec("INSERT INTO a_chk (accountID,month,bal) VALUES (?,?,?)", id, bcdate.Epoch(), 0)
	if err != nil {
		return fmt.Errorf("NewAccount.Insert.a_chk -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NewAccount.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) UpdateAccount(a model.Account) error {
	var oldDebt bool

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("UpdateAccount.Begin-- %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow("SELECT debt FROM a WHERE ID = ?", a.ID)
	if err := row.Scan(&oldDebt); err != nil {
		return fmt.Errorf("UpdateAccount.Scan.a -- %w", err)
	}

	_, err = tx.Exec("UPDATE a SET hidden = ?, offbudget = ?, debt = ?, institution = ?, name = ?, class = ? WHERE ID = ?", a.Hidden, a.Offbudget, a.Debt, a.Institution, a.Name, a.Class, a.ID)
	if err != nil {
		return fmt.Errorf("UpdateAccount.Update.a -- %w", err)
	}

	if a.Debt {
		if !oldDebt {
			if err := s.newDebtEnvelope(tx, a.ID, a.DebtEnvelopeName()); err != nil {
				return fmt.Errorf("UpdateAccount.newDebtEnvelope -- %s", err.Error())
			}
		} else {
			if err := s.updateDebtEnvelope(tx, a.ID, a.DebtEnvelopeName()); err != nil {
				return fmt.Errorf("UpdateAccount.updateDebtEnvelope -- %s", err.Error())
			}
		}
	} else if !a.Debt && oldDebt {
		if err := s.deleteDebtEnvelope(tx, a.ID); err != nil {
			return fmt.Errorf("UpdateAccount.deleteDebtEnvelope -- %s", err.Error())
		}
		if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
			return fmt.Errorf("UpdateAccount.updateSummaries -- %s", err.Error())
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateAccount.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) DeleteAccount(id model.PKEY) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteAccount.Begin-- %w", err)
	}
	defer tx.Rollback()

	var debt bool
	row := tx.QueryRow("SELECT debt FROM a WHERE ID = ?", id)
	if err := row.Scan(&debt); err != nil {
		return fmt.Errorf("DeleteAccount.Scan.a -- %w", err)
	}

	if debt {
		s.deleteDebtEnvelope(tx, id)
	}

	type atupdate struct {
		postdate   bcdate.BCDate
		envelopeID sql.NullInt32
	}
	atus := make([]atupdate, 0)
	rows, err := s.db.Query("SELECT min(postDate) AS postDate, envelopeID FROM a_t GROUP BY envelopeID")
	if err != nil {
		return fmt.Errorf("DeleteAccount.Select.a_t -- %w", err)
	}
	for rows.Next() {
		var atu atupdate
		if err := rows.Scan(
			&atu.postdate,
			&atu.envelopeID,
		); err != nil {
			return fmt.Errorf("DeleteAccount.Scan.a_t -- %w", err)
		}
		if atu.envelopeID.Valid {
			atus = append(atus, atu)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("DeleteAccount.Select.a_t.Err -- %w", err)
	}
	rows.Close()

	_, err = tx.Exec("DELETE FROM a_t WHERE accountID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteAccount.Delete.a_t -- %w", err)
	}

	for _, atu := range atus {
		s.updateEnvelopeSummaries(tx, atu.postdate, model.PKEY(atu.envelopeID.Int32))
	}

	_, err = tx.Exec("DELETE FROM a_chk WHERE accountID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteAccount.Delete.a_chk -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM a WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteAccount.Delete.a -- %w", err)
	}

	if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
		return fmt.Errorf("DeleteAccount.updateSummaries -- %s", err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteAccount.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) GetStartingBalance(id model.PKEY) (int, error) {
	var sbal int
	row := s.db.QueryRow("SELECT bal FROM a_chk WHERE accountID = ? AND month = ?", id, bcdate.Epoch())
	if err := row.Scan(&sbal); err != nil {
		return sbal, fmt.Errorf("GetStartingBalance.Scan.a_chk -- %w", err)
	}
	return sbal, nil
}
func (s *SQLite) SetStartingBalance(id model.PKEY, balance int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("SetStartingBalance.Begin-- %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE a_chk SET bal = ? WHERE accountID = ? AND month = ?", balance, id, bcdate.Epoch())
	if err != nil {
		return fmt.Errorf("SetStartingBalance.Update.a_chk -- %w", err)
	}

	if err := s.updateAccountSummaries(tx, bcdate.Epoch(), id); err != nil {
		return fmt.Errorf("SetStartingBalance.updateAccountSummaries -- %s", err.Error())
	}
	if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
		return fmt.Errorf("SetStartingBalance.updateSummaries -- %s", err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("SetStartingBalance.Commit -- %w", err)
	}
	return nil
}

func (s *SQLite) GetEnvelopeGroups() ([]model.EnvelopeGroup, error) {
	egs := make([]model.EnvelopeGroup, 0)

	rows, err := s.db.Query("SELECT * FROM e_grp ORDER BY sort ASC, name ASC")
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var eg model.EnvelopeGroup
		if err := rows.Scan(
			&eg.ID,
			&eg.Name,
			&eg.Sort,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
		}
		egs = append(egs, eg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopeGroups -- %w", err)
	}
	return egs, nil
}
func (s *SQLite) GetEnvelopeGroup(id model.PKEY) (model.EnvelopeGroup, error) {
	eg := model.EnvelopeGroup{}
	row := s.db.QueryRow("SELECT * FROM e_grp WHERE ID = ?", id)
	if err := row.Scan(
		&eg.ID,
		&eg.Name,
		&eg.Sort,
	); err != nil {
		return eg, fmt.Errorf("GetEnvelopeGroup.Scan.e_grp -- %w", err)
	}
	return eg, nil
}
func (s *SQLite) NewEnvelopeGroup(eg *model.EnvelopeGroup) error {
	var eid int
	row := s.db.QueryRow("INSERT INTO e_grp (name,sort) VALUES (?,?) RETURNING ID", eg.Name, eg.Sort)
	if err := row.Scan(&eid); err != nil {
		return fmt.Errorf("NewEnvelopeGroup.Insert.e_grp.Scan -- %w", err)
	}
	eg.ID = model.PKEY(eid)
	return nil
}
func (s *SQLite) UpdateEnvelopeGroup(eg model.EnvelopeGroup) error {
	_, err := s.db.Exec("UPDATE e_grp SET name = ?, sort = ? WHERE ID = ?", eg.Name, eg.Sort, eg.ID)
	if err != nil {
		return fmt.Errorf("UpdateEnvelopeGroup.Update.e_grp -- %w", err)
	}
	return nil
}
func (s *SQLite) DeleteEnvelopeGroup(id model.PKEY) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeGroup.Begin-- %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE e SET groupID = ? WHERE envelopeID = ?", 1, id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeGroup.Update.a_t -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM e_grp WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeGroup.Delete.e -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteEnvelopeGroup.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) GetEnvelopesInGroup(id model.PKEY) ([]model.Envelope, error) {
	es := make([]model.Envelope, 0)

	rows, err := s.db.Query("SELECT * FROM e WHERE groupID = ? ORDER BY sort ASC, name ASC", id)
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopesInGroup.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var e model.Envelope
		if err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.Hidden,
			&e.DebtAccount,
			&e.Name,
			&e.Notes,
			&e.Goal,
			&e.GoalAmt,
			&e.GoalTgt,
			&e.Sort,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopesInGroup.Scan -- %w", err)
		}
		es = append(es, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopesInGroup.Err -- %w", err)
	}
	return es, nil
}

func (s *SQLite) GetEnvelopes() ([]model.Envelope, error) {
	es := make([]model.Envelope, 0)

	rows, err := s.db.Query("SELECT * FROM e ORDER BY sort ASC, name ASC")
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopes.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var e model.Envelope
		if err := rows.Scan(
			&e.ID,
			&e.GroupID,
			&e.Hidden,
			&e.DebtAccount,
			&e.Name,
			&e.Notes,
			&e.Goal,
			&e.GoalAmt,
			&e.GoalTgt,
			&e.Sort,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopes.Scan -- %w", err)
		}
		es = append(es, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopes.Err -- %w", err)
	}
	return es, nil
}
func (s *SQLite) GetEnvelope(id model.PKEY) (model.Envelope, error) {
	e := model.Envelope{}

	row := s.db.QueryRow("SELECT * FROM e WHERE ID = ?", id)
	if err := row.Scan(
		&e.ID,
		&e.GroupID,
		&e.Hidden,
		&e.DebtAccount,
		&e.Name,
		&e.Notes,
		&e.Goal,
		&e.GoalAmt,
		&e.GoalTgt,
		&e.Sort,
	); err != nil {
		return e, fmt.Errorf("GetEnvelope.Scan.e -- %w", err)
	}
	return e, nil
}
func (s *SQLite) NewEnvelope(e *model.Envelope) error {
	var id int

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("NewEnvelope.Begin-- %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow("INSERT INTO e (groupID,hidden,name,notes,goalType,goalAmt,goalTgt,sort) VALUES (?,?,?,?,?,?,?,?) RETURNING ID", e.GroupID, e.Hidden, e.Name, e.Notes, e.Goal, e.GoalAmt, e.GoalTgt, e.Sort)
	if err := row.Scan(&id); err != nil {
		return fmt.Errorf("NewEnvelope.Insert.e.Scan -- %w", err)
	}
	e.ID = model.PKEY(id)

	_, err = tx.Exec("INSERT INTO e_chk (envelopeID,month,bal) VALUES (?,?,?)", id, bcdate.Epoch(), 0)
	if err != nil {
		return fmt.Errorf("NewEnvelope.Insert.e_chk -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NewEnvelope.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) UpdateEnvelope(e model.Envelope) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("UpdateEnvelope.Begin-- %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE e SET groupID = ?, hidden = ?, name = ?, notes = ?, goalType = ?, goalAmt = ?, goalTgt = ?, sort = ? WHERE ID = ?", e.GroupID, e.Hidden, e.Name, e.Notes, e.Goal, e.GoalAmt, e.GoalTgt, e.Sort, e.ID)
	if err != nil {
		return fmt.Errorf("UpdateEnvelope.Update.e -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateEnvelope.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) DeleteEnvelope(id model.PKEY) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteEnvelope.Begin-- %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE a_t SET envelopeID = ? WHERE envelopeID = ?", nil, id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelope.Update.a_t -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM e_t WHERE envelopeID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelope.Delete.e_t -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM e_chk WHERE envelopeID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelope.Delete.e_chk -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM e WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelope.Delete.e -- %w", err)
	}

	if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
		return fmt.Errorf("DeleteEnvelope.updateSummaries -- %s", err.Error())
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteEnvelope.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error) {
	ats := make([]model.AccountTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM a_t WHERE postDate-mod(postDate,100) = ? ORDER BY postDate DESC", month)
	if err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		at := model.AccountTransaction{}
		if err := rows.Scan(
			&at.ID,
			&at.AccountID,
			&at.Typ,
			&at.EnvelopeID,
			&at.PostDate,
			&at.Amount,
			&at.Cleared,
			&at.Memo,
		); err != nil {
			return nil, fmt.Errorf("GetAllTransactions.Scan -- %w", err)
		}
		ats = append(ats, at)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Err -- %w", err)
	}
	return ats, nil
}

func (s *SQLite) GetAllAccountTransactions(id model.PKEY) ([]model.AccountTransaction, error) {
	ats := make([]model.AccountTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM a_t WHERE accountID = ? ORDER BY postDate DESC", id)
	if err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		at := model.AccountTransaction{}
		if err := rows.Scan(
			&at.ID,
			&at.AccountID,
			&at.Typ,
			&at.EnvelopeID,
			&at.PostDate,
			&at.Amount,
			&at.Cleared,
			&at.Memo,
		); err != nil {
			return nil, fmt.Errorf("GetAllTransactions.Scan -- %w", err)
		}
		ats = append(ats, at)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Err -- %w", err)
	}
	return ats, nil
}
func (s *SQLite) GetAccountTransactions(month bcdate.BCDate, id model.PKEY) ([]model.AccountTransaction, error) {
	ats := make([]model.AccountTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM a_t WHERE accountID = ? AND postDate-mod(postDate,100) = ? ORDER BY postDate DESC", id, month)
	if err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		at := model.AccountTransaction{}
		if err := rows.Scan(
			&at.ID,
			&at.AccountID,
			&at.Typ,
			&at.EnvelopeID,
			&at.PostDate,
			&at.Amount,
			&at.Cleared,
			&at.Memo,
		); err != nil {
			return nil, fmt.Errorf("GetAllTransactions.Scan -- %w", err)
		}
		ats = append(ats, at)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllTransactions.Err -- %w", err)
	}
	return ats, nil
}

func (s *SQLite) NewAccountTransaction(at *model.AccountTransaction) error {
	var atid int
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("NewAccountTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow("INSERT INTO a_t (accountID,type,envelopeID,postDate,amount,cleared,memo) VALUES (?,?,?,?,?,?,?) RETURNING ID", at.AccountID, at.Typ, at.EnvelopeID, at.PostDate, at.Amount, at.Cleared, at.Memo)
	if err := row.Scan(&atid); err != nil {
		return fmt.Errorf("NewAccountTransaction.Insert.a_t.Scan -- %w", err)
	}
	at.ID = model.PKEY(atid)

	if at.EnvelopeID.Valid {
		if err := s.updateEnvelopeSummaries(tx, at.PostDate, model.PKEY(at.EnvelopeID.Int32)); err != nil {
			return fmt.Errorf("NewAccountTransaction.updateEnvelopeSummaries -- %w", err)
		}
	}
	if err := s.updateAccountSummaries(tx, at.PostDate, at.AccountID); err != nil {
		return fmt.Errorf("NewAccountTransaction.updateAccountSummaries -- %w", err)
	}
	if err := s.updateSummaries(tx, at.PostDate); err != nil {
		return fmt.Errorf("NewAccountTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NewAccountTransaction.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) UpdateAccountTransaction(at model.AccountTransaction) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("NewAccountTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	var oldest bcdate.BCDate
	var oldeid sql.NullInt32
	row := tx.QueryRow("SELECT postDate, envelopeID FROM a_t WHERE ID = ?", at.ID)
	if err := row.Scan(&oldest, &oldeid); err != nil {
		return fmt.Errorf("NewAccountTransaction.Select.a_t.Scan -- %w", err)
	}
	oldest = bcdate.Oldest(oldest, at.PostDate)

	_, err = tx.Exec("UPDATE a_t SET accountID = ?, type = ?, envelopeID = ?, postDate = ?, amount = ?, cleared = ?, memo = ? WHERE ID = ?", at.AccountID, at.Typ, at.EnvelopeID, at.PostDate, at.Amount, at.Cleared, at.Memo, at.ID)
	if err != nil {
		return fmt.Errorf("NewAccountTransaction.Update.a_t -- %w", err)
	}

	if at.EnvelopeID.Valid && (!oldeid.Valid || oldeid.Int32 != at.EnvelopeID.Int32) {
		if err := s.updateEnvelopeSummaries(tx, oldest, model.PKEY(at.EnvelopeID.Int32)); err != nil {
			return fmt.Errorf("NewAccountTransaction.updateEnvelopeSummaries -- %w", err)
		}
	}
	if oldeid.Valid && (!at.EnvelopeID.Valid || oldeid.Int32 != at.EnvelopeID.Int32) {
		if err := s.updateEnvelopeSummaries(tx, oldest, model.PKEY(oldeid.Int32)); err != nil {
			return fmt.Errorf("NewAccountTransaction.updateEnvelopeSummaries.oldeid -- %w", err)
		}
	}

	if err := s.updateAccountSummaries(tx, oldest, at.AccountID); err != nil {
		return fmt.Errorf("NewAccountTransaction.updateAccountSummaries -- %w", err)
	}
	if err := s.updateSummaries(tx, oldest); err != nil {
		return fmt.Errorf("DeleteAccountTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NewAccountTransaction.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) DeleteAccountTransaction(id model.PKEY) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteAccountTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	var eid sql.NullInt32
	var aid model.PKEY
	var postdate bcdate.BCDate
	row := tx.QueryRow("SELECT envelopeID, accountID, postDate FROM a_t WHERE ID = ?", id)
	if err := row.Scan(&eid, &aid, &postdate); err != nil {
		return fmt.Errorf("DeleteAccountTransaction.Select.a_t.Scan -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM a_t WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteAccountTransaction.Update.a_t -- %w", err)
	}

	if eid.Valid {
		if err := s.updateEnvelopeSummaries(tx, postdate, model.PKEY(eid.Int32)); err != nil {
			return fmt.Errorf("DeleteAccountTransaction.updateEnvelopeSummaries -- %w", err)
		}
	}

	if err := s.updateAccountSummaries(tx, postdate, aid); err != nil {
		return fmt.Errorf("DeleteAccountTransaction.updateAccountSummaries -- %w", err)
	}
	if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
		return fmt.Errorf("DeleteAccountTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteAccountTransaction.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) GetAllEnvelopeTransactions(id model.PKEY) ([]model.EnvelopeTransaction, error) {
	ets := make([]model.EnvelopeTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM e_t WHERE envelopeID = ? ORDER BY postDate DESC", id)
	if err != nil {
		return nil, fmt.Errorf("GetAllEnvelopeTransactions.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		et := model.EnvelopeTransaction{}
		if err := rows.Scan(
			&et.ID,
			&et.EnvelopeID,
			&et.PostDate,
			&et.Amount,
		); err != nil {
			return nil, fmt.Errorf("GetAllEnvelopeTransactions.Scan -- %w", err)
		}
		ets = append(ets, et)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAllEnvelopeTransactions.Err -- %w", err)
	}
	return ets, nil
}
func (s *SQLite) GetEnvelopeTransactions(month bcdate.BCDate, id model.PKEY) ([]model.EnvelopeTransaction, error) {
	ets := make([]model.EnvelopeTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM e_t WHERE envelopeID = ? AND postDate-mod(postDate,100) = ? ORDER BY postDate DESC", id, month)
	if err != nil {
		return nil, fmt.Errorf("GetEnvelopeTransactions.Select -- %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		et := model.EnvelopeTransaction{}
		if err := rows.Scan(
			&et.ID,
			&et.EnvelopeID,
			&et.PostDate,
			&et.Amount,
		); err != nil {
			return nil, fmt.Errorf("GetEnvelopeTransactions.Scan -- %w", err)
		}
		ets = append(ets, et)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetEnvelopeTransactions.Err -- %w", err)
	}
	return ets, nil
}

func (s *SQLite) NewEnvelopeTransaction(et *model.EnvelopeTransaction) error {
	var etid int
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("NewEnvelopeTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow("INSERT INTO e_t (envelopeID,postDate,amount) VALUES (?,?,?) RETURNING ID", et.EnvelopeID, et.PostDate, et.Amount)
	if err := row.Scan(&etid); err != nil {
		return fmt.Errorf("NewEnvelopeTransaction.Insert.a_t.Scan -- %w", err)
	}
	et.ID = model.PKEY(etid)

	if err := s.updateEnvelopeSummaries(tx, et.PostDate, et.EnvelopeID); err != nil {
		return fmt.Errorf("NewEnvelopeTransaction.updateEnvelopeSummaries -- %w", err)
	}

	if err := s.updateSummaries(tx, et.PostDate); err != nil {
		return fmt.Errorf("NewEnvelopeTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NewEnvelopeTransaction.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) UpdateEnvelopeTransaction(et model.EnvelopeTransaction) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	var oldest bcdate.BCDate
	row := tx.QueryRow("SELECT postDate FROM e_t WHERE ID = ?", et.ID)
	if err := row.Scan(&oldest); err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.Select.e_t.Scan -- %w", err)
	}
	oldest = bcdate.Oldest(oldest, et.PostDate)

	_, err = tx.Exec("UPDATE e_t SET envelopeID = ?, postDate = ?, amount = ? WHERE ID = ?", et.EnvelopeID, et.PostDate, et.Amount, et.ID)
	if err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.Update.e_t -- %w", err)
	}

	if err := s.updateEnvelopeSummaries(tx, oldest, et.EnvelopeID); err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.updateEnvelopeSummaries -- %w", err)
	}

	if err := s.updateSummaries(tx, oldest); err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("UpdateEnvelopeTransaction.Commit -- %w", err)
	}

	return nil
}
func (s *SQLite) DeleteEnvelopeTransaction(id model.PKEY) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	var eid model.PKEY
	var postdate bcdate.BCDate
	row := tx.QueryRow("SELECT envelopeID, postDate FROM e_t WHERE ID = ?", id)
	if err := row.Scan(&eid, &postdate); err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.Select.e_t.Scan -- %w", err)
	}

	_, err = tx.Exec("DELETE FROM e_t WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.Delete.e_t -- %w", err)
	}

	if err := s.updateEnvelopeSummaries(tx, postdate, eid); err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.updateEnvelopeSummaries -- %w", err)
	}

	if err := s.updateSummaries(tx, postdate); err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("DeleteEnvelopeTransaction.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) GetAccountSummary(month bcdate.BCDate, id model.PKEY) (model.AccountSummary, error) {
	summ := model.AccountSummary{}
	row := s.db.QueryRow("SELECT * FROM a_chk WHERE month <= ? AND accountID = ? ORDER BY month DESC LIMIT 1", month, id)
	if err := row.Scan(
		&summ.AccountID,
		&summ.Month,
		&summ.Bal,
		&summ.In,
		&summ.Out,
		&summ.Uncleared,
	); err != nil {
		return summ, fmt.Errorf("GetAccountSummary.Scan.a_chk -- %w", err)
	}

	return summ, nil
}
func (s *SQLite) GetEnvelopeSummary(month bcdate.BCDate, id model.PKEY) (model.EnvelopeSummary, error) {
	summ := model.EnvelopeSummary{}
	row := s.db.QueryRow("SELECT * FROM e_chk WHERE month <= ? AND envelopeID = ? ORDER BY month DESC LIMIT 1", month, id)
	if err := row.Scan(
		&summ.EnvelopeID,
		&summ.Month,
		&summ.Bal,
		&summ.In,
		&summ.Out,
	); err != nil {
		return summ, fmt.Errorf("GetEnvelopeSummary.Scan.e_chk -- %w", err)
	}

	return summ, nil
}
func (s *SQLite) GetOverallSummary(month bcdate.BCDate) (model.Summary, error) {
	summ := model.Summary{}
	row := s.db.QueryRow("SELECT * FROM s_chk WHERE month <= ? ORDER BY month DESC LIMIT 1", month)
	if err := row.Scan(
		&summ.Month,
		&summ.Float,
		&summ.Income,
		&summ.Expenses,
		&summ.Delta,
		&summ.Banked,
		&summ.NetWorth,
	); err != nil {
		return summ, fmt.Errorf("GetOverallSummary.Scan.s_chk -- %w", err)
	}

	return summ, nil
}

func (s *SQLite) newDebtEnvelope(tx *sql.Tx, aID model.PKEY, eName string) error {
	var eid int

	row := tx.QueryRow("INSERT INTO e (groupID,debtAccount,Name) VALUES (?,?,?) RETURNING ID", 6, aID, eName)
	if err := row.Scan(&eid); err != nil {
		return fmt.Errorf("newDebtEnvelope.Insert.e.Scan -- %w", err)
	}
	_, err := tx.Exec("INSERT INTO e_chk (envelopeID,month,bal) VALUES (?,?,?)", eid, bcdate.Epoch(), 0)
	if err != nil {
		return fmt.Errorf("NewAccount.Insert.e_chk -- %w", err)
	}

	return nil
}
func (s *SQLite) updateDebtEnvelope(tx *sql.Tx, aID model.PKEY, eName string) error {
	_, err := tx.Exec("UPDATE e SET Name = ? WHERE debtAccount = ?", eName, aID)
	if err != nil {
		return fmt.Errorf("updateDebtEnvelope.Update.e -- %w", err)
	}

	return nil
}
func (s *SQLite) deleteDebtEnvelope(tx *sql.Tx, aID model.PKEY) error {
	var eid int

	row := tx.QueryRow("SELECT ID FROM e WHERE debtAccount = ?", aID)
	if err := row.Scan(&eid); err != nil {
		return fmt.Errorf("deleteDebtEnvelope.Select.e.Scan -- %w", err)
	}

	_, err := tx.Exec("DELETE FROM e WHERE ID = ?", eid)
	if err != nil {
		return fmt.Errorf("deleteDebtEnvelope.Delete.e -- %w", err)
	}
	_, err = tx.Exec("DELETE FROM e_chk WHERE envelopeID = ?", eid)
	if err != nil {
		return fmt.Errorf("deleteDebtEnvelope.Delete.e_chk -- %w", err)
	}

	return nil
}

func (s *SQLite) updateAccountSummaries(tx *sql.Tx, oldest bcdate.BCDate, aID model.PKEY) error {
	oldest = oldest - (oldest % 100)

	var oldest_t sql.NullInt32
	var oldest_m bcdate.BCDate
	err := tx.QueryRow("SELECT min(postDate) FROM a_t WHERE accountID = ? AND postDate >= ?", aID, oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateAccountSummaries.Select.a_t.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}
	err = tx.QueryRow("SELECT min(month) FROM a_chk WHERE accountID = ? AND month >= ? AND month > 0", aID, oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateAccountSummaries.Select.a_chk.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}

	if oldest == bcdate.Epoch() {
		return nil
	}

	var latest bcdate.BCDate = bcdate.CurrentMonth()
	var latest_t sql.NullInt32
	var latest_m bcdate.BCDate
	err = tx.QueryRow("SELECT max(postDate) FROM a_t WHERE accountID = ?", aID).Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateAccountSummaries.Select.a_t.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(month) FROM a_chk WHERE accountID = ? AND month > 0", aID).Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateAccountSummaries.Select.a_chk.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}

	for ; oldest <= latest; oldest += 100 {

		var lastbal int
		var bal int
		var in int
		var out int
		var uncleared int

		err := tx.QueryRow("SELECT bal FROM a_chk WHERE accountID = ? AND month < ? ORDER BY month DESC LIMIT 1", aID, oldest).Scan(&lastbal)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("updateAccountSummaries.Select.a_chk.lastbal -- %w", err)
		}

		err = tx.QueryRow("SELECT sum(amount) FROM a_t WHERE accountID = ? AND postDate-mod(postDate,100) = ? AND amount > 0", aID, oldest).Scan(&in)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("updateAccountSummaries.Select.a_t.in -- %w", err)
		}

		err = tx.QueryRow("SELECT sum(amount) FROM a_t WHERE accountID = ? AND postDate-mod(postDate,100) = ? AND amount < 0", aID, oldest).Scan(&out)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("updateAccountSummaries.Select.a_t.out -- %w", err)
		}
		bal = lastbal + in + out

		err = tx.QueryRow("SELECT sum(amount) FROM a_t WHERE accountID = ? AND postDate-mod(postDate,100) = ? AND cleared = 0", aID, oldest).Scan(&uncleared)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("updateAccountSummaries.Select.a_t.uncleared -- %w", err)
		}

		_, err = tx.Exec("INSERT OR REPLACE a_chk (accountID,month,bal,in,out,uncleared) VALUES (?,?,?,?,?,?)", aID, oldest, bal, in, out, uncleared)
		if err != nil {
			return fmt.Errorf("updateAccountSummaries.Replace.a_chk -- %w", err)
		}

	}

	return nil
}
func (s *SQLite) updateEnvelopeSummaries(tx *sql.Tx, oldest bcdate.BCDate, eID model.PKEY) error {
	oldest = oldest - (oldest % 100)

	var oldest_t sql.NullInt32
	var oldest_m bcdate.BCDate
	err := tx.QueryRow("SELECT min(postDate) FROM a_t WHERE envelopeID = ? AND postDate >= ?", eID, oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.a_t.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}
	err = tx.QueryRow("SELECT min(postDate) FROM e_t WHERE envelopeID = ? AND postDate >= ?", eID, oldest).Scan(&oldest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.e_t.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}
	err = tx.QueryRow("SELECT min(month) FROM e_chk WHERE envelopeID = ? AND month >= ? AND month > 0", eID, oldest).Scan(&oldest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.e_chk.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}

	if oldest == bcdate.Epoch() {
		return nil
	}

	var latest bcdate.BCDate = bcdate.CurrentMonth()
	var latest_t sql.NullInt32
	var latest_m bcdate.BCDate
	err = tx.QueryRow("SELECT max(postDate) FROM a_t WHERE envelopeID = ?", eID).Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.a_t.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(postDate) FROM e_t WHERE envelopeID = ?", eID).Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.e_t.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(month) FROM e_chk WHERE envelopeID = ? AND month > 0", eID).Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateEnvelopeSummaries.Select.e_chk.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}

	for ; oldest <= latest; oldest += 100 {

		var lastbal int
		var bal int
		var in int
		var in_a int
		var out int
		var out_a int

		if err := tx.QueryRow("SELECT bal FROM e_chk WHERE envelopeID = ? AND month < ? ORDER BY month DESC LIMIT 1", eID, oldest).Scan(&lastbal); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateEnvelopeSummaries.Select.e_chk.lastbal -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM a_t WHERE envelopeID = ? AND postDate-mod(postDate,100) = ? AND amount > 0", eID, oldest).Scan(&in_a); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateEnvelopeSummaries.Select.a_t.in -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM a_t WHERE envelopeID = ? AND postDate-mod(postDate,100) = ? AND amount < 0", eID, oldest).Scan(&out_a); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateEnvelopeSummaries.Select.a_t.out -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM e_t WHERE envelopeID = ? AND postDate-mod(postDate,100) = ? AND amount > 0", eID, oldest).Scan(&in); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateEnvelopeSummaries.Select.e_t.in -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM e_t WHERE envelopeID = ? AND postDate-mod(postDate,100) = ? AND amount < 0", eID, oldest).Scan(&out); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateEnvelopeSummaries.Select.e_t.out -- %w", err)
			}
		}
		bal = lastbal + in_a + out_a + in + out

		_, err = tx.Exec("INSERT OR REPLACE e_chk (envelopeID,month,bal,in,out) VALUES (?,?,?,?,?)", eID, oldest, bal, in, out)
		if err != nil {
			return fmt.Errorf("updateEnvelopeSummaries.Replace.e_chk -- %w", err)
		}

	}

	return nil
}
func (s *SQLite) updateSummaries(tx *sql.Tx, oldest bcdate.BCDate) error {
	oldest = oldest - (oldest % 100)

	var oldest_t sql.NullInt32
	var oldest_m bcdate.BCDate
	err := tx.QueryRow("SELECT min(postDate) FROM a_t WHERE postDate >= ?", oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.a_t.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}

	err = tx.QueryRow("SELECT min(postDate) FROM e_t WHERE postDate >= ?", oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.e_t.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}
	err = tx.QueryRow("SELECT min(month) FROM a_chk WHERE month >= ? AND month > 0", oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.a_chk.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}
	err = tx.QueryRow("SELECT min(month) FROM e_chk WHERE month >= ? AND month > 0", oldest).Scan(&oldest_t)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.e_chk.Min -- %w", err)
	}
	if oldest_t.Valid {
		oldest_m = bcdate.BCDate(oldest_t.Int32)
		oldest_m = oldest_m - (oldest_m % 100)

		oldest = bcdate.Oldest(oldest, oldest_m)
	}

	if oldest == bcdate.Epoch() {
		return nil
	}

	var latest bcdate.BCDate = bcdate.CurrentMonth()
	var latest_t sql.NullInt32
	var latest_m bcdate.BCDate
	err = tx.QueryRow("SELECT max(postDate) FROM a_t").Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.a_t.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(postDate) FROM e_t").Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.e_t.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(month) FROM e_chk WHERE month > 0").Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.e_chk.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(month) FROM a_chk WHERE month > 0").Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.a_chk.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}
	err = tx.QueryRow("SELECT max(month) FROM s_chk WHERE month > 0").Scan(&latest_t)
	if err != nil && err == sql.ErrNoRows {
		return fmt.Errorf("updateSummaries.Select.s_chk.Max -- %w", err)
	}
	if latest_t.Valid {
		latest_m = bcdate.BCDate(latest_t.Int32)
		latest_m = latest_m - (latest_m % 100)

		latest = bcdate.Latest(latest, latest_m)
	}

	for ; oldest <= latest; oldest += 100 {

		var a_bal int
		var e_bal int

		if err := tx.QueryRow("SELECT sum(bal) FROM ( SELECT bal, max(month) FROM a_chk JOIN a ON a_chk.accountID = a.ID WHERE month <= ? AND debt = 0 AND offbudget = 0 GROUP BY accountID )", oldest).Scan(&a_bal); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.a_chk.debt.bal -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(bal) FROM ( SELECT bal, max(month) FROM e_chk WHERE month <= ? GROUP BY envelopeID )", oldest).Scan(&e_bal); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.e_chk.bal -- %w", err)
			}
		}

		var float int = a_bal - e_bal

		var banked int
		if err := tx.QueryRow("SELECT sum(bal) FROM ( SELECT bal, max(month) FROM a_chk JOIN a ON a_chk.accountID = a.ID WHERE month <= ? AND offbudget = 0 GROUP BY accountID )", oldest).Scan(&banked); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.a_chk.debt.bal -- %w", err)
			}
		}

		var nw int
		if err := tx.QueryRow("SELECT sum(bal) FROM ( SELECT bal, max(month) FROM a_chk WHERE month <= ? GROUP BY accountID )", oldest).Scan(&nw); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.a_chk.banked -- %w", err)
			}
		}

		var inc int
		var exp int
		var delta int

		if err := tx.QueryRow("SELECT sum(amount) FROM a_t WHERE postDate-mod(postDate,100) = ? AND type = 1", oldest).Scan(&inc); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.inc -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM a_t WHERE postDate-mod(postDate,100) = ? AND type = 0 AND amount < 0", oldest).Scan(&exp); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.inc -- %w", err)
			}
		}
		if err := tx.QueryRow("SELECT sum(amount) FROM a_t WHERE postDate-mod(postDate,100) = ? AND type = 0", oldest).Scan(&delta); err != nil {
			if err != sql.ErrNoRows {
				return fmt.Errorf("updateSummaries.Select.inc -- %w", err)
			}
		}

		_, err = tx.Exec("INSERT OR REPLACE s_chk (month,float,income,expenses,delta,banked,netWorth) VALUES (?,?,?,?,?,?,?)", oldest, float, inc, exp, delta, banked, nw)
		if err != nil {
			return fmt.Errorf("updateEnvelopeSummaries.Replace.e_chk -- %w", err)
		}

	}

	return nil
}
