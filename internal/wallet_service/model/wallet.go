package model

type Wallet struct {
	ID      uint    `gorm:"primaryKey"`
	UserID  uint    `gorm:"not null"`
	Balance float64 `gorm:"type:numeric(10,2);default:0"`
}
