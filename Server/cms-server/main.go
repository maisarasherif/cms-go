package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	controller "github.com/maisarasherif/cms-go/Server/cms-server/controllers"
)

func main() {

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, CMS!")
	})

	router.GET("/personnel", controller.GetPersonnel())

	router.GET("/person/:company_id", controller.GetPerson())

	router.POST("/addpersonnel", controller.AddPersonnel())

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
