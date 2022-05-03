package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pianisimo/ecommerce/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"time"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Query("id")
		if userId == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid search index"})
			c.Abort()
			return
		}

		address, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err.Error())
		}

		var addresses models.Address
		addresses.ID = primitive.NewObjectID()

		err = c.BindJSON(&addresses)
		if err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		filterMatch := bson.D{
			{Key: "$match", Value: bson.D{
				primitive.E{Key: "_id", Value: address},
			}},
		}

		unwind := bson.D{
			{Key: "$unwind", Value: bson.D{
				primitive.E{Key: "path", Value: "$address"},
			}},
		}

		grouping := bson.D{
			{Key: "$group", Value: bson.D{
				primitive.E{Key: "_id", Value: "$_id"},
				{Key: "count", Value: bson.D{
					primitive.E{Key: "$sum", Value: 1},
				}},
			}},
		}

		aggregateCursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{
			filterMatch, unwind, grouping,
		})
		if err != nil {
			c.IndentedJSON(http.StatusNotAcceptable, err.Error())
		}

		var addressInfo []bson.M
		err = aggregateCursor.All(ctx, &addresses)
		if err != nil {
			panic(err.(any))
		}

		var size int32
		for _, addressNo := range addressInfo {
			count := addressNo["count"]
			size = count.(int32)
		}

		if size < 2 {
			filter := bson.D{
				primitive.E{
					Key: "_id", Value: address,
				},
			}
			update := bson.D{
				primitive.E{
					Key: "$push", Value: bson.D{
						primitive.E{Key: "address", Value: addresses},
					},
				},
			}
			_, err = UserCollection.UpdateOne(ctx, filter, update)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			c.IndentedJSON(http.StatusBadRequest, "Not allowed")
		}
		defer cancelFunc()
		ctx.Done()
	}
}

func EditHomeAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Query("id")
		if userId == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid"})
			c.Abort()
			return
		}

		hexId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		var editAddress models.Address

		err = c.BindJSON(&editAddress)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		filter := bson.D{primitive.E{Key: "_id", Value: hexId}}

		update := bson.D{{Key: "$set", Value: bson.D{
			primitive.E{Key: "address.0.house_name", Value: editAddress.House},
			primitive.E{Key: "address.0.street_name", Value: editAddress.Street},
			primitive.E{Key: "address.0.city_name", Value: editAddress.City},
			primitive.E{Key: "address.0.pin_code", Value: editAddress.Pincode},
		}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancelFunc()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully updated address")
	}
}

func EditWorkAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Query("id")
		if userId == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid"})
			c.Abort()
			return
		}

		hexId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		var editAddress models.Address

		err = c.BindJSON(&editAddress)
		if err != nil {
			c.IndentedJSON(http.StatusBadRequest, err.Error())
			return
		}

		filter := bson.D{primitive.E{Key: "_id", Value: hexId}}

		update := bson.D{{Key: "$set", Value: bson.D{
			primitive.E{Key: "address.1.house_name", Value: editAddress.House},
			primitive.E{Key: "address.1.street_name", Value: editAddress.Street},
			primitive.E{Key: "address.1.city_name", Value: editAddress.City},
			primitive.E{Key: "address.1.pin_code", Value: editAddress.Pincode},
		}}}

		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}
		defer cancelFunc()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully updated work address")
	}
}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		userId := c.Query("id")
		if userId == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "invalid search index"})
			c.Abort()
			return
		}

		addresses := make([]models.Address, 0)
		hexId, err := primitive.ObjectIDFromHex(userId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancelFunc()

		filter := bson.D{primitive.E{Key: "_id", Value: hexId}}
		update := bson.D{
			{Key: "$set", Value: bson.D{
				primitive.E{Key: "address", Value: addresses},
			}},
		}
		_, err = UserCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "Wrong backend db command")
			return
		}
		defer cancelFunc()
		ctx.Done()
		c.IndentedJSON(http.StatusOK, "Successfully Deleted")
	}
}
