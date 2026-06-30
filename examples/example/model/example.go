package model

import (
	"time"
)

type ExampleModel struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"column:name;unique;size:100;not null"`
	Email     string    `gorm:"column:email;unique;size:255"`
	Status    int       `gorm:"column:status;default:1;index"`
	DeletedAt time.Time `gorm:"column:deleted_at;index"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (ExampleModel) TableName() string {
	return "example_models"
}
