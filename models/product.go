package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:string"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string `json:"name"`
	Price     string `json:"price"`
	Quantity  string `json:"quantity"`
}

func (product *Product) BeforeCreate(tx *gorm.DB) (err error) {
	product.ID = uuid.New()
	return
}
