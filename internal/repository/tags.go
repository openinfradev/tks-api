package repository

import (
	"gorm.io/gorm"
)

type Tag struct {
	gorm.Model
	Id          string `gorm:"primaryKey"`
	ReferenceId string
	Key         string
	Value       string
}
