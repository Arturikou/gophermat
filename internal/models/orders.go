package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	Number     string
	Status     OrderStatus
	Accrual    decimal.Decimal
	UploadedAt time.Time
}
