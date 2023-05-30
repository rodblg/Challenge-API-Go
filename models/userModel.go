package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	First_name    *string            `json:"first_name" validate:"required,min=2,max=100"`
	Email         *string            `json:"email" validate:"email,required"`
	Password      *string            `json:"password" validate:"required,min=6"`
	Phone         *string            `json:"phone" validate:"required"`
	Token         *string            `json:"token"`
	Refresh_token *string            `json:"refresh_token"`
	Balance       *float64           `json:"balance"`
	Created_at    time.Time          `json:"created_at"`
	Updated_at    time.Time          `json:"updated_at"`
	User_id       string             `json:"user_id"`
	Movements     []Transaction      `json:"movements" bson:"movements"`
}

type BankStatement struct {
	ID              primitive.ObjectID `bson:"_id"`
	Transaction_ind []Transaction      `json:"transaction_ind" bson:"transaction_ind"`
	User_id         string             `json:"user_id"`
}

type Service struct {
	ID              primitive.ObjectID `bson:"_id"`
	Service_name    *string            `json:"service_name" validate:"required,min=3"`
	Contract_number *string            `json:"contract_number" validate:"required,min=7"`
}
