package model

import (
	"time"
)

type ExampleModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
