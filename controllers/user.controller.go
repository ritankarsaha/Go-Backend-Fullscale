package controllers

import (
	"context"
	"fmt"
	"github.com/ritankarsaha/Go-Backend-Fullscale/database"
	// "github.com/ritankarsaha/Go-Backend-Fullscale/helpers"
	"github.com/ritankarsaha/Go-Backend-Fullscale/models"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

func GetUsers() gin.HandlerFunc{
	return func ( c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(), 100*time.Second)

		recordPerPage,err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1{
			recordPerPage = 10
		}

		page, err1 := strconv.Atoi(c.Query("page"))
		if err1 != nil || page < 1{
			page = 1
		}

		startIndex := (page - 1) * recordPerPage
		startIndex, err = strconv.Atoi(c.Query("startIndex"))


		matchStage := bson.D{
			{Key: "$match", Value: bson.D{}},  
		}
		
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0}, 
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

        result , err := userCollection.Aggregate(ctx, mongo.Pipeline{
			matchStage , projectStage,
		})
		defer cancel()
		if err != nil{
		    c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing user items. "})	
		}
		var allUsers []bson.M
		if err = result.All(ctx, &allUsers); err != nil{
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allUsers[0])


	}
}

func GetUser() gin.HandlerFunc{

	return func(c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(), 100*time.Second)
		userId := c.Param("user_id")
		var user models.User
		err := userCollection.FindOne(ctx,bson.M{"user_id":userId}).Decode(&user)
		defer cancel()
		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error has occured while fetching these items"})

		}
		c.JSON(http.StatusOK,user)
	}
}

func SignUp() gin.HandlerFunc{

	return func (c *gin.Context){
		var ctx,cancel = context.WithTimeout(context.Background(),100*time.Second)
		var user models.User


		//converting the json data being received into data which can be read by golang.
		if err := c.BindJSON(&user); err != nil{
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validating the user model based on the user struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
        
		
	}
}