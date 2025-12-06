package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	routes "github.com/maisarasherif/cms-go/Server/cms-server/routes"
)

func main() {

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, CMS!")
	})

	routes.SetupUnprotectedRoutes(router)
	routes.SetupProtectedRoutes(router)

	if err := router.Run(":8080"); err != nil {
		fmt.Println("Failed to start server", err)
	}
}
