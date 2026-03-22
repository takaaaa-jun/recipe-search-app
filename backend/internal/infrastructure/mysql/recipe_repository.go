// Package mysql はドメイン層のリポジトリインターフェースに対するMySQL実装を提供します。
// このパッケージはinfrastructure層に属し、外部DB技術（MySQL）の詳細をカプセル化します。
// ドメイン層のrepositoryインターフェースを実装するため、これを依存関係として使用します。
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend/internal/domain/entity"
)

// RecipeRepository はRecipeRepository インターフェースのMySQL実装です。
type RecipeRepository struct {
	// db はデータベース接続プールです。Read-onlyな操作のみ行います。
	db *sql.DB
}

// NewRecipeRepository はRecipeRepositoryの新しいインスタンスを生成します。
// cmd/server/main.go のDIコンテナからのみ呼び出されます。
func NewRecipeRepository(db *sql.DB) *RecipeRepository {
	return &RecipeRepository{db: db}
}

// FindIDsByIngredientSynonyms は食材の同義語グループに一致するレシピIDを検索します。
// Pythonの search_recipes() 関数のロジックを移植したものです。
//
// アルゴリズムの概要:
//   1. 単一グループ: 各同義語ごとに range scan を実行し、結果をマージ（scatter-gather）
//   2. 複数グループ: レアリティファースト戦略 + バッチ検証（paged driver + vectorized verification）
func (r *RecipeRepository) FindIDsByIngredientSynonyms(
	ctx context.Context,
	synonymGroups [][]string,
	startID, limit int,
) ([]int, error) {
	if len(synonymGroups) == 0 {
		return nil, nil
	}

	if len(synonymGroups) == 1 {
		return r.findIDsSingleGroup(ctx, synonymGroups[0], startID, limit)
	}
	return r.findIDsMultipleGroups(ctx, synonymGroups, startID, limit)
}

// findIDsSingleGroup は単一の同義語グループに対するScatter-Gather最適化検索を実行します。
// 各同義語を個別にRange Scan（カバリングインデックス活用）し、結果をマージ・重複排除します。
func (r *RecipeRepository) findIDsSingleGroup(
	ctx context.Context,
	synonyms []string,
	startID, limit int,
) ([]int, error) {
	gatheredIDs := []int{}

	for _, syn := range synonyms {
		// name カラムのカバリングインデックス（name, recipe_id）を活用したRange Scan
		query := `
			SELECT recipe_id
			FROM ingredients
			WHERE name = ?
			AND recipe_id >= ?
			ORDER BY recipe_id ASC
			LIMIT ?`

		rows, err := r.db.QueryContext(ctx, query, syn, startID, limit)
		if err != nil {
			return nil, fmt.Errorf("foodIngredient single group query failed (synonym=%s): %w", syn, err)
		}
		defer rows.Close()

		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err != nil {
				return nil, fmt.Errorf("failed to scan recipe_id: %w", err)
			}
			gatheredIDs = append(gatheredIDs, id)
		}
	}

	// 重複排除と昇順ソート
	return deduplicateAndSort(gatheredIDs, limit), nil
}

// findIDsMultipleGroups は複数の食材グループに対するレアリティファースト + バッチ検証を実行します。
// 最も件数の少ない食材グループを「ドライバー」として使い、
// それに一致するIDを取得してから他のグループで絞り込みます。
func (r *RecipeRepository) findIDsMultipleGroups(
	ctx context.Context,
	synonymGroups [][]string,
	startID, limit int,
) ([]int, error) {
	const fetchBatchSize = 1000
	const maxScanCandidates = 10000

	// Step 1: 各グループの件数を推定し、レアリティ順にソート
	type groupWithCount struct {
		synonyms []string
		count    int
	}
	sortedGroups := make([]groupWithCount, 0, len(synonymGroups))

	for _, grp := range synonymGroups {
		total := 0
		for _, syn := range grp {
			var cnt int
			err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ingredients WHERE name = ?", syn).Scan(&cnt)
			if err == nil {
				total += cnt
			}
		}
		sortedGroups = append(sortedGroups, groupWithCount{synonyms: grp, count: total})
	}

	// 件数の少ない順（レアリティ順）でソート
	for i := 0; i < len(sortedGroups)-1; i++ {
		for j := i + 1; j < len(sortedGroups); j++ {
			if sortedGroups[i].count > sortedGroups[j].count {
				sortedGroups[i], sortedGroups[j] = sortedGroups[j], sortedGroups[i]
			}
		}
	}

	driverSynonyms := sortedGroups[0].synonyms
	otherGroups := make([][]string, 0, len(sortedGroups)-1)
	for _, g := range sortedGroups[1:] {
		otherGroups = append(otherGroups, g.synonyms)
	}

	// Step 2: ドライバー + バッチ検証ループ
	foundIDs := []int{}
	currentStartID := startID
	scannedCount := 0

	for len(foundIDs) < limit && scannedCount < maxScanCandidates {
		// ドライバーグループで候補IDを取得
		candidates := []int{}
		for _, syn := range driverSynonyms {
			rows, err := r.db.QueryContext(ctx,
				`SELECT recipe_id FROM ingredients WHERE name = ? AND recipe_id >= ? ORDER BY recipe_id ASC LIMIT ?`,
				syn, currentStartID, fetchBatchSize)
			if err != nil {
				return nil, fmt.Errorf("driver query failed: %w", err)
			}
			defer rows.Close()
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					return nil, err
				}
				candidates = append(candidates, id)
			}
		}

		if len(candidates) == 0 {
			break
		}

		candidates = deduplicateAndSort(candidates, fetchBatchSize)
		lastCandidateID := candidates[len(candidates)-1]
		scannedCount += len(candidates)

		// 他のグループで候補を絞り込む（AND条件の検証）
		matchingIDs := toSet(candidates)
		for _, grp := range otherGroups {
			if len(matchingIDs) == 0 {
				break
			}
			verified, err := r.verifyBatch(ctx, setToSlice(matchingIDs), grp)
			if err != nil {
				return nil, err
			}
			matchingIDs = verified
		}

		// 結果をマージ（重複排除）
		for _, id := range sortedInts(setToSlice(matchingIDs)) {
			alreadyFound := false
			for _, f := range foundIDs {
				if f == id {
					alreadyFound = true
					break
				}
			}
			if !alreadyFound {
				foundIDs = append(foundIDs, id)
			}
			if len(foundIDs) >= limit {
				break
			}
		}

		currentStartID = lastCandidateID + 1
	}

	return foundIDs, nil
}

// verifyBatch は候補IDリストの中で、指定した同義語グループの食材を持つIDを返します。
// Pythonの verify_batch() 関数に相当します。
func (r *RecipeRepository) verifyBatch(ctx context.Context, candidateIDs []int, synonyms []string) (map[int]struct{}, error) {
	if len(candidateIDs) == 0 {
		return map[int]struct{}{}, nil
	}

	idPlaceholders := strings.Repeat("?,", len(candidateIDs))
	idPlaceholders = idPlaceholders[:len(idPlaceholders)-1]
	namePlaceholders := strings.Repeat("?,", len(synonyms))
	namePlaceholders = namePlaceholders[:len(namePlaceholders)-1]

	query := fmt.Sprintf(`
		SELECT DISTINCT recipe_id
		FROM ingredients
		WHERE name IN (%s)
		AND recipe_id IN (%s)`, namePlaceholders, idPlaceholders)

	args := make([]interface{}, 0, len(synonyms)+len(candidateIDs))
	for _, s := range synonyms {
		args = append(args, s)
	}
	for _, id := range candidateIDs {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("verifyBatch query failed: %w", err)
	}
	defer rows.Close()

	result := map[int]struct{}{}
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		result[id] = struct{}{}
	}
	return result, nil
}

// FindDetailByID は指定IDのレシピの完全詳細情報をDBから取得します。
// recipes, ingredients, steps, ingredient_structured, ingredient_units,
// nutritions, recipe_nutrition_info を結合した大きなクエリを実行します。
func (r *RecipeRepository) FindDetailByID(ctx context.Context, id int) (*entity.Recipe, error) {
	query := `
		SELECT
			r.id, r.title, r.description,
			r.cooking_time, r.serving_for, r.published_at, r.attribute,
			i.id AS ingredient_id,
			i.name AS ingredient_name, i.quantity,
			s.position, s.memo AS step_memo,
			ist.normalized_name,
			iu.normalized_quantity,
			n.enerc_kcal, n.prot, n.fat, n.choavldf, n.fib, n.nacl_eq,
			rni.serving_size,
			rni.calories AS total_calories,
			rni.protein AS total_protein,
			rni.fat AS total_fat,
			rni.carbohydrates AS total_carbohydrates,
			rni.fiber AS total_fiber,
			rni.salt AS total_salt
		FROM recipes AS r
		LEFT JOIN ingredients AS i ON r.id = i.recipe_id
		LEFT JOIN steps AS s ON r.id = s.recipe_id
		LEFT JOIN ingredient_structured AS ist ON i.id = ist.ingredient_id
		LEFT JOIN ingredient_units AS iu ON i.id = iu.ingredient_id
		LEFT JOIN nutritions AS n ON ist.normalized_name = n.name COLLATE utf8mb4_general_ci
		LEFT JOIN recipe_nutrition_info AS rni ON r.id = rni.recipe_id
		WHERE r.id = ?
		ORDER BY i.id, s.position ASC`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("FindDetailByID query failed (id=%d): %w", id, err)
	}
	defer rows.Close()

	recipe, err := buildRecipeFromRows(rows)
	if err != nil {
		return nil, err
	}
	return recipe, nil
}

// FindSummariesByIDs は指定されたIDリストに一致するレシピサマリーを取得します。
// FIELD関数でIDの順序を保持します。
func (r *RecipeRepository) FindSummariesByIDs(ctx context.Context, ids []int) ([]entity.RecipeSummary, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := strings.Repeat("?,", len(ids))
	placeholders = placeholders[:len(placeholders)-1]

	// FIELD(id, ...)でIDの指定順序を維持する
	query := fmt.Sprintf(
		"SELECT id, title, description, published_at FROM recipes WHERE id IN (%s) ORDER BY FIELD(id, %s)",
		placeholders, placeholders,
	)

	// パラメータを2回（IN句 + FIELD句）分用意する
	args := make([]interface{}, 0, len(ids)*2)
	for _, id := range ids {
		args = append(args, id)
	}
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("FindSummariesByIDs query failed: %w", err)
	}
	defer rows.Close()

	summaries := []entity.RecipeSummary{}
	for rows.Next() {
		var s entity.RecipeSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Description, &s.PublishedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recipe summary: %w", err)
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

// ===========================================================================
// 内部ヘルパー関数群
// ===========================================================================

// buildRecipeFromRows はJOINクエリの結果行からRecipeエンティティを構築します。
// Pythonの build_recipes_dict() + process_recipe_rows() に相当します。
func buildRecipeFromRows(rows *sql.Rows) (*entity.Recipe, error) {
	// 材料と手順の重複排除のためにマップを使用
	ingredientMap := map[int]*entity.Ingredient{}
	stepMap := map[int]entity.Step{}

	var recipe *entity.Recipe

	for rows.Next() {
		// JOINの結果でNULLになり得る列をポインタ型でスキャン
		var (
			id, cookingTimeCode        int
			title, description         string
			servingFor, attribute      sql.NullString
			publishedAt                sql.NullTime
			ingredientID               sql.NullInt64
			ingredientName, quantity   sql.NullString
			stepPosition               sql.NullInt64
			stepMemo, normalizedName   sql.NullString
			normalizedQuantity         sql.NullFloat64
			enercKcal, prot            sql.NullFloat64
			fat, choavldf, fib, naclEq sql.NullFloat64
			servingSize                sql.NullInt64
			totalCalories              sql.NullFloat64
			totalProtein, totalFat     sql.NullFloat64
			totalCarbs, totalFiber     sql.NullFloat64
			totalSalt                  sql.NullFloat64
		)

		if err := rows.Scan(
			&id, &title, &description,
			&cookingTimeCode, &servingFor, &publishedAt, &attribute,
			&ingredientID, &ingredientName, &quantity,
			&stepPosition, &stepMemo,
			&normalizedName, &normalizedQuantity,
			&enercKcal, &prot, &fat, &choavldf, &fib, &naclEq,
			&servingSize,
			&totalCalories, &totalProtein, &totalFat, &totalCarbs, &totalFiber, &totalSalt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan recipe detail row: %w", err)
		}

		// レシピのベース情報は最初の行でのみ設定（JOIN重複分は無視）
		if recipe == nil {
			srvSize := int(servingSize.Int64)
			if srvSize <= 0 {
				srvSize = 1 // ゼロ除算防止
			}
			recipe = &entity.Recipe{
				ID:          id,
				Title:       title,
				Description: description,
				CookingTime: entity.CookingTimeMap[cookingTimeCode],
				ServingFor:  servingFor.String,
				ServingSize: srvSize,
				NutritionTotals: entity.NutritionInfo{
					EnergyKcal: totalCalories.Float64,
					ProteinG:   totalProtein.Float64,
					FatG:       totalFat.Float64,
					CarbsG:     totalCarbs.Float64,
					FiberG:     totalFiber.Float64,
					SaltG:      totalSalt.Float64,
				},
				Standards: entity.DefaultNutritionStandards,
			}
			if publishedAt.Valid {
				recipe.PublishedAt = &publishedAt.Time
			}
		}

		// 材料情報の追加（同一材料IDの重複スキップ）
		if ingredientID.Valid {
			ingID := int(ingredientID.Int64)
			if _, exists := ingredientMap[ingID]; !exists {
				quantityG := normalizedQuantity.Float64

				// 100gあたりの栄養素 × グラム数 = 実量の栄養素
				ingNutrition := entity.IngredientNutrition{
					EnergyKcal: (enercKcal.Float64 / 100.0) * quantityG,
					ProteinG:   (prot.Float64 / 100.0) * quantityG,
					FatG:       (fat.Float64 / 100.0) * quantityG,
					CarbsG:     ((choavldf.Float64 + fib.Float64) / 100.0) * quantityG,
				}
				ingredientMap[ingID] = &entity.Ingredient{
					ID:                  ingID,
					Name:                ingredientName.String,
					Quantity:            quantity.String,
					NormalizedQuantityG: quantityG,
					Nutrition:           ingNutrition,
				}
			}
		}

		// 手順情報の追加（同一positionの重複スキップ）
		if stepMemo.Valid && stepPosition.Valid {
			pos := int(stepPosition.Int64)
			if _, exists := stepMap[pos]; !exists {
				stepMap[pos] = entity.Step{Position: pos, Memo: stepMemo.String}
			}
		}
	}

	if recipe == nil {
		return nil, nil // レコードなし
	}

	// マップから順序付きスライスへ変換
	recipe.Ingredients = mapToIngredientSlice(ingredientMap)
	recipe.Steps = mapToStepSlice(stepMap)

	return recipe, nil
}

// mapToIngredientSlice はingredientMapをID昇順のスライスに変換します。
func mapToIngredientSlice(m map[int]*entity.Ingredient) []entity.Ingredient {
	result := make([]entity.Ingredient, 0, len(m))
	for _, ing := range m {
		result = append(result, *ing)
	}
	// ID昇順でソート
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].ID > result[j].ID {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// mapToStepSlice はstepMapをposition昇順のスライスに変換します。
func mapToStepSlice(m map[int]entity.Step) []entity.Step {
	result := make([]entity.Step, 0, len(m))
	for _, step := range m {
		result = append(result, step)
	}
	// position昇順でソート
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i].Position > result[j].Position {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return result
}

// deduplicateAndSort はintスライスを昇順ソートして重複を除去し、先頭limit件を返します。
func deduplicateAndSort(ids []int, limit int) []int {
	seen := map[int]struct{}{}
	unique := []int{}
	for _, id := range ids {
		if _, ok := seen[id]; !ok {
			seen[id] = struct{}{}
			unique = append(unique, id)
		}
	}
	// バブルソート（件数が少ないため十分）
	for i := 0; i < len(unique)-1; i++ {
		for j := i + 1; j < len(unique); j++ {
			if unique[i] > unique[j] {
				unique[i], unique[j] = unique[j], unique[i]
			}
		}
	}
	if len(unique) > limit {
		return unique[:limit]
	}
	return unique
}

// toSet はスライスをsetに変換します。
func toSet(ids []int) map[int]struct{} {
	s := make(map[int]struct{}, len(ids))
	for _, id := range ids {
		s[id] = struct{}{}
	}
	return s
}

// setToSlice はsetをスライスに変換します。
func setToSlice(s map[int]struct{}) []int {
	result := make([]int, 0, len(s))
	for id := range s {
		result = append(result, id)
	}
	return result
}

// sortedInts はintスライスを昇順にソートして返します。
func sortedInts(ids []int) []int {
	for i := 0; i < len(ids)-1; i++ {
		for j := i + 1; j < len(ids); j++ {
			if ids[i] > ids[j] {
				ids[i], ids[j] = ids[j], ids[i]
			}
		}
	}
	return ids
}
