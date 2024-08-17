package main

import(
	"os"
	"github.com/ritankarsaha/Go-Backend-Fullscale/database"
	routes "github.com/ritankarsaha/Go-Backend-Fullscale/routes"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)


var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
func main(){


	// setting up the port on whoch the server is to be exposed.
	port := os.Getenv("PORT")

	if port == ""{
		port = "8000"
	}


	//setting up the gin router
	router := gin.New()
	router.Use(gin.Logger())



	//defining all the routes - pre-planned
	routes.UserRoutes(router)
	routes.FoodRoutes(router)
	routes.MenuRoutes(router)
	routes.TableRoutes(router)
	routes.OrderRoutes(router)
	routes.OrderItemRoutes(router)
	routes.InvoiceRoutes(router)

	router.Run(":" + port)

}