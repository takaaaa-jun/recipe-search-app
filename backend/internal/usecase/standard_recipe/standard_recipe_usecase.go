// Package standard_recipe は基準レシピに関するユースケース（ビジネスロジック）を提供するパッケージです。
package standard_recipe

import (
	"context"
	"strings"
	"unicode"

	"backend/internal/domain/entity"
	"backend/internal/domain/repository"
)

// SearchStandardRecipeUsecase は基準レシピ検索のユースケースを実装します。
// 「レシピ名で検索」と「材料名で検索」の2つのモードをサポートします。
type SearchStandardRecipeUsecase struct {
	standardRepo repository.StandardRecipeRepository
	synonymRepo  repository.SynonymRepository
}

// NewSearchStandardRecipeUsecase はSearchStandardRecipeUsecaseの新しいインスタンスを生成します。
func NewSearchStandardRecipeUsecase(
	standardRepo repository.StandardRecipeRepository,
	synonymRepo repository.SynonymRepository,
) *SearchStandardRecipeUsecase {
	return &SearchStandardRecipeUsecase{
		standardRepo: standardRepo,
		synonymRepo:  synonymRepo,
	}
}

// SearchMode は基準レシピの検索モードを表す型です。
type SearchMode string

const (
	// SearchModeRecipe はカテゴリ名（レシピ名）で検索するモードです。
	SearchModeRecipe SearchMode = "recipe"
	// SearchModeIngredient は材料名で検索するモードです。
	SearchModeIngredient SearchMode = "ingredient"
)

// SearchInput は基準レシピ検索ユースケースへの入力パラメータです。
type SearchInput struct {
	// Query は検索クエリ文字列（スペース区切り、'-'プレフィックスでNOT条件）
	Query string
	// Mode は検索モード（"recipe" または "ingredient"）
	Mode SearchMode
}

// SearchOutput は基準レシピ検索ユースケースの出力です。
type SearchOutput struct {
	// Recipes は検索結果の基準レシピリスト（最大5件）
	Recipes []entity.StandardRecipe
}

// Execute は基準レシピ検索ユースケースを実行します。
// Pythonの search_standard_recipes() に相当します。
func (uc *SearchStandardRecipeUsecase) Execute(ctx context.Context, input SearchInput) (*SearchOutput, error) {
	// クエリのパース（全角スペース対応 + AND/NOT分割）
	inclusions, exclusions := parseStandardQuery(input.Query)
	if len(inclusions) == 0 && len(exclusions) == 0 {
		return &SearchOutput{Recipes: []entity.StandardRecipe{}}, nil
	}

	var ids []int
	var err error

	// 検索モードに応じてIDを取得
	switch input.Mode {
	case SearchModeIngredient:
		ids, err = uc.standardRepo.FindIDsByIngredient(ctx, inclusions, exclusions)
	default: // SearchModeRecipe がデフォルト
		ids, err = uc.standardRepo.FindIDsByName(ctx, inclusions, exclusions)
	}
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return &SearchOutput{Recipes: []entity.StandardRecipe{}}, nil
	}

	// IDリストから完全詳細情報を取得
	recipes, err := uc.standardRepo.FindDetailByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return &SearchOutput{Recipes: recipes}, nil
}

// GetStandardRecipeDetailUsecase は基準レシピ詳細取得のユースケースです。
type GetStandardRecipeDetailUsecase struct {
	standardRepo repository.StandardRecipeRepository
}

// NewGetStandardRecipeDetailUsecase はGetStandardRecipeDetailUsecaseのインスタンスを生成します。
func NewGetStandardRecipeDetailUsecase(standardRepo repository.StandardRecipeRepository) *GetStandardRecipeDetailUsecase {
	return &GetStandardRecipeDetailUsecase{standardRepo: standardRepo}
}

// DetailInput は基準レシピ詳細取得ユースケースへの入力パラメータです。
type DetailInput struct {
	// RecipeID は取得する基準レシピのID
	RecipeID int
}

// DetailOutput は基準レシピ詳細取得ユースケースの出力です。
type DetailOutput struct {
	// Recipe は基準レシピの完全詳細情報（Nilの場合はレシピが存在しない）
	Recipe *entity.StandardRecipe
}

// Execute は基準レシピ詳細取得ユースケースを実行します。
func (uc *GetStandardRecipeDetailUsecase) Execute(ctx context.Context, input DetailInput) (*DetailOutput, error) {
	recipe, err := uc.standardRepo.FindDetailByID(ctx, input.RecipeID)
	if err != nil {
		return nil, err
	}
	return &DetailOutput{Recipe: recipe}, nil
}

// parseStandardQuery はクエリ文字列を包含キーワードと除外キーワードに分割します。
// '-' プレフィックスのキーワードはexclusionsに、それ以外はinclusionsに格納します。
func parseStandardQuery(query string) (inclusions []string, exclusions []string) {
	normalized := strings.ReplaceAll(query, "　", " ")
	keywords := strings.FieldsFunc(normalized, func(r rune) bool {
		return unicode.IsSpace(r)
	})

	for _, kw := range keywords {
		if strings.HasPrefix(kw, "-") && len(kw) > 1 {
			exclusions = append(exclusions, kw[1:]) // '-' を除いたキーワード
		} else if !strings.HasPrefix(kw, "-") {
			inclusions = append(inclusions, kw)
		}
	}
	return inclusions, exclusions
}
