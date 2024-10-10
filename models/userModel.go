package models

import (
	"gorm.io/gorm"
)

type User struct {
	//used to create  a table in db
	gorm.Model
	Email    string `gorm:"type:varchar(100);unique"`
	Password string `gorm:"size:255"`
	Movies   []Movie
}
