package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/maisarasherif/cms-go/Server/cms-server/controllers"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupUnprotectedRoutes(router *gin.Engine, client *mongo.Client) {

	router.POST("/register", controller.RegisterUser(client))
	router.POST("/login", controller.LoginUser(client))
	router.GET("/personnel", controller.GetPersonnel(client))
}
