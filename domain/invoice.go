package domain

import "time"

type Invoice struct {
    ID          string
    ClosingDate time.Time
    DueDate     time.Time
    TotalAmount Money
    PaidAmount  Money
}
