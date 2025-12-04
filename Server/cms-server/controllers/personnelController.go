package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	database "github.com/maisarasherif/cms-go/Server/cms-server/database"
	"github.com/maisarasherif/cms-go/Server/cms-server/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var personnelCollection *mongo.Collection = database.OpenCollection("Personnel")

var validate = validator.New()

func GetPersonnel() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		var personnel []models.Personnel

		cursor, err := personnelCollection.Find(ctx, bson.M{})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch personnel."})
			return
		}
		defer cursor.Close(ctx)

		if err = cursor.All(ctx, &personnel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode personnel."})
			return

		}

		c.JSON(http.StatusOK, personnel)
	}
}

func GetPerson() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		PersonnelID := c.Param("company_id")

		if PersonnelID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "personnel company id is required"})
			return
		}
		var personnelStruct models.Personnel

		err := personnelCollection.FindOne(ctx, bson.M{"company_id": PersonnelID}).Decode(&personnelStruct)

		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "personnel not found"})
			return
		}

		c.JSON(http.StatusOK, personnelStruct)
	}
}

func AddPersonnel() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()

		var personnel models.Personnel
		if err := c.ShouldBindJSON(&personnel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
			return
		}
		if err := validate.Struct(personnel); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed!", "details": err.Error()})
			return
		}

		result, err := personnelCollection.InsertOne(ctx, personnel)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add personnel"})
			return
		}

		c.JSON(http.StatusCreated, result)

	}
}
