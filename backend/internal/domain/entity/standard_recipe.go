// Package entity はアプリケーションのドメインエンティティを定義するパッケージです。
package entity

// ===========================================================================
// 基準レシピ（標準レシピ）関連エンティティ
// ===========================================================================

// StandardRecipe は「基準レシピ」のドメインエンティティです。
// 多数の個人レシピを分析し、代表的な材料と手順をまとめた標準的なレシピ情報を表します。
// DBのstandard_recipesテーブルに対応します。
type StandardRecipe struct {
	ID           int    // 基準レシピID（PK）
	Name         string // カテゴリ名（例: "野菜炒め"）。DBのcategory_mediumカラムに相当
	RecipeCount  int    // この基準レシピに含まれる個人レシピの件数
	CookingTime  int    // 調理時間コード（1-6: "5分以内"〜"1時間以上"）
	AverageSteps int    // 平均手順数

	// IngredientGroups は材料をグループ（カテゴリ）ごとに分類した情報です。
	// キーはグループ名（例: "野菜", "肉類"）、値はそのグループ内の材料リスト。
	IngredientGroups map[string]StandardIngredientGroup

	// Steps は代表的な調理手順のリストです（使用頻度順）。
	Steps []StandardStep
}

// StandardIngredientGroup は1つのカテゴリに属する材料の集合情報を表します。
type StandardIngredientGroup struct {
	// TotalCount はグループ内の全材料の使用件数の合計
	TotalCount int
	// Items はグループ内の個別材料リスト（件数の多い順）
	Items []StandardIngredientItem
}

// StandardIngredientItem は基準レシピの1つの材料を表します。
// DBのstandard_recipe_ingredientsテーブルに対応します。
type StandardIngredientItem struct {
	Name  string // 材料の正規化名（normalized_name）
	Count int    // この材料を使用しているレシピの件数
}

// StandardStep は基準レシピの代表的な1手順を表します。
// DBのstandard_recipe_stepsテーブルに対応します。
type StandardStep struct {
	FoodName string // 対象食材名
	Action   string // 調理アクション（例: "切る", "炒める"）
	Count    int    // この手順を実施しているレシピの件数
}

// CookingTimeMap はDBの調理時間コード（数値）から表示用テキストへのマッピングです。
// Pythonコードの COOKING_TIME_MAP 定数に相当します。
var CookingTimeMap = map[int]string{
	1: "5分以内",
	2: "約10分",
	3: "約15分",
	4: "約30分",
	5: "約1時間",
	6: "1時間以上",
}
