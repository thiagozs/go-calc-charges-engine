package domain

import "time"

type Money int64

type Transaction struct {
	ID            string
	Amount        Money
	Date          time.Time
	International bool
}
