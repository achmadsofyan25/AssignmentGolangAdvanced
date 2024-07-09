package model

type Transaction struct {
	ID     uint    `gorm:"primaryKey"`
	UserID uint    `gorm:"not null"`
	Type   string  `gorm:"size:50;not null"`
	Amount float64 `gorm:"type:numeric(10,2);not null"`
}
