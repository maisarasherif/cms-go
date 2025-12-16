package routes

import (
	"github.com/gin-gonic/gin"
	controller "github.com/maisarasherif/cms-go/Server/cms-server/controllers"
)

func SetupUnprotectedRoutes(router *gin.Engine) {

	router.POST("/register", controller.RegisterUser())
	router.POST("/login", controller.LoginUser())
	router.GET("/personnel", controller.GetPersonnel())
	router.PATCH("/person/:company_id", controller.SummaryUpdate())
}
