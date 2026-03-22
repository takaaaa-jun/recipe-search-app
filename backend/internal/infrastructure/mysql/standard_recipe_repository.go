// Package mysql はドメイン層のリポジトリインターフェースに対するMySQL実装を提供します。
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend/internal/domain/entity"
)

// StandardRecipeRepository はStandardRecipeRepositoryインターフェースのMySQL実装です。
type StandardRecipeRepository struct {
	db *sql.DB
}

// NewStandardRecipeRepository はStandardRecipeRepositoryの新しいインスタンスを生成します。
func NewStandardRecipeRepository(db *sql.DB) *StandardRecipeRepository {
	return &StandardRecipeRepository{db: db}
}

// FindIDsByName はカテゴリ名（category_medium）に対するLIKE検索でレシピIDを返します。
// AND/NOT条件をサポートし、recipe_count降順で最大5件取得します。
func (r *StandardRecipeRepository) FindIDsByName(
	ctx context.Context,
	inclusions, exclusions []string,
) ([]int, error) {
	if len(inclusions) == 0 && len(exclusions) == 0 {
		return nil, nil
	}

	conditions := []string{}
	args := []interface{}{}

	// 含む条件（LIKE）
	for _, kw := range inclusions {
		conditions = append(conditions, "category_medium LIKE ?")
		args = append(args, "%"+kw+"%")
	}

	// 除外条件（NOT LIKE）
	for _, kw := range exclusions {
		conditions = append(conditions, "category_medium NOT LIKE ?")
		args = append(args, "%"+kw+"%")
	}

	query := fmt.Sprintf(
		"SELECT id FROM standard_recipes WHERE %s ORDER BY recipe_count DESC LIMIT 5",
		strings.Join(conditions, " AND "),
	)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("FindIDsByName query failed: %w", err)
	}
	defer rows.Close()

	ids := []int{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan standard recipe id: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// FindIDsByIngredient は材料名でレシピIDを検索します（AND交差 + スコアリング）。
// Pythonの search_standard_recipes() の ingredient モードに相当します。
func (r *StandardRecipeRepository) FindIDsByIngredient(
	ctx context.Context,
	inclusions, exclusions []string,
) ([]int, error) {
	if len(inclusions) == 0 {
		return nil, nil
	}

	// 各キーワードに対して {standard_recipe_id: count} マップを取得し、AND交差する
	type matchMap map[int]int // recipe_id → count

	keywordMatches := []matchMap{}
	for _, kw := range inclusions {
		mm := matchMap{}
		rows, err := r.db.QueryContext(ctx,
			"SELECT standard_recipe_id, count FROM standard_recipe_ingredients WHERE ingredient_name LIKE ?",
			"%"+kw+"%")
		if err != nil {
			return nil, fmt.Errorf("ingredient search query failed (keyword=%s): %w", kw, err)
		}
		defer rows.Close()
		for rows.Next() {
			var rID, cnt int
			if err := rows.Scan(&rID, &cnt); err != nil {
				return nil, err
			}
			mm[rID] = cnt
		}
		keywordMatches = append(keywordMatches, mm)
	}

	if len(keywordMatches) == 0 {
		return nil, nil
	}

	// 全キーワードのIDを AND 交差
	commonIDs := map[int]struct{}{}
	for rID := range keywordMatches[0] {
		commonIDs[rID] = struct{}{}
	}
	for _, mm := range keywordMatches[1:] {
		for rID := range commonIDs {
			if _, ok := mm[rID]; !ok {
				delete(commonIDs, rID)
			}
		}
	}

	if len(commonIDs) == 0 {
		return nil, nil
	}

	// スコアリング: 全キーワードのcount合計が高い順にソート
	type scoredRecipe struct {
		id    int
		score int
	}
	scored := []scoredRecipe{}
	for rID := range commonIDs {
		score := 0
		for _, mm := range keywordMatches {
			score += mm[rID]
		}
		scored = append(scored, scoredRecipe{id: rID, score: score})
	}
	// スコア降順ソート
	for i := 0; i < len(scored)-1; i++ {
		for j := i + 1; j < len(scored); j++ {
			if scored[i].score < scored[j].score {
				scored[i], scored[j] = scored[j], scored[i]
			}
		}
	}

	// 除外条件の適用
	if len(exclusions) > 0 {
		excludedIDs := map[int]struct{}{}
		for _, kw := range exclusions {
			rows, err := r.db.QueryContext(ctx,
				"SELECT DISTINCT standard_recipe_id FROM standard_recipe_ingredients WHERE ingredient_name LIKE ?",
				"%"+kw+"%")
			if err != nil {
				return nil, fmt.Errorf("exclusion query failed: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var rID int
				if err := rows.Scan(&rID); err != nil {
					return nil, err
				}
				excludedIDs[rID] = struct{}{}
			}
		}
		filtered := []scoredRecipe{}
		for _, s := range scored {
			if _, excluded := excludedIDs[s.id]; !excluded {
				filtered = append(filtered, s)
			}
		}
		scored = filtered
	}

	// 上位5件のIDを返す
	result := []int{}
	for i, s := range scored {
		if i >= 5 {
			break
		}
		result = append(result, s.id)
	}
	return result, nil
}

// FindDetailByIDs は指定IDリストの基準レシピ完全情報を取得します。
// standard_recipes, standard_recipe_ingredients, standard_recipe_steps を別々に取得して組み立てます。
func (r *StandardRecipeRepository) FindDetailByIDs(ctx context.Context, ids []int) ([]entity.StandardRecipe, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	args := make([]interface{}, len(ids))
	for i, id := range ids {
		args[i] = id
	}

	// Step 1: 基本情報の取得
	recipesMap := map[int]*entity.StandardRecipe{}
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT * FROM standard_recipes WHERE id IN (%s)", placeholders), args...)
	if err != nil {
		return nil, fmt.Errorf("FindDetailByIDs basic query failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var rec entity.StandardRecipe
		var cookingTime, avgSteps int
		if err := rows.Scan(&rec.ID, &rec.Name, &rec.RecipeCount, &cookingTime, &avgSteps); err != nil {
			return nil, fmt.Errorf("failed to scan standard recipe: %w", err)
		}
		rec.CookingTime = cookingTime
		rec.AverageSteps = avgSteps
		rec.IngredientGroups = map[string]entity.StandardIngredientGroup{}
		recipesMap[rec.ID] = &rec
	}

	// Step 2: 材料の取得
	if err := r.loadIngredientsForMap(ctx, recipesMap, ids, placeholders, args); err != nil {
		return nil, err
	}

	// Step 3: 手順の取得
	if err := r.loadStepsForMap(ctx, recipesMap, ids, placeholders, args); err != nil {
		return nil, err
	}

	// IDの順序を保持して返す
	result := make([]entity.StandardRecipe, 0, len(ids))
	for _, id := range ids {
		if rec, ok := recipesMap[id]; ok {
			result = append(result, *rec)
		}
	}
	return result, nil
}

// FindDetailByID は指定IDの基準レシピ1件の完全詳細情報を取得します。
func (r *StandardRecipeRepository) FindDetailByID(ctx context.Context, id int) (*entity.StandardRecipe, error) {
	results, err := r.FindDetailByIDs(ctx, []int{id})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	return &results[0], nil
}

// loadIngredientsForMap は指定IDリストの材料情報をロードしてrecipesMapに追加します。
func (r *StandardRecipeRepository) loadIngredientsForMap(
	ctx context.Context,
	recipesMap map[int]*entity.StandardRecipe,
	ids []int,
	placeholders string,
	args []interface{},
) error {
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT standard_recipe_id, group_name, ingredient_name, count FROM standard_recipe_ingredients WHERE standard_recipe_id IN (%s)", placeholders),
		args...)
	if err != nil {
		return fmt.Errorf("ingredient load query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rID, count int
		var groupName, ingredientName sql.NullString
		if err := rows.Scan(&rID, &groupName, &ingredientName, &count); err != nil {
			return fmt.Errorf("failed to scan ingredient: %w", err)
		}

		rec, ok := recipesMap[rID]
		if !ok {
			continue
		}

		name := ingredientName.String
		if name == "all" {
			continue // "all"は集計用の行のためスキップ
		}

		grpName := groupName.String
		if grpName == "" {
			grpName = "その他"
		}

		grp := rec.IngredientGroups[grpName]
		grp.TotalCount += count
		grp.Items = append(grp.Items, entity.StandardIngredientItem{Name: name, Count: count})
		rec.IngredientGroups[grpName] = grp
	}
	return nil
}

// loadStepsForMap は指定IDリストの手順情報をロードしてrecipesMapに追加します。
func (r *StandardRecipeRepository) loadStepsForMap(
	ctx context.Context,
	recipesMap map[int]*entity.StandardRecipe,
	ids []int,
	placeholders string,
	args []interface{},
) error {
	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT standard_recipe_id, food_name, action, count FROM standard_recipe_steps WHERE standard_recipe_id IN (%s) ORDER BY count DESC", placeholders),
		args...)
	if err != nil {
		return fmt.Errorf("steps load query failed: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rID, count int
		var foodName, action string
		if err := rows.Scan(&rID, &foodName, &action, &count); err != nil {
			return fmt.Errorf("failed to scan step: %w", err)
		}
		if rec, ok := recipesMap[rID]; ok {
			rec.Steps = append(rec.Steps, entity.StandardStep{
				FoodName: foodName,
				Action:   action,
				Count:    count,
			})
		}
	}
	return nil
}
