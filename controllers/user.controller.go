package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/ritankarsaha/Go-Backend-Fullscale/database"
	helper "github.com/ritankarsaha/Go-Backend-Fullscale/helpers"
	"github.com/ritankarsaha/Go-Backend-Fullscale/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")



// obtaining the user data through an id from the database
func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		if err := helper.MathUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var user models.User
		objID, _ := primitive.ObjectIDFromHex(userId)
		err := userCollection.FindOne(c, bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, gin.H{
					"error": "User not found",
				})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}

// resgistering the user in the database.
func RegisterUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		var newUser struct {

			Fullname string `json:"fullname" binding:"required,min=2,max=100"`
			Email    string `json:"email" binding:"required,email"`
			Avatar   string `json:"avatar" binding:"required"`
			Username string `json:"username" binding:"required,min=3,max=100"`
		}

		if err := ctx.ShouldBindJSON(&newUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid input data",
			})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var existingUser models.User
		err := userCollection.FindOne(c, bson.M{"email": newUser.Email}).Decode(&existingUser)
		if err == nil {
			ctx.JSON(http.StatusConflict, gin.H{
				"error": "User with similar email id is already present. Please login to continue.",
			})
			return
		} else if err != mongo.ErrNoDocuments {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		user := models.User{
			ID:        primitive.NewObjectID(),
			Fullname:  &newUser.Fullname,
			Email:     &newUser.Email,
			Avatar:    &newUser.Avatar,
			Phone:     nil,
			Role:      nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err = userCollection.InsertOne(c, user)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": "User can't be registered in the database successfully! ",
			})
			return
		}

		ctx.JSON(http.StatusCreated, gin.H{
			"message": "User has been registered successfully",
			"user_id": user.ID.Hex(),
		})
	}
}



// updating the existing user in the database.
func UpdateUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		if err := helper.MathUserTypeToUid(ctx, userId); 
		err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		var updateData bson.M
		if err := ctx.BindJSON(&updateData); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid input",
			})
			return
		}

		userRole, _ := ctx.Get("user_role") 
		if _, exists := updateData["address"]; exists && userRole != "MANAGER" {
			ctx.JSON(http.StatusForbidden, gin.H{
				"error": "User has been unauthorized to change/modify.",
			})
			return
		}

		delete(updateData, "_id")
		delete(updateData, "email") 
		delete(updateData, "role")  

		if len(updateData) == 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Please enter a valid field to update the user data",
			})
			return
		}

		updateData["updated_at"] = time.Now()

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		objID, _ := primitive.ObjectIDFromHex(userId)
		filter := bson.M{"_id": objID}

		update := bson.M{
			"$set": updateData,
		}

		result, err := userCollection.UpdateOne(c, filter, update)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if result.MatchedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User not found or doesn't exist.",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "User data has been updated successfully!",
		})
	}
}


// deleting the user from the database.
func DeleteUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")
		if err := helper.MathUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		objID, _ := primitive.ObjectIDFromHex(userId)
		filter := bson.M{"_id": objID}
		result, err := userCollection.DeleteOne(c, filter)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		if result.DeletedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User doesn't exist in the database.",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "User has been removed successfully! ",
		})
	}
}



// listing all the users in the database.
func ListUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		cursor, err := userCollection.Find(c, bson.M{})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer cursor.Close(c)

		var users []models.User
		if err = cursor.All(c, &users); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"users": users,
		})
	}
}




// changing the user role 
func ChangeUserRole() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		var roleChange struct {
			Role string `json:"role" binding:"required,eq=JOURNALIST|eq=CONSUMER|eq=AUDITOR"`
		}

		if err := ctx.ShouldBindJSON(&roleChange); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid input",
			})
			return
		}

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		objID, _ := primitive.ObjectIDFromHex(userId)
		filter := bson.M{"_id": objID}

		update := bson.M{
			"$set": bson.M{
				"role":       roleChange.Role,
				"updated_at": time.Now(),
			},
		}

		result, err := userCollection.UpdateOne(c, filter, update)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		if result.MatchedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{
				"error": "User not found or doesn't exist.",
			})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"message": "User role updated successfully!",
		})
	}
}



//getting the entire user profile from the database
func GetUserProfile() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var user models.User
		objID, _ := primitive.ObjectIDFromHex(userId)
		err := userCollection.FindOne(c, bson.M{"_id": objID}).Decode(&user)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				ctx.JSON(http.StatusNotFound, gin.H{
					"error": "User not found",
				})
			} else {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
			}
			return
		}

		ctx.JSON(http.StatusOK, gin.H{
			"user": user,
		})
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
		//you'll check if the email has already been used by another user
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

		//you'll also check if the phone no. has already been used by another user

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

		user.CreatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.UpdatedAt, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		user.ID = primitive.NewObjectID()
		// user.User_id = user.User_id.Hex()
		//generate token and refersh token (generate all tokens function from helper)
		token, refreshToken, _ := helper.GenerateAllTokens(*user.Email, *user.First_name, *user.Last_name, *user.User_id)
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

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		var foundUser models.User

		//convert the login data from postman which is in JSON to golang readable format

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		//find a user with that email and see if that user even exists

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found, login seems to be incorrect"})
			return
		}

		//then you will verify the password

		passwordIsValid, msg := VerifyPassword(*user.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		//if all goes well, then you'll generate tokens

		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.First_name, *foundUser.Last_name, *foundUser.User_id)

		//update tokens - token and refersh token
		helper.UpdateAllTokens(token, refreshToken, *foundUser.User_id)

		//return statusOK
		c.JSON(http.StatusOK, foundUser)
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