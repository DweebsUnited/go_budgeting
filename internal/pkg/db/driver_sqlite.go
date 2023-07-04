package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"database/sql"
	"errors"
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

		return nil
	} else {
		return fmt.Errorf("failed to stat file: %w", err)
	}
}

func (s *SQLite) Init() error {
	if s.db == nil {
		return fmt.Errorf("cannot init DB before opening")
	}

	query, err := ioutil.ReadFile("init/sqlite3.sql")
	if err != nil {
		return fmt.Errorf("failed reading DB setup file: %w", err)
	}

	if _, err := s.db.Exec(string(query)); err != nil {
		return fmt.Errorf("failed running DB setup command: %w", err)
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
		if err := s.updateAccountSummaries(tx, bcdate.Epoch(), a.ID); err != nil {
			return fmt.Errorf("UpdateAccount.updateAccountSummaries -- %s", err.Error())
		}
		if err := s.updateSummaries(tx, bcdate.Epoch()); err != nil {
			return fmt.Errorf("UpdateAccount.updateSummaries -- %s", err.Error())
		}
	}

	_, err = tx.Exec("UPDATE a SET hidden = ?, offbudget = ?, debt = ?, institution = ?, name = ?, class = ? WHERE ID = ?", a.Hidden, a.Offbudget, a.Debt, a.Institution, a.Name, a.Class, a.ID)
	if err != nil {
		return fmt.Errorf("UpdateAccount.Update.a -- %w", err)
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

	_, err = tx.Exec("DELETE FROM a_t WHERE accountID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteAccount.Delete.a_t -- %w", err)
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
	row := s.db.QueryRow("INSERT INTO e_grp (name,sort) VALUES (?,?)", eg.Name, eg.Sort)
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
	_, err := s.db.Exec("DELETE FROM e_grp WHERE ID = ?", id)
	if err != nil {
		return fmt.Errorf("DeleteEnvelopeGroup.Delete.e_grp -- %w", err)
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
func (s *SQLite) NewEnvelope(*model.Envelope) error {
	return errors.New("not implemented")
}
func (s *SQLite) UpdateEnvelope(model.Envelope) error {
	return errors.New("not implemented")
}
func (s *SQLite) DeleteEnvelope(id model.PKEY) error {
	return errors.New("not implemented")
}

func (s *SQLite) GetAllTransactions(month bcdate.BCDate) ([]model.AccountTransaction, error) {
	ats := make([]model.AccountTransaction, 0)

	rows, err := s.db.Query("SELECT * FROM a_t WHERE mod(postDate,100) = ? ORDER BY postDate DESC", month)
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

	rows, err := s.db.Query("SELECT * FROM a_t WHERE accountID = ? AND mod(postDate,100) = ? ORDER BY postDate DESC", id, month)
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

func (s *SQLite) NewAccountTransaction(*model.AccountTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) UpdateAccountTransaction(model.AccountTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) DeleteAccountTransaction(id model.PKEY) error { return errors.New("not implemented") }

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

	rows, err := s.db.Query("SELECT * FROM e_t WHERE envelopeID = ? AND mod(postDate,100) = ? ORDER BY postDate DESC", id, month)
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

func (s *SQLite) NewEnvelopeTransaction(*model.EnvelopeTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) UpdateEnvelopeTransaction(model.EnvelopeTransaction) error {
	return errors.New("not implemented")
}
func (s *SQLite) DeleteEnvelopeTransaction(id model.PKEY) error { return errors.New("not implemented") }

func (s *SQLite) GetAccountSummary(month bcdate.BCDate, id model.PKEY) (model.AccountSummary, error) {
	summ := model.AccountSummary{}
	row := s.db.QueryRow("SELECT * FROM a_chk WHERE month <= ? AND accountID = ? ORDER BY month DESC LIMIT 1", month, id)
	if err := row.Scan(
		&summ.AccountID,
		&summ.Month,
		&summ.Bal,
		&summ.In,
		&summ.Out,
		&summ.Cleared,
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
	return errors.New("not implemented")
}
func (s *SQLite) updateEnvelopeSummaries(tx *sql.Tx, oldest bcdate.BCDate, eID model.PKEY) error {
	return errors.New("not implemented")
}
func (s *SQLite) updateSummaries(tx *sql.Tx, oldest bcdate.BCDate) error {
	return errors.New("not implemented")
}
