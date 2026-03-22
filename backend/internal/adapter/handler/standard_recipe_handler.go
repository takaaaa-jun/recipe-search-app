// Package handler はHTTPリクエストの受け取りとレスポンス生成を担うアダプター層のパッケージです。
package handler

import (
	"log"
	"net/http"
	"strconv"

	"backend/internal/domain/entity"
	"backend/internal/usecase/standard_recipe"

	"github.com/gin-gonic/gin"
)

// StandardRecipeHandler は基準レシピ関連のHTTPハンドラーを保持する構造体です。
type StandardRecipeHandler struct {
	// searchUsecase は基準レシピ検索ユースケースです
	searchUsecase *standard_recipe.SearchStandardRecipeUsecase
	// detailUsecase は基準レシピ詳細取得ユースケースです
	detailUsecase *standard_recipe.GetStandardRecipeDetailUsecase
}

// NewStandardRecipeHandler はStandardRecipeHandlerの新しいインスタンスを生成します。
func NewStandardRecipeHandler(
	searchUsecase *standard_recipe.SearchStandardRecipeUsecase,
	detailUsecase *standard_recipe.GetStandardRecipeDetailUsecase,
) *StandardRecipeHandler {
	return &StandardRecipeHandler{
		searchUsecase: searchUsecase,
		detailUsecase: detailUsecase,
	}
}

// ===========================================================================
// リクエスト・レスポンス用 DTO
// ===========================================================================

// StandardRecipeSearchRequest は基準レシピ検索APIのリクエストボディ構造体です。
type StandardRecipeSearchRequest struct {
	// Query は検索クエリ文字列
	Query string `json:"query" binding:"required"`
	// SearchMode は検索モード: "recipe"（レシピ名）または "ingredient"（材料名）
	SearchMode string `json:"search_mode"`
}

// StandardIngredientItemResponse は材料アイテムのJSONレスポンス構造体です。
type StandardIngredientItemResponse struct {
	Name  string `json:"name"`  // 材料名
	Count int    `json:"count"` // 使用レシピ件数
}

// StandardIngredientGroupResponse は材料グループのJSONレスポンス構造体です。
type StandardIngredientGroupResponse struct {
	GroupName  string                          `json:"group_name"`  // グループ名（例: "野菜"）
	TotalCount int                             `json:"total_count"` // グループ全体の件数
	Items      []StandardIngredientItemResponse `json:"items"`       // メンバー材料リスト
}

// StandardStepResponse は基準レシピ1手順のJSONレスポンス構造体です。
type StandardStepResponse struct {
	FoodName string `json:"food_name"` // 対象食材名
	Action   string `json:"action"`    // 調理アクション
	Count    int    `json:"count"`     // 使用レシピ件数
}

// StandardRecipeResponse は基準レシピ1件のJSONレスポンス構造体です。
type StandardRecipeResponse struct {
	ID               int                              `json:"id"`                // 基準レシピID
	Name             string                           `json:"name"`              // カテゴリ名
	RecipeCount      int                              `json:"recipe_count"`      // 個人レシピ件数
	CookingTimeLabel string                           `json:"cooking_time_label"` // 調理時間テキスト
	AverageSteps     int                              `json:"average_steps"`     // 平均手順数
	IngredientGroups []StandardIngredientGroupResponse `json:"ingredient_groups"` // 材料グループリスト
	Steps            []StandardStepResponse           `json:"steps"`             // 手順リスト
}

// SearchStandardRecipesResponse は基準レシピ検索APIのレスポンス構造体です。
type SearchStandardRecipesResponse struct {
	Recipes []StandardRecipeResponse `json:"recipes"` // 検索結果一覧（最大5件）
}

// ===========================================================================
// ハンドラーメソッド
// ===========================================================================

// SearchStandardRecipes は基準レシピ検索APIのハンドラーです。
//
// エンドポイント: POST /api/standard-recipes/search
// リクエストボディ (JSON):
//
//	{ "query": "野菜炒め", "search_mode": "recipe" }
func (h *StandardRecipeHandler) SearchStandardRecipes(c *gin.Context) {
	var req StandardRecipeSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "リクエストが不正です: " + err.Error()})
		return
	}

	// search_modeのデフォルト値を設定
	mode := standard_recipe.SearchModeRecipe
	if req.SearchMode == "ingredient" {
		mode = standard_recipe.SearchModeIngredient
	}

	output, err := h.searchUsecase.Execute(c.Request.Context(), standard_recipe.SearchInput{
		Query: req.Query,
		Mode:  mode,
	})
	if err != nil {
		log.Printf("ERROR SearchStandardRecipes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "検索処理中にエラーが発生しました"})
		return
	}

	// エンティティをDTOに変換
	responses := make([]StandardRecipeResponse, 0, len(output.Recipes))
	for _, r := range output.Recipes {
		responses = append(responses, toStandardRecipeResponse(r))
	}

	c.JSON(http.StatusOK, SearchStandardRecipesResponse{Recipes: responses})
}

// GetStandardRecipeDetail は基準レシピ詳細取得APIのハンドラーです。
//
// エンドポイント: GET /api/standard-recipes/:id
// パスパラメータ:
//   - id: 基準レシピID（数値）
func (h *StandardRecipeHandler) GetStandardRecipeDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}

	output, err := h.detailUsecase.Execute(c.Request.Context(), standard_recipe.DetailInput{RecipeID: id})
	if err != nil {
		log.Printf("ERROR GetStandardRecipeDetail (id=%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ取得中にエラーが発生しました"})
		return
	}

	if output.Recipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "基準レシピが見つかりませんでした"})
		return
	}

	c.JSON(http.StatusOK, toStandardRecipeResponse(*output.Recipe))
}

// toStandardRecipeResponse はStandardRecipeエンティティをDTOに変換します。
func toStandardRecipeResponse(r entity.StandardRecipe) StandardRecipeResponse {
	// 材料グループをスライスに変換（JSON配列として返す）
	groups := make([]StandardIngredientGroupResponse, 0, len(r.IngredientGroups))
	for grpName, grp := range r.IngredientGroups {
		items := make([]StandardIngredientItemResponse, 0, len(grp.Items))
		for _, item := range grp.Items {
			items = append(items, StandardIngredientItemResponse{
				Name:  item.Name,
				Count: item.Count,
			})
		}
		groups = append(groups, StandardIngredientGroupResponse{
			GroupName:  grpName,
			TotalCount: grp.TotalCount,
			Items:      items,
		})
	}

	// 手順の変換
	steps := make([]StandardStepResponse, 0, len(r.Steps))
	for _, step := range r.Steps {
		steps = append(steps, StandardStepResponse{
			FoodName: step.FoodName,
			Action:   step.Action,
			Count:    step.Count,
		})
	}

	// 調理時間を文字列ラベルに変換
	cookingTimeLabel := entity.CookingTimeMap[r.CookingTime]

	return StandardRecipeResponse{
		ID:               r.ID,
		Name:             r.Name,
		RecipeCount:      r.RecipeCount,
		CookingTimeLabel: cookingTimeLabel,
		AverageSteps:     r.AverageSteps,
		IngredientGroups: groups,
		Steps:            steps,
	}
}
