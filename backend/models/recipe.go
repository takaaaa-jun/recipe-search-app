package models

import (
	"time"
)

// Recipe : レシピデータベース内のレシピデータの構造
type Recipe struct {
	ID int `json:"id"`
	Title string `json:"title"`
	Description string `json:"description"`
	CookingTime int `json:"cooking_time"`
	Serving_for string `json:"serving_for"`
	Published_at time.Time `json:"published_at"`
}