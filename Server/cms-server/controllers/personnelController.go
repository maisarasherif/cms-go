package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	databases "github.com/maisarasherif/cms-go/Server/cms-server/database"
	"github.com/maisarasherif/cms-go/Server/cms-server/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var personnelCollection *mongo.Collection = databases.OpenCollection("Personnel")

func GetPersonnel() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		var personnel []models.Personnel

		cursor, err := personnelCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch personnel."})
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &personnel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode personnel."})

		}

		c.JSON(http.StatusOK, personnel)
	}
}
