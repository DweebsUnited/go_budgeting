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
	ret := fmt.Sprintf("%d: %s:%s -- ", a.ID, a.Institution, a.Name)
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

func (eg EnvelopeGroup) String() string {
	return fmt.Sprintf("%d: %s", eg.ID, eg.Name)
}

func (e Envelope) String() string {
	ret := fmt.Sprintf("%d.%d %s -- ", e.GroupID, e.ID, e.Name)
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
	if e.DebtAccount > 0 {
		ret += fmt.Sprintf(" -> %d", e.DebtAccount)
	}
	return ret
}
