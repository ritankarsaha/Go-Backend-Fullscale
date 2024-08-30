package models

import(

	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID            primitive.ObjectID `bson:"_id"`
	Fullname      *string            `json:"fullname" validate:"required,min=2,max=100"`
	First_name    *string            `json:"first_name" validate:"required,min=2,max=100"`
	Last_name      *string           `json:"last_name" validate:"required,min=2,max=100"`
	Email         *string            `json:"email" validate:"email,required"`
	Avatar        *string            `json:"avatar"`
	Phone         *string            `json:"phone" validate:"required"`
	Role          *string            `json:"user_type" validate:"required,eq=JOURNALIST|eq=CONSUMER|eq=AUDITOR"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
	RefreshToken  *string            `json:"refresh_token" bson:"refresh_token"`
	Preference    *string            `json:"preference"`
    Address		  *string 			 `json:"address,omitempty"`
	Password      *string            `json:"Password" validate:"required,min=6"`
	Token         *string            `json:"token"`
	User_id       *string			 `json:"user_id" bson:"user_id"`
}

