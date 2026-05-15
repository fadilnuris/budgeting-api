package models

import "time"

type Transaction struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint
	Type      string // income / expense
	Amount    int
	Note      string
	Date      string
	CreatedAt time.Time
}
