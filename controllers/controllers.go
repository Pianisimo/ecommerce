package controllers

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/pianisimo/ecommerce/database"
	"github.com/pianisimo/ecommerce/models"
	"github.com/pianisimo/ecommerce/tokens"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"time"
)

var (
	UserCollection    = database.UserData(database.Client, "Users")
	ProductCollection = database.ProductData(database.Client, "Products")
	Validate          = validator.New()
)

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		var user models.User
		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err = Validate.Struct(user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		documents, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if documents > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		documents, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if documents > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user already exists"})
			return
		}

		password := HashPassword(*user.Password)
		user.Password = &password

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.UserId = user.ID.Hex()

		token, refreshToken, _ := tokens.TokenGenerator(*user.Email, *user.FirstName, *user.LastName, user.UserId)

		user.Token = &token
		user.RefreshToken = &refreshToken
		user.UserCart = make([]models.Product, 0)
		user.AddressDetails = make([]models.Address, 0)
		user.OrderStatus = make([]models.Order, 0)

		_, err = UserCollection.InsertOne(ctx, user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "the user did not get created"})
			return
		}
		defer cancelFunc()

		c.JSON(http.StatusCreated, "successfully signed in")
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		var user models.User
		err := c.BindJSON(&user)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var foundUser models.User
		UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancelFunc()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "login or password incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancelFunc()

		if !passwordIsValid {
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}

		token, refreshToken, _ := tokens.TokenGenerator(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, foundUser.UserId)
		defer cancelFunc()
		tokens.UpdateAllTokens(token, refreshToken, foundUser.UserId)

		c.JSON(http.StatusFound, foundUser)
	}
}

func ProductViewerAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		var products models.Product
		defer cancelFunc()
		err := c.BindJSON(&products)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		products.ID = primitive.NewObjectID()
		_, err = ProductCollection.InsertOne(ctx, products)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not inserted"})
			return
		}

		c.JSON(http.StatusOK, "Successfully added")
	}
}
func SearchProduct() gin.HandlerFunc {
	return func(c *gin.Context) {
		var productList []models.Product
		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		cursor, err := ProductCollection.Find(ctx, bson.D{{}})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "something went wrong")
			return
		}

		err = cursor.All(ctx, &productList)
		if err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		defer cursor.Close(ctx)
		err = cursor.Err()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}
		defer cancelFunc()

		c.IndentedJSON(http.StatusOK, productList)
	}
}
func SearchProductByQuery() gin.HandlerFunc {
	return func(c *gin.Context) {
		var searchedProducts []models.Product
		queryParam := c.Query("name")
		if queryParam == "" {
			log.Println("query is empty")
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"Error": "Invalid search index"})
			return
		}

		ctx, cancelFunc := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelFunc()

		searchQueryDb, err := ProductCollection.Find(ctx, bson.M{
			"name": bson.M{"$regex": queryParam},
		})
		if err != nil {
			c.IndentedJSON(http.StatusNotFound, "something went wrong while fetching the data")
			return
		}

		err = searchQueryDb.All(ctx, &searchedProducts)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid")
			return
		}
		defer searchQueryDb.Close(ctx)
		err = searchQueryDb.Err()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusBadRequest, "invalid request")
			return
		}
		defer cancelFunc()

		c.IndentedJSON(http.StatusOK, searchedProducts)
	}
}

func HashPassword(password string) (msg string) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panicln(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword, givenPassword string) (passwordIsValid bool, msg string) {
	err := bcrypt.CompareHashAndPassword([]byte(givenPassword), []byte(userPassword))
	passwordIsValid = true
	msg = ""

	if err != nil {
		msg = "Login or password is incorrect"
		passwordIsValid = false
	}

	return passwordIsValid, msg
}
