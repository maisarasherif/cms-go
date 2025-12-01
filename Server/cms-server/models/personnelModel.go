package models

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Skills struct {
	SkillID    int    `bson:"skill_id" json:"skill_id" validate:"required"`
	SkillName  string `bson:"skill_name" json:"skill_name" validate:"required,min=1,max=50"`
	SkillLevel string `bson:"skill_level" json:"skill_level" validate:"omitempty,min=1,max=50"`
}

type Personnel struct {
	ID           bson.ObjectID `bson:"_id" json:"_id"`
	Name         string        `bson:"name" json:"name" validate:"required,min=2,max=100"`
	Position     string        `bson:"position" json:"position" validate:"required,min=2,max=100"`
	Image        *string       `bson:"image" json:"image" validate:"omitempty,url"`
	T_BOSIET_EXP *time.Time    `bson:"t_bosiet_exp" json:"t_bosiet_exp" validate:"omitempty,gte=now"`
	Skills       *[]Skills     `bson:"skills" json:"skills" validate:"omitempty,dive"`
	Summary      *string       `bson:"summary" json:"summary" validate:"omitempty"`
}
