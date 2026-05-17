package models

import "time"

type BudgetPlan struct {
	ID         uint             `gorm:"primaryKey" json:"id"`
	UserID     uint             `gorm:"uniqueIndex:idx_user_month" json:"user_id"`
	Month      string           `gorm:"uniqueIndex:idx_user_month" json:"month"` // e.g., "2026-05"
	Income     int              `json:"income"`
	Items      []BudgetItem     `gorm:"foreignKey:PlanID" json:"items"`
	Categories []BudgetCategory `gorm:"foreignKey:PlanID" json:"categories"`
	CreatedAt  time.Time        `json:"created_at"`
}

type BudgetItem struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PlanID    uint      `json:"plan_id"`
	Name      string    `json:"name"`
	Nominal   int       `json:"nominal"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

type BudgetCategory struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	PlanID uint   `json:"plan_id"`
	Name   string `json:"name" gorm:"not null"`
	Color  string `json:"color" gorm:"default:'blue'"` // 'blue', 'amber', 'pink', 'purple', 'emerald', 'teal'
}
