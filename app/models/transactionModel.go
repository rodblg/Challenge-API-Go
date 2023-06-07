package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID            primitive.ObjectID `bson:"_id"`
	Value         *float64           `json:"value" validate:"required"`
	Name_movement *string            `json:"name_movement" validate:"min=3"`
	Created_at    time.Time          `json:"created_at"`
	User_id       string             `json:"user_id" validate:"required"`
}

type ServicePayment struct {
	ID            primitive.ObjectID `bson:"_id"`
	Service       Service            `json:"service_name bson" bson:"service_name"`
	Value         *float64           `json:"value" validate:"required"`
	Name_movement *string            `json:"name_movement"`
	User_id       string             `json:"user_id"`
}
