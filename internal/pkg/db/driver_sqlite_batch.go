package db

import (
	"budgeting/internal/pkg/bcdate"
	"budgeting/internal/pkg/model"
	"fmt"
)

func (s *SQLite) Batch_NewAccountTransaction(ats []model.AccountTransaction) error {
	var atid int
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("Batch_NewAccountTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	aids := make([]model.PKEY, 0)
	eids := make([]model.PKEY, 0)

	oldest := bcdate.CurrentMonth()

	for _, at := range ats {
		row := tx.QueryRow("INSERT INTO a_t (accountID,type,envelopeID,postDate,amount,cleared,memo) VALUES (?,?,?,?,?,?,?) RETURNING ID", at.AccountID, at.Typ, at.EnvelopeID, at.PostDate, at.Amount, at.Cleared, at.Memo)
		if err := row.Scan(&atid); err != nil {
			return fmt.Errorf("Batch_NewAccountTransaction.Insert.a_t.Scan -- %w", err)
		}

		aids = append(aids, at.AccountID)

		if at.EnvelopeID.Valid {
			eids = append(eids, model.PKEY(at.EnvelopeID.Int32))
		}

		oldest = bcdate.Oldest(oldest, at.PostDate)
	}

	for _, eid := range eids {
		if err := s.updateEnvelopeSummaries(tx, oldest, eid); err != nil {
			return fmt.Errorf("Batch_NewAccountTransaction.updateEnvelopeSummaries -- %w", err)
		}
	}
	for _, aid := range aids {
		if err := s.updateAccountSummaries(tx, oldest, aid); err != nil {
			return fmt.Errorf("Batch_NewAccountTransaction.updateAccountSummaries -- %w", err)
		}
	}
	if err := s.updateSummaries(tx, oldest); err != nil {
		return fmt.Errorf("Batch_NewAccountTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Batch_NewAccountTransaction.Commit -- %w", err)
	}

	return nil
}

func (s *SQLite) Batch_NewEnvelopeTransaction(ets []model.EnvelopeTransaction) error {
	var etid int
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("Batch_NewEnvelopeTransaction.Begin -- %w", err)
	}
	defer tx.Rollback()

	eids := make([]model.PKEY, 0)

	oldest := bcdate.CurrentMonth()

	for _, et := range ets {
		row := tx.QueryRow("INSERT INTO e_t (envelopeID,postDate,amount) VALUES (?,?,?) RETURNING ID", et.EnvelopeID, et.PostDate, et.Amount)
		if err := row.Scan(&etid); err != nil {
			return fmt.Errorf("Batch_NewEnvelopeTransaction.Insert.a_t.Scan -- %w", err)
		}

		eids = append(eids, et.EnvelopeID)

		oldest = bcdate.Oldest(oldest, et.PostDate)
	}

	for _, eid := range eids {
		if err := s.updateEnvelopeSummaries(tx, oldest, eid); err != nil {
			return fmt.Errorf("Batch_NewEnvelopeTransaction.updateEnvelopeSummaries -- %w", err)
		}
	}

	if err := s.updateSummaries(tx, oldest); err != nil {
		return fmt.Errorf("Batch_NewEnvelopeTransaction.updateSummaries -- %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("Batch_NewEnvelopeTransaction.Commit -- %w", err)
	}

	return nil
}
