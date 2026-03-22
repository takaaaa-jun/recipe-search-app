// Package recipe はパーソナルレシピに関するユースケース（ビジネスロジック）を提供するパッケージです。
// ユースケース層はドメイン層（entity + repository）のみに依存し、
// HTTP, DB技術（MySQL等）の詳細を知りません。
package recipe

import (
	"context"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"backend/internal/domain/entity"
	"backend/internal/domain/repository"
)

// SearchPersonalRecipeUsecase はパーソナルレシピ検索のユースケースを実装します。
// 検索クエリの正規化 → 同義語展開 → DB検索 → サマリー取得 の一連の処理を担います。
type SearchPersonalRecipeUsecase struct {
	recipeRepo  repository.RecipeRepository
	synonymRepo repository.SynonymRepository
}

// NewSearchPersonalRecipeUsecase はSearchPersonalRecipeUsecaseの新しいインスタンスを生成します。
// 依存するリポジトリはDI（依存性注入）で外部から渡されます。
func NewSearchPersonalRecipeUsecase(
	recipeRepo repository.RecipeRepository,
	synonymRepo repository.SynonymRepository,
) *SearchPersonalRecipeUsecase {
	return &SearchPersonalRecipeUsecase{
		recipeRepo:  recipeRepo,
		synonymRepo: synonymRepo,
	}
}

// SearchInput は検索ユースケースへの入力パラメータです。
type SearchInput struct {
	// Query はユーザーが入力した検索クエリ文字列
	// スペース区切りで複数キーワード "鶏肉 -卵" のように指定
	Query string
	// StartID はカーソルページネーションの開始ID（省略時はランダム）
	// 0を渡すと自動でランダムスタートIDを生成します
	StartID int
	// Limit は取得する最大件数（デフォルト: 10）
	Limit int
}

// SearchOutput は検索ユースケースの出力です。
type SearchOutput struct {
	// Recipes は検索結果のレシピサマリーリスト
	Recipes []entity.RecipeSummary
}

// Execute は検索ユースケースを実行します。
// Pythonの personal.py の search() + search.py の search_recipes() に相当します。
func (uc *SearchPersonalRecipeUsecase) Execute(ctx context.Context, input SearchInput) (*SearchOutput, error) {
	// デフォルト値の設定
	if input.Limit <= 0 {
		input.Limit = 10
	}

	// ランダムスタートIDの設定（Pythonの random.randint(1, 1500000) に相当）
	startID := input.StartID
	if startID <= 0 {
		//nolint:gosec // G404: 検索のランダム性はセキュリティ要件ではないため math/rand で十分
		startID = rand.Intn(1500000) + 1
	}

	// Step 1: クエリのパース
	synonymGroups, err := uc.parseQuery(ctx, input.Query)
	if err != nil {
		return nil, err
	}
	if len(synonymGroups) == 0 {
		return &SearchOutput{Recipes: []entity.RecipeSummary{}}, nil
	}

	// Step 2: 食材検索でレシピIDを取得（ランダムスタートから最大limit件）
	foundIDs, err := uc.recipeRepo.FindIDsByIngredientSynonyms(ctx, synonymGroups, startID, input.Limit)
	if err != nil {
		return nil, err
	}

	// Step 3: ヒット数が足りない場合は先頭（ID=1）から補完（ラップアラウンド）
	if len(foundIDs) < input.Limit && startID > 1 {
		needed := input.Limit - len(foundIDs)
		additionalIDs, err := uc.recipeRepo.FindIDsByIngredientSynonyms(ctx, synonymGroups, 1, needed)
		if err != nil {
			return nil, err
		}
		// 重複しないIDのみ追加
		existingIDSet := map[int]struct{}{}
		for _, id := range foundIDs {
			existingIDSet[id] = struct{}{}
		}
		for _, id := range additionalIDs {
			if _, exists := existingIDSet[id]; !exists {
				foundIDs = append(foundIDs, id)
			}
		}
	}

	if len(foundIDs) == 0 {
		return &SearchOutput{Recipes: []entity.RecipeSummary{}}, nil
	}

	// Step 4: IDリストからサマリー情報を取得
	summaries, err := uc.recipeRepo.FindSummariesByIDs(ctx, foundIDs)
	if err != nil {
		return nil, err
	}

	return &SearchOutput{Recipes: summaries}, nil
}

// parseQuery は検索クエリ文字列をパースし、同義語展開済みのグループリストを返します。
// Pythonの _parse_query() 関数に相当します。
//
// クエリ形式:
//   - スペース区切りでAND条件
//   - '-' プレフィックスでNOT条件（現在のパーソナル検索では除外は使用しないが将来の拡張用）
//   - 全角スペース対応
func (uc *SearchPersonalRecipeUsecase) parseQuery(ctx context.Context, query string) ([][]string, error) {
	// 全角スペースを半角スペースに正規化
	normalized := strings.ReplaceAll(query, "　", " ")
	keywords := splitByWhitespace(normalized)
	if len(keywords) == 0 {
		return nil, nil
	}

	// '-' プレフィックスを除いた包含キーワード（Inclusion）のみを抽出
	// （パーソナル検索では exclude は使用しない）
	rawInclusions := []string{}
	for _, k := range keywords {
		if !strings.HasPrefix(k, "-") {
			rawInclusions = append(rawInclusions, k)
		}
	}
	if len(rawInclusions) == 0 {
		return nil, nil
	}

	// 同義語辞書でキーワードを統合（重複同義語の正規化）
	unified, err := uc.synonymRepo.UnifyKeywords(ctx, rawInclusions)
	if err != nil {
		return nil, err
	}

	// 各統合キーワードを同義語グループに展開
	// 例: "鶏肉" → ["鶏肉", "チキン", "とり", "とりにく"]
	synonymGroups := make([][]string, 0, len(unified))
	for _, kw := range unified {
		syns, err := uc.synonymRepo.GetSynonyms(ctx, kw)
		if err != nil {
			return nil, err
		}
		synonymGroups = append(synonymGroups, syns)
	}

	return synonymGroups, nil
}

// splitByWhitespace は文字列をUnicodeのホワイトスペースで分割します。
// strings.Fields と同等ですが、全角スペースの正規化済み入力を想定します。
func splitByWhitespace(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return unicode.IsSpace(r)
	})
}

// GetRecipeDetailUsecase はパーソナルレシピ詳細取得のユースケースです。
type GetRecipeDetailUsecase struct {
	recipeRepo repository.RecipeRepository
}

// NewGetRecipeDetailUsecase はGetRecipeDetailUsecaseの新しいインスタンスを生成します。
func NewGetRecipeDetailUsecase(recipeRepo repository.RecipeRepository) *GetRecipeDetailUsecase {
	return &GetRecipeDetailUsecase{recipeRepo: recipeRepo}
}

// DetailInput はレシピ詳細取得ユースケースへの入力パラメータです。
type DetailInput struct {
	// RecipeID は取得するレシピのID
	RecipeID int
}

// DetailOutput はレシピ詳細取得ユースケースの出力です。
type DetailOutput struct {
	// Recipe はレシピの完全詳細情報（Nilの場合はレシピが存在しない）
	Recipe *entity.Recipe
}

// Execute はレシピ詳細取得ユースケースを実行します。
// 栄養素の1人分計算と基準充足率の計算もここで実施します（ドメインロジック）。
// Pythonの get_recipe_details() + process_recipe_rows() に相当します。
func (uc *GetRecipeDetailUsecase) Execute(ctx context.Context, input DetailInput) (*DetailOutput, error) {
	recipe, err := uc.recipeRepo.FindDetailByID(ctx, input.RecipeID)
	if err != nil {
		return nil, err
	}
	if recipe == nil {
		return &DetailOutput{Recipe: nil}, nil
	}

	// 栄養素の1人分計算（NutritionTotals / ServingSize）
	srvSize := float64(recipe.ServingSize)
	if srvSize <= 0 {
		srvSize = 1
	}
	recipe.NutritionPerServing = entity.NutritionInfo{
		EnergyKcal: recipe.NutritionTotals.EnergyKcal / srvSize,
		ProteinG:   recipe.NutritionTotals.ProteinG / srvSize,
		FatG:       recipe.NutritionTotals.FatG / srvSize,
		CarbsG:     recipe.NutritionTotals.CarbsG / srvSize,
		FiberG:     recipe.NutritionTotals.FiberG / srvSize,
		SaltG:      recipe.NutritionTotals.SaltG / srvSize,
	}

	// 基準充足率の計算（1人分 / 基準摂取量 × 100）
	std := entity.DefaultNutritionStandards
	recipe.NutritionRatios = entity.NutritionRatio{
		Energy:  safeDivide(recipe.NutritionPerServing.EnergyKcal, std.EnergyKcal) * 100,
		Protein: safeDivide(recipe.NutritionPerServing.ProteinG, std.ProteinG) * 100,
		Fat:     safeDivide(recipe.NutritionPerServing.FatG, std.FatG) * 100,
		Carbs:   safeDivide(recipe.NutritionPerServing.CarbsG, std.CarbsG) * 100,
		Fiber:   safeDivide(recipe.NutritionPerServing.FiberG, std.FiberG) * 100,
		Salt:    safeDivide(recipe.NutritionPerServing.SaltG, std.SaltG) * 100,
	}

	return &DetailOutput{Recipe: recipe}, nil
}

// safeDivide はゼロ除算を防ぐ除算ヘルパー関数です。
func safeDivide(numerator, denominator float64) float64 {
	if denominator == 0 {
		return 0
	}
	return numerator / denominator
}

// init はmath/randのグローバルソースを現在時刻でシードします。
// Go 1.20以降は自動でシードされるため不要ですが、互換性のために記載します。
func init() {
	rand.Seed(time.Now().UnixNano()) //nolint:staticcheck
}
