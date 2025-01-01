package entity

import "gorm.io/gorm"


type User struct {
	gorm.Model
	UserName 		string  `gorm:"name"`
	Age  			int32   `gorm:"age"`
	HashedPassword 	[]byte	`gorm:"hashed_password"`
	Email 			string	`gorm:"email"`
}