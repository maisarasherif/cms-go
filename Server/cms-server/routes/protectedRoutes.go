package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/maisarasherif/cms-go/Server/cms-server/controllers"
	"github.com/maisarasherif/cms-go/Server/cms-server/middleware"
)

func SetupProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())

	router.GET("/person/:company_id", controller.GetPerson())
	router.POST("/addpersonnel", controller.AddPersonnel())
}
