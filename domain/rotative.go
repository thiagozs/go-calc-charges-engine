package domain

import "time"

type RotativeBalance struct {
    Principal Money
    StartDate time.Time
}
