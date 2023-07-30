package model

import "fmt"

func (ac AccountClass) String() string {
	switch ac {
	case AT_CHECKING:
		return "Checking"
	case AT_SAVINGS:
		return "Savings"
	case AT_INVESTMENT:
		return "Investment"
	case AT_LOAN:
		return "Loan"
	case AT_CREDITCARD:
		return "CreditCard"
	default:
		return "UNKNOWN"
	}
}

func (a Account) String() string {
	ret := fmt.Sprintf("%03d: %5s:%20s -- ", a.ID, a.Institution, a.Name)
	if a.Hidden {
		ret += "H"
	} else {
		ret += " "
	}
	ret += ":"
	if a.Offbudget {
		ret += "O"
	} else {
		ret += " "
	}
	ret += ":"
	if a.Debt {
		ret += "D"
	} else {
		ret += " "
	}
	ret += ":"
	ret += a.Class.String()
	return ret
}

func (tt TransactionType) String() string {
	switch tt {
	case TT_NORM:
		return ""
	case TT_INCOME:
		return "Income"
	case TT_TRANSFER:
		return "Transfer"
	case TT_ADJUST:
		return "Adjust"
	default:
		return "UNKNOWN"
	}
}

func (a Account) DebtEnvelopeName() string {
	return "Debt Account: " + a.Institution + ":" + a.Name
}

func (s AccountSummary) String() string {
	return fmt.Sprintf("%03d -- %08d -- %05d / %05d --  ->%05d  <-%05d", s.AccountID, s.Month, s.Bal, s.Uncleared, s.In, s.Out)
}

func (eg EnvelopeGroup) String() string {
	return fmt.Sprintf("%03d: %20s", eg.ID, eg.Name)
}

func (e Envelope) String() string {
	ret := fmt.Sprintf("%03d.%03d %20s -- ", e.GroupID, e.ID, e.Name)
	if e.Hidden {
		ret += "H"
	} else {
		ret += " "
	}
	ret += " -- "
	switch e.Goal {
	case GT_NONE:
		ret += " "
	case GT_RECUR:
		ret += "R"
	case GT_TGT:
		ret += "G"
	case GT_RECTIL:
		ret += "T"
	default:
		ret += "X"
	}
	ret += fmt.Sprintf("=%d/%d", e.GoalAmt, e.GoalTgt)
	if e.DebtAccount.Valid {
		ret += fmt.Sprintf(" -> %03d", e.DebtAccount.Int32)
	}
	return ret
}

func (s EnvelopeSummary) String() string {
	return fmt.Sprintf("%03d -- %08d -- %05d --  ->%05d  <-%05d", s.EnvelopeID, s.Month, s.Bal, s.In, s.Out)
}

func FormatVal(v int) string {
	if v < 0 {
		return fmt.Sprintf("$\u00A0(%.2f)", float32(-v)/100.0)
	} else {
		return fmt.Sprintf("$\u00A0%.2f", float32(v)/100.0)
	}
}
