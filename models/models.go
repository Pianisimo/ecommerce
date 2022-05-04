package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID             primitive.ObjectID `json:"_id" bson:"_id"`
	FirstName      *string            `json:"first_name" validate:"required,min=2,max=30"`
	LastName       *string            `json:"last_name" validate:"required,min=2,max=30"`
	Password       *string            `json:"password" validate:"required,min=6,max=30"`
	Email          *string            `json:"email" validate:"required,email"`
	Phone          *string            `json:"phone" validate:"required"`
	Token          *string            `json:"token"`
	RefreshToken   *string            `json:"refresh_token" bson:"refresh_token"`
	CreatedAt      time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at" bson:"updated_at"`
	UserId         string             `json:"user_id" bson:"user_id"`
	UserCart       []Product          `json:"user_cart" bson:"user_cart"`
	AddressDetails []Address          `json:"address" bson:"address"`
	OrderStatus    []Order            `json:"orders" bson:"orders"`
}

type Product struct {
	ID     primitive.ObjectID `bson:"_id"`
	Name   *string            `json:"name" bson:"name"`
	Price  *uint64            `json:"price" bson:"price"`
	Rating *uint8             `json:"rating" bson:"rating"`
	Image  *string            `json:"image" bson:"image"`
}

type Address struct {
	ID      primitive.ObjectID `json:"_id" bson:"_id"`
	House   *string            `json:"house" bson:"house"`
	Street  *string            `json:"street" bson:"street"`
	City    *string            `json:"city" bson:"city"`
	Pincode *string            `json:"pincode" bson:"pincode"`
}

type Order struct {
	ID            primitive.ObjectID `json:"_id" bson:"_id"`
	Cart          []Product          `json:"order_list" bson:"order_list"`
	OrderedAt     time.Time          `json:"order_at" bson:"order_at"`
	Price         *uint64            `json:"total_price" bson:"total_price"`
	Discount      *int               `json:"discount" bson:"discount"`
	PaymentMethod Payment            `json:"payment_method" bson:"payment_method"`
}

type Payment struct {
	Digital bool
	COD     bool
}
