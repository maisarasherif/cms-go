package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	database "github.com/maisarasherif/cms-go/Server/cms-server/database"
	"github.com/maisarasherif/cms-go/Server/cms-server/models"
	"github.com/tmc/langchaingo/llms/openai"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var personnelCollection *mongo.Collection = database.OpenCollection("Personnel")
var skillCollection *mongo.Collection = database.OpenCollection("Skills")

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

func SummaryUpdate() gin.HandlerFunc {
	return func(c *gin.Context) {
		personnelId := c.Param("company_id")
		if personnelId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "personeel id required"})
			return
		}
		var req struct {
			Summary string `json:"summary"`
		}
		var resp struct {
			SkillName string `json:"skill_name"`
			Summary   string `json:"summary"`
		}

		if err := c.ShouldBind(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		sentiment, skillVal, err := GetSkillRanking(req.Summary)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting skill ranking"})
			return
		}

		filter := bson.M{"company_id": personnelId}

		update := bson.M{
			"$set": bson.M{
				"summary": req.Summary,
				"skills": bson.M{
					"skill_value": skillVal,
					"skill_name":  sentiment,
				},
			},
		}
		var ctx, cancel = context.WithTimeout(c.Request.Context(), 100*time.Second)
		defer cancel()
		result, err := personnelCollection.UpdateOne(ctx, filter, update)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error updating personnel"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "personnel not found"})
			return
		}

		resp.SkillName = sentiment
		resp.Summary = req.Summary

		c.JSON(http.StatusOK, resp)

	}

}

func GetSkillRanking(summary string) (string, int, error) {
	skills, err := GetSkills()

	if err != nil {
		return "", 0, err
	}

	sentimentDelimited := ""

	for _, skill := range skills {
		if skill.SkillValue != 999 {
			sentimentDelimited = sentimentDelimited + skill.SkillName + ","
		}
	}

	sentimentDelimited = strings.Trim(sentimentDelimited, ",")

	err = godotenv.Load(".env")

	if err != nil {
		log.Println("Warning: .env file not found")
	}

	OpenAiApiKey := os.Getenv("OPENAI_API_KEY")

	if OpenAiApiKey == "" {
		return "", 0, errors.New("could not read OPENAI_API_KEY")
	}

	llm, err := openai.New(openai.WithToken(OpenAiApiKey))

	if err != nil {
		return "", 0, err
	}

	base_prompt_template := os.Getenv("BASE_PROMPT_TEMPLATE")

	base_prompt := strings.Replace(base_prompt_template, "{skills}", sentimentDelimited, 1)

	response, err := llm.Call(context.Background(), base_prompt+summary)

	if err != nil {
		return "", 0, err
	}
	skillVal := 0

	for _, skill := range skills {
		if skill.SkillName == response {
			skillVal = skill.SkillValue
			break
		}
	}
	return response, skillVal, nil

}

func GetSkills() ([]models.Skills, error) {
	var skills []models.Skills

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := skillCollection.Find(ctx, bson.M{})

	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &skills); err != nil {
		return nil, err
	}

	return skills, nil
}

func GetRecommendedPersonnel() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			RequiredSkills []string `json:"required_skills" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "required_skills array is required"})
			return
		}

		if len(req.RequiredSkills) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one skill is required"})
			return
		}

		err := godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		// Find all personnel who have ANY of the required skills
		filter := bson.M{"skills.skill_name": bson.M{"$in": req.RequiredSkills}}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := personnelCollection.Find(ctx, filter)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended personnel"})
			return
		}
		defer cursor.Close(ctx)

		var allPersonnel []models.Personnel

		if err := cursor.All(ctx, &allPersonnel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Calculate match scores for each person
		type PersonnelWithScore struct {
			Personnel     models.Personnel `json:"personnel"`
			MatchScore    float64          `json:"match_score"`
			MatchCount    int              `json:"match_count"`
			AvgSkillValue float64          `json:"avg_skill_value"`
		}

		var scoredPersonnel []PersonnelWithScore

		for _, person := range allPersonnel {
			if person.Skills == nil {
				continue
			}

			var totalSkillValue int
			var matchCount int

			// Calculate score based on matching skills
			for _, skill := range *person.Skills {
				for _, requiredSkill := range req.RequiredSkills {
					if skill.SkillName == requiredSkill {
						totalSkillValue += skill.SkillValue
						matchCount++
						break
					}
				}
			}

			if matchCount > 0 {
				avgSkillValue := float64(totalSkillValue) / float64(matchCount)
				// Lower avg skill value = better match
				// More matches = better candidate
				// Score formula: avgSkillValue / matchCount (lower is better)
				score := avgSkillValue / float64(matchCount)

				scoredPersonnel = append(scoredPersonnel, PersonnelWithScore{
					Personnel:     person,
					MatchScore:    score,
					MatchCount:    matchCount,
					AvgSkillValue: avgSkillValue,
				})
			}
		}

		// Sort by score (lower is better)
		sort.Slice(scoredPersonnel, func(i, j int) bool {
			// First compare by match count (more matches is better)
			if scoredPersonnel[i].MatchCount != scoredPersonnel[j].MatchCount {
				return scoredPersonnel[i].MatchCount > scoredPersonnel[j].MatchCount
			}
			// Then by average skill value (lower is better)
			return scoredPersonnel[i].AvgSkillValue < scoredPersonnel[j].AvgSkillValue
		})

		// Get limit from env
		var recommendedPersonnelLimitVal int = 10
		recommendedPersonnelLimitStr := os.Getenv("RECOMMENDED_PERSONNEL_LIMIT")
		if recommendedPersonnelLimitStr != "" {
			if limit, err := strconv.Atoi(recommendedPersonnelLimitStr); err == nil {
				recommendedPersonnelLimitVal = limit
			}
		}

		// Apply limit
		if len(scoredPersonnel) > recommendedPersonnelLimitVal {
			scoredPersonnel = scoredPersonnel[:recommendedPersonnelLimitVal]
		}

		c.JSON(http.StatusOK, scoredPersonnel)
	}
}

/*
func GetRecommendedPersonnel() gin.HandlerFunc {
	return func(c *gin.Context) {
		companyId, err := utils.GetPersonnelID(c)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "personnel id not found"})
			return
		}
		skills, err := GetPersonnelSkills(companyId)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		err = godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found")
		}

		var recommendedPersonnelLimitVal int64 = 5

		recommendedPersonnelLimitStr := os.Getenv("RECOMMENDED_PERSONNEL_LIMIT")

		if recommendedPersonnelLimitStr != "" {
			recommendedPersonnelLimitVal, _ = strconv.ParseInt(recommendedPersonnelLimitStr, 10, 64)
		}

		findOptions := options.Find()

		findOptions.SetSort(bson.D{{Key: "skills.skill_value", Value: 1}})

		findOptions.SetLimit(recommendedPersonnelLimitVal)

		filter := bson.M{"skills.skill_name": bson.M{"$in": skills}}

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		cursor, err := personnelCollection.Find(ctx, filter, findOptions)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching recommended personnel"})
			return
		}
		defer cursor.Close(ctx)

		var recommendedPersonnel []models.Personnel

		if err := cursor.All(ctx, &recommendedPersonnel); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, recommendedPersonnel)

	}
}

func GetPersonnelSkills(CompanyID string) ([]string, error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"company_id": CompanyID}

	projection := bson.M{
		"skills.skill_name": 1,
		"_id":               0,
	}

	opts := options.FindOne().SetProjection(projection)

	var results bson.M

	err := personnelCollection.FindOne(ctx, filter, opts).Decode(&results)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []string{}, nil
		}
	}

	SkillsArray, ok := results["skills"].(bson.A)

	if !ok {
		return []string{}, errors.New("unable to retrieve user data")
	}

	var skillName []string

	for _, item := range SkillsArray {
		if skillMap, ok := item.(bson.D); ok {
			for _, elem := range skillMap {
				if elem.Key == "skill_name" {
					if name, ok := elem.Value.(string); ok {
						skillName = append(skillName, name)
					}
				}
			}
		}
	}

	return skillName, nil
}
*/
