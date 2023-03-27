package repository

import "gorm.io/gorm"

type FilterFunc func(user *gorm.DB) *gorm.DB
