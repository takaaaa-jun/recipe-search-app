package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"backend/models"
	"github.com/gin-gonic/gin"
)

// PersonalRecipeHandler : パーソナルレシピのレシピデータの構造
type PersonalRecipeHandler struct {
	DB *sql.DB
}

// PersonalRecipes : パーソナルレシピ検索
func (h *PersonalRecipeHandler) PersonalRecipes(c *gin.Context) {
	query := c.Query("query")
	
	var rows *sql.Rows
	var err error

	// h.DB（レシピデータ）から検索
	if query != "" {
		// 検索語あり
		fmt.Println("検索語:", query)
		rows, err = h.DB.Query("SELECT id, title, description, serving_for, published_at FROM recipes WHERE title LIKE ?", "%"+query+"%")
	} else {
		// 検索語なし
		fmt.Println("1件取得")
		rows, err = h.DB.Query("SELECT id, title, description, serving_for, published_at FROM recipes LIMIT 1")
	}

	if err != nil {
		// エラー処理
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ接続失敗:" + err.Error()})
		return
	}
	defer rows.Close()

	recipes := []models.Recipe{}
	for rows.Next() {
		var r models.Recipe
		// データの取得
		if err := rows.Scan(&r.ID, &r.Title, &r.Description, &r.Serving_for, &r.Published_at); err != nil {
			fmt.Println("データ取得失敗:", err)
			log.Println("取得しようとしたデータ:", r)

			c.JSON(http.StatusInternalServerError, gin.H{"error": "データ取得失敗:" + err.Error()})
			return
		}
		recipes = append(recipes, r)
	}
	c.JSON(http.StatusOK, gin.H{"recipes": recipes})
}