package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:string"`
	CreatedAt time.Time
	ProductId uuid.UUID `json:"product_id"`
	Product   Product   `gorm:"foreignKey:ProductId"`
	UserId    uuid.UUID `json:"user_id"`
	User      User      `gorm:"foreignKey:UserId"`
}

func (order *Order) BeforeCreate(tx *gorm.DB) (err error) {
	order.ID = uuid.New()
	return
}
