package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/maisarasherif/cms-go/Server/cms-server/controllers"
	"github.com/maisarasherif/cms-go/Server/cms-server/middleware"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupProtectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.Use(middleware.AuthMiddleware())

	router.GET("/person/:company_id", controller.GetPerson(client))
	router.POST("/addpersonnel", controller.AddPersonnel(client))
	router.POST("/recommendedpersonnel", controller.GetRecommendedPersonnel(client))
}
