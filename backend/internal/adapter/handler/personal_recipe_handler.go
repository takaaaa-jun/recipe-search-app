// Package handler はHTTPリクエストの受け取りとレスポンス生成を担うアダプター層のパッケージです。
// ハンドラーはHTTPの詳細（リクエストパラメータの取得、JSONレスポンスの返却）を担当し、
// ビジネスロジックはユースケース層に委任します。
package handler

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/usecase/recipe"

	"github.com/gin-gonic/gin"
)

// PersonalRecipeHandler はパーソナルレシピ関連のHTTPハンドラーを保持する構造体です。
// 依存するユースケースはDI（依存性注入）で設定されます。
type PersonalRecipeHandler struct {
	// searchUsecase はパーソナルレシピ検索ユースケースです
	searchUsecase *recipe.SearchPersonalRecipeUsecase
	// detailUsecase はパーソナルレシピ詳細取得ユースケースです
	detailUsecase *recipe.GetRecipeDetailUsecase
}

// NewPersonalRecipeHandler はPersonalRecipeHandlerの新しいインスタンスを生成します。
func NewPersonalRecipeHandler(
	searchUsecase *recipe.SearchPersonalRecipeUsecase,
	detailUsecase *recipe.GetRecipeDetailUsecase,
) *PersonalRecipeHandler {
	return &PersonalRecipeHandler{
		searchUsecase: searchUsecase,
		detailUsecase: detailUsecase,
	}
}

// ===========================================================================
// レスポンス用 DTO（Data Transfer Object）
// ===========================================================================

// RecipeSummaryResponse は検索結果の1件を表すJSONレスポンス構造体です。
// フロントエンドに返す際の表示用フォーマットに合わせています。
type RecipeSummaryResponse struct {
	ID          int    `json:"id"`           // レシピID
	Title       string `json:"title"`        // タイトル
	Description string `json:"description"`  // 説明（先頭20文字程度）
	PublishedAt string `json:"published_at"` // 公開日（ISO 8601形式）
}

// SearchResponse は検索APIのレスポンスです。
type SearchResponse struct {
	Recipes []RecipeSummaryResponse `json:"recipes"` // 検索結果リスト
}

// IngredientResponse は材料情報のJSONレスポンス構造体です。
type IngredientResponse struct {
	ID                  int     `json:"id"`                   // 材料ID
	Name                string  `json:"name"`                 // 材料名
	Quantity            string  `json:"quantity"`             // 分量（表示用）
	NormalizedQuantityG float64 `json:"normalized_quantity_g"` // グラム換算量
}

// NutritionInfoResponse は栄養素情報のJSONレスポンス構造体です。
type NutritionInfoResponse struct {
	EnergyKcal float64 `json:"energy_kcal"` // エネルギー (kcal)
	ProteinG   float64 `json:"protein_g"`   // たんぱく質 (g)
	FatG       float64 `json:"fat_g"`       // 脂質 (g)
	CarbsG     float64 `json:"carbs_g"`     // 炭水化物 (g)
	FiberG     float64 `json:"fiber_g"`     // 食物繊維 (g)
	SaltG      float64 `json:"salt_g"`      // 食塩相当量 (g)
}

// NutritionRatioResponse は栄養素充足率のJSONレスポンス構造体です。
type NutritionRatioResponse struct {
	Energy  float64 `json:"energy"`  // エネルギー充足率 (%)
	Protein float64 `json:"protein"` // たんぱく質充足率 (%)
	Fat     float64 `json:"fat"`     // 脂質充足率 (%)
	Carbs   float64 `json:"carbs"`   // 炭水化物充足率 (%)
	Fiber   float64 `json:"fiber"`   // 食物繊維充足率 (%)
	Salt    float64 `json:"salt"`    // 食塩相当量充足率 (%)
}

// NutritionStandardsResponse は栄養素基準摂取量のJSONレスポンス構造体です。
type NutritionStandardsResponse struct {
	EnergyKcal float64 `json:"energy_kcal"` // エネルギー基準量 (kcal)
	ProteinG   float64 `json:"protein_g"`   // たんぱく質基準量 (g)
	FatG       float64 `json:"fat_g"`       // 脂質基準量 (g)
	CarbsG     float64 `json:"carbs_g"`     // 炭水化物基準量 (g)
	FiberG     float64 `json:"fiber_g"`     // 食物繊維基準量 (g)
	SaltG      float64 `json:"salt_g"`      // 食塩相当量基準量 (g)
}

// RecipeDetailResponse はレシピ詳細APIのレスポンス構造体です。
type RecipeDetailResponse struct {
	ID                  int                        `json:"id"`
	Title               string                     `json:"title"`
	Description         string                     `json:"description"`
	CookingTime         string                     `json:"cooking_time"`
	ServingFor          string                     `json:"serving_for"`
	ServingSize         int                        `json:"serving_size"`
	PublishedAt         string                     `json:"published_at"`
	Ingredients         []IngredientResponse       `json:"ingredients"`
	Steps               []StepResponse             `json:"steps"`
	NutritionTotals     NutritionInfoResponse      `json:"nutrition_totals"`
	NutritionPerServing NutritionInfoResponse      `json:"nutrition_per_serving"`
	NutritionRatios     NutritionRatioResponse     `json:"nutrition_ratios"`
	Standards           NutritionStandardsResponse `json:"standards"`
}

// StepResponse は手順情報のJSONレスポンス構造体です。
type StepResponse struct {
	Position int    `json:"position"` // 手順番号（1始まり）
	Memo     string `json:"memo"`     // 手順説明文
}

// ===========================================================================
// ハンドラーメソッド
// ===========================================================================

// SearchRecipes はパーソナルレシピ検索APIのハンドラーです。
//
// エンドポイント: GET /api/recipes/search
// クエリパラメータ:
//   - query: 検索キーワード（スペース区切り）
//   - start_id: カーソルページネーションの開始ID（省略時はランダム）
func (h *PersonalRecipeHandler) SearchRecipes(c *gin.Context) {
	query := c.Query("query")
	startIDStr := c.Query("start_id")

	startID := 0
	if startIDStr != "" {
		parsed, err := strconv.Atoi(startIDStr)
		if err == nil && parsed > 0 {
			startID = parsed
		}
	}

	output, err := h.searchUsecase.Execute(c.Request.Context(), recipe.SearchInput{
		Query:   query,
		StartID: startID,
		Limit:   10,
	})
	if err != nil {
		log.Printf("ERROR SearchRecipes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "検索処理中にエラーが発生しました"})
		return
	}

	// エンティティをDTOに変換
	responses := make([]RecipeSummaryResponse, 0, len(output.Recipes))
	for _, r := range output.Recipes {
		publishedAt := ""
		if r.PublishedAt != nil {
			publishedAt = r.PublishedAt.Format(time.RFC3339)
		}
		// 説明文を先頭20文字で切り取り（フロントエンドでも対応可だが互換性のため）
		desc := r.Description
		if len([]rune(desc)) > 20 {
			desc = string([]rune(desc)[:20]) + "..."
		}
		responses = append(responses, RecipeSummaryResponse{
			ID:          r.ID,
			Title:       r.Title,
			Description: desc,
			PublishedAt: publishedAt,
		})
	}

	c.JSON(http.StatusOK, SearchResponse{Recipes: responses})
}

// GetRecipeDetail はパーソナルレシピ詳細取得APIのハンドラーです。
//
// エンドポイント: GET /api/recipes/:id
// パスパラメータ:
//   - id: レシピID（数値）
func (h *PersonalRecipeHandler) GetRecipeDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDが不正です"})
		return
	}

	output, err := h.detailUsecase.Execute(c.Request.Context(), recipe.DetailInput{RecipeID: id})
	if err != nil {
		log.Printf("ERROR GetRecipeDetail (id=%d): %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "データ取得中にエラーが発生しました"})
		return
	}

	if output.Recipe == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "レシピが見つかりませんでした"})
		return
	}

	r := output.Recipe

	// 材料DTOへの変換
	ingredients := make([]IngredientResponse, 0, len(r.Ingredients))
	for _, ing := range r.Ingredients {
		ingredients = append(ingredients, IngredientResponse{
			ID:                  ing.ID,
			Name:                ing.Name,
			Quantity:            ing.Quantity,
			NormalizedQuantityG: ing.NormalizedQuantityG,
		})
	}

	// 手順DTOへの変換
	steps := make([]StepResponse, 0, len(r.Steps))
	for _, step := range r.Steps {
		steps = append(steps, StepResponse{
			Position: step.Position,
			Memo:     step.Memo,
		})
	}

	publishedAt := ""
	if r.PublishedAt != nil {
		publishedAt = r.PublishedAt.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, RecipeDetailResponse{
		ID:          r.ID,
		Title:       r.Title,
		Description: r.Description,
		CookingTime: r.CookingTime,
		ServingFor:  r.ServingFor,
		ServingSize: r.ServingSize,
		PublishedAt: publishedAt,
		Ingredients: ingredients,
		Steps:       steps,
		NutritionTotals: NutritionInfoResponse{
			EnergyKcal: r.NutritionTotals.EnergyKcal,
			ProteinG:   r.NutritionTotals.ProteinG,
			FatG:       r.NutritionTotals.FatG,
			CarbsG:     r.NutritionTotals.CarbsG,
			FiberG:     r.NutritionTotals.FiberG,
			SaltG:      r.NutritionTotals.SaltG,
		},
		NutritionPerServing: NutritionInfoResponse{
			EnergyKcal: r.NutritionPerServing.EnergyKcal,
			ProteinG:   r.NutritionPerServing.ProteinG,
			FatG:       r.NutritionPerServing.FatG,
			CarbsG:     r.NutritionPerServing.CarbsG,
			FiberG:     r.NutritionPerServing.FiberG,
			SaltG:      r.NutritionPerServing.SaltG,
		},
		NutritionRatios: NutritionRatioResponse{
			Energy:  r.NutritionRatios.Energy,
			Protein: r.NutritionRatios.Protein,
			Fat:     r.NutritionRatios.Fat,
			Carbs:   r.NutritionRatios.Carbs,
			Fiber:   r.NutritionRatios.Fiber,
			Salt:    r.NutritionRatios.Salt,
		},
		Standards: NutritionStandardsResponse{
			EnergyKcal: r.Standards.EnergyKcal,
			ProteinG:   r.Standards.ProteinG,
			FatG:       r.Standards.FatG,
			CarbsG:     r.Standards.CarbsG,
			FiberG:     r.Standards.FiberG,
			SaltG:      r.Standards.SaltG,
		},
	})
}

// truncateRunes はUnicode文字単位で文字列をmaxLen文字に切り捨てます。
// 日本語などマルチバイト文字に対応しています。
func truncateRunes(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// _ はstrings.Containsを参照するだけのブランク変数（未使用インポート回避）
var _ = strings.Contains
