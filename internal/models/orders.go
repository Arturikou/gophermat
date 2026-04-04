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

var PendingOrderStatuses = []OrderStatus{
	OrderStatusNew,
	OrderStatusProcessing,
}

type Order struct {
	Number     string
	Status     OrderStatus
	Accrual    decimal.Decimal
	UploadedAt time.Time
}

type OrderUpdate struct {
	Number  string
	UserID  int
	Status  OrderStatus
	Accrual decimal.Decimal
}
