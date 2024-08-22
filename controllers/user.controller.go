package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"github.com/ritankarsaha/Go-Backend-Fullscale/database"
	helper "github.com/ritankarsaha/Go-Backend-Fullscale/helpers"
	"github.com/ritankarsaha/Go-Backend-Fullscale/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

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

func SignUp() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User



		//convert the JSON data coming from postman to something that golang understands
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//validate the data based on user struct
		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}


		// check if the email has already been used by another user
		count, err := userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the email"})
			return
		}
		//hash password

		password := HashPassword(*user.Password)
		user.Password = &password

		// check if the phone no. has already been used by another user
		count, err = userCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while checking for the phone number"})
			return
		}

		if count > 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "this email or phone number already exsits"})
			return
		}

		//create some extra details for the user object - created_at, updated_at, ID

		user.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()

		//generate token and refersh token (generate all tokens function from helper)

		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, user.User_id)
		user.Token = &token
		user.RefreshToken = &refreshToken
		//if all ok, then you insert this new user into the user collection

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			msg := fmt.Sprintf("User item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		//return status OK and send the result back

		c.JSON(http.StatusOK, resultInsertionNumber)
	}
}


func Login() gin.HandlerFunc{
	return func ( c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100 *time.Second)
		var user models.User
		var founduser models.User

		// converting data from postman to golang readable data
		if err:= c.BindJSON(&user); err!=nil{
			c.JSON(http.StatusBadRequest, gin.H{"error":err.Error()})
			return
		}

		// finding the user with the email to see if the user even exists
		err := userCollection.FindOne(ctx, bson.M{"email":user.Email}).Decode(&founduser)
		defer cancel()

		if err != nil{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "An error has occured while finding the user in the databse!"})
			return
		}

		// verifying the password
		passwordIsValid, msg := VerifyPassword(*user.Password, *founduser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}


		//updating the tokens for the user as it is authenticicated.
		token, refreshToken, _ := helper.GenerateAllTokens(*founduser.Email, *founduser.First_name, *founduser.Last_name, *&founduser.User_id)
		helper.UpdateAllTokens(token, refreshToken, founduser.User_id)
		c.JSON(http.StatusOK, founduser)

	}
}

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}

	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {

	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("login or password is incorrect")
		check = false
	}
	return check, msg
}