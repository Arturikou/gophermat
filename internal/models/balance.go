package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type Withdrawal struct {
	ID          int
	UserID      int
	OrderNumber string
	Amount      decimal.Decimal
	ProcessedAt time.Time
}

type Balance struct {
	Current   decimal.Decimal
	Withdrawn decimal.Decimal
}
