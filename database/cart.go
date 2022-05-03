package database

import (
	"context"
	"errors"
	"github.com/pianisimo/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

var (
	ErrCantFindProduct        = errors.New("can not find product")
	ErrCantDecodeProducts     = errors.New("can not decode product")
	ErrUserIdNotValid         = errors.New("this user is not valid")
	ErrCantUpdateUser         = errors.New("can not add this product to the cart")
	ErrCantRemoveItemFromCart = errors.New("can not remove this item from the cart")
	ErrCantGetItem            = errors.New("unable to get the item from the cart")
	ErrCantBuyCartItem        = errors.New("can not update the purchase")
)

func AddProductToCart(ctx context.Context, productCollection, userCollection *mongo.Collection, productId primitive.ObjectID, userId string) error {
	searchFromdb, err := productCollection.Find(ctx, bson.M{"_id": productId})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}

	var productCart []models.Product
	err = searchFromdb.All(ctx, &productCart)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	id, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdNotValid
	}

	filter := bson.D{primitive.E{
		Key:   "_id",
		Value: id,
	}}
	update := bson.D{{
		Key: "$push",
		Value: bson.D{primitive.E{
			Key: "user_cart",
			Value: bson.D{{
				Key:   "$each",
				Value: productCart,
			}},
		}},
	}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}

	return nil
}

func RemoveCartItem(ctx context.Context, userCollection *mongo.Collection, productId primitive.ObjectID, userId string) (int32, uint64, error) {
	objectIDFromHex, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return 0, 0, ErrUserIdNotValid
	}

	unwind := bson.D{{
		Key: "$unwind",
		Value: bson.D{primitive.E{
			Key:   "path",
			Value: "$user_cart",
		}},
	}}

	grouping := bson.D{{
		Key: "$group",
		Value: bson.D{primitive.E{
			Key:   "_id",
			Value: "$_id",
		},
			{
				Key: "total",
				Value: bson.D{primitive.E{
					Key:   "$sum",
					Value: 1,
				}},
			},
			{
				Key: "price",
				Value: bson.D{primitive.E{
					Key:   "$sum",
					Value: "$user_cart.price",
				}},
			},
		},
	}}

	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})

	var cartItems []bson.M
	err = currentResults.All(ctx, &cartItems)
	if err != nil {
		panic(err.(any))
	}
	var count int32
	var price uint64

	for _, item := range cartItems {
		total := item["total"]
		count = total.(int32)

		priceTotal := item["price"]
		log.Println(priceTotal)
		price = uint64(priceTotal.(int64))
	}

	filter := bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}
	update := bson.D{{
		Key: "$pull",
		Value: bson.M{
			"user_cart": bson.M{"_id": productId},
		},
	}}

	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		return 0, 0, err
	}
	if err != nil {
		log.Println(err)
		return 0, 0, ErrCantUpdateUser
	}

	return count, price, nil
}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userId string) error {
	objectIDFromHex, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdNotValid
	}

	var getCartItems models.User
	var orderCart models.Order

	orderCart.ID = primitive.NewObjectID()
	orderCart.OrderedAt = time.Now()
	orderCart.Cart = make([]models.Product, 0)
	orderCart.PaymentMethod = models.Payment{
		Digital: false,
		COD:     true,
	}

	unwind := bson.D{{
		Key: "$unwind",
		Value: bson.D{primitive.E{
			Key:   "path",
			Value: "$user_cart",
		}},
	}}

	grouping := bson.D{{
		Key: "$group",
		Value: bson.D{primitive.E{
			Key:   "_id",
			Value: "$_id",
		},
			{
				Key: "total",
				Value: bson.D{primitive.E{
					Key:   "$sum",
					Value: "$user_cart.price",
				}},
			},
		},
	}}

	currentResults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		log.Println(err)
		panic(err.(any))
		return ErrCantBuyCartItem
	}

	var getUserCart []bson.M
	err = currentResults.All(ctx, &getUserCart)
	if err != nil {
		panic(err.(any))
	}
	var totalPrice uint64
	for _, item := range getUserCart {
		price := item["total"]
		totalPrice = price.(uint64)
	}

	orderCart.Price = &totalPrice

	filter := bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}
	update := bson.D{{
		Key: "$push",
		Value: bson.D{primitive.E{
			Key:   "orders",
			Value: orderCart,
		}},
	}}

	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
		panic(err.(any))
		return ErrCantBuyCartItem
	}

	err = userCollection.FindOne(ctx, bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}).Decode(&getCartItems)
	if err != nil {
		log.Println(err)
		panic(err.(any))
		return ErrCantDecodeProducts
	}

	filter = bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}
	update = bson.D{primitive.E{
		Key:   "$push",
		Value: bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}},
	}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		panic(err.(any))
		return ErrCantBuyCartItem
	}

	userCartEmpty := make([]models.Product, 0)

	filter = bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}
	update = bson.D{primitive.E{
		Key: "$push",
		Value: bson.D{{
			Key: "$set",
			Value: bson.D{primitive.E{
				Key:   "user_cart",
				Value: userCartEmpty,
			}},
		}},
	}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		panic(err.(any))
		return ErrCantBuyCartItem
	}

	return nil
}

func InstantBuyer(ctx context.Context, productCollection, userCollection *mongo.Collection, productId primitive.ObjectID, userId string) error {
	objectIDFromHex, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println(err)
		return ErrUserIdNotValid
	}

	var productDetails models.Product
	var orderDetail models.Order

	orderDetail.ID = primitive.NewObjectID()
	orderDetail.OrderedAt = time.Now()
	orderDetail.Cart = make([]models.Product, 0)
	orderDetail.PaymentMethod = models.Payment{
		Digital: false,
		COD:     true,
	}

	err = productCollection.FindOne(ctx, bson.D{
		primitive.E{
			Key:   "_id",
			Value: productId,
		},
	}).Decode(&productDetails)
	if err != nil {
		log.Println(err)
		return ErrCantDecodeProducts
	}

	orderDetail.Price = productDetails.Price

	filter := bson.D{primitive.E{
		Key:   "_id",
		Value: objectIDFromHex,
	}}
	update := bson.D{{
		Key: "$push",
		Value: bson.D{primitive.E{
			Key:   "orders",
			Value: orderDetail,
		}},
	}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantBuyCartItem
	}

	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": productDetails}}

	_, err = userCollection.UpdateOne(ctx, filter, update2)
	if err != nil {
		log.Println(err)
		return ErrCantBuyCartItem
	}

	return nil
}
