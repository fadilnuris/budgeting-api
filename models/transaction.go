package models

type Transaction struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint
	Type   string // income / expense
	Amount int
	Note   string
}
