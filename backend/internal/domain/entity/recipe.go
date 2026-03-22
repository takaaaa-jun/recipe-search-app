// Package entity はアプリケーションのドメインエンティティ（ビジネスオブジェクト）を定義するパッケージです。
// このパッケージはドメイン層の最内側に位置し、他の内部パッケージに依存しません。
// エンティティはデータの構造と振る舞いを保持し、DBやHTTPといった外部の詳細を知りません。
package entity

import "time"

// ===========================================================================
// パーソナルレシピ（個人投稿レシピ）関連エンティティ
// ===========================================================================

// Recipe はパーソナルレシピのドメインエンティティです。
// DBのrecipesテーブルを中心に、材料・手順・栄養素情報を集約します。
type Recipe struct {
	ID          int        // レシピID（PK）
	Title       string     // レシピタイトル
	Description string     // レシピの説明文
	CookingTime string     // 調理時間（表示用テキスト例: "約15分"）
	ServingFor  string     // 何人分か（表示用テキスト例: "2人前"）
	ServingSize int        // 何人分か（数値: 栄養計算で使用）
	PublishedAt *time.Time // 公開日

	Ingredients []Ingredient // 材料リスト
	Steps       []Step       // 調理手順リスト

	// NutritionTotals はレシピ全体の栄養素合計（DBのrecipe_nutrition_infoテーブルより）
	NutritionTotals NutritionInfo
	// NutritionPerServing は1人分の栄養素（NutritionTotals / ServingSize で計算）
	NutritionPerServing NutritionInfo
	// NutritionRatios は基準摂取量に対する充足率（パーセンテージ）
	NutritionRatios NutritionRatio
	// Standards は栄養素の基準摂取量（フロントエンドへの表示用）
	Standards NutritionStandards
}

// RecipeSummary は検索結果一覧表示用の軽量なレシピ情報です。
// 詳細データ（materials, steps, nutritionなど）は含みません。
type RecipeSummary struct {
	ID          int        // レシピID
	Title       string     // レシピタイトル
	Description string     // レシピの説明文（先頭20文字程度で切り捨て可）
	PublishedAt *time.Time // 公開日
}

// Ingredient は1つの材料情報を表すエンティティです。
// DBのingredientsテーブルとingredient_structured / ingredient_unitsを結合した情報を保持します。
type Ingredient struct {
	ID       int    // 材料ID（DBのingredients.id）
	Name     string // 材料名（表示名）
	Quantity string // 分量（表示用テキスト例: "大さじ1"）
	// NormalizedQuantityG は栄養計算に使用するグラム換算量（ingredient_units.normalized_quantity）
	NormalizedQuantityG float64
	// Nutrition は各材料の栄養素情報（nutritionsテーブルのデータと量から計算）
	Nutrition IngredientNutrition
}

// IngredientNutrition は1つの材料の栄養素情報です（実量ベースで計算済み）。
type IngredientNutrition struct {
	EnergyKcal float64 // エネルギー (kcal)
	ProteinG   float64 // たんぱく質 (g)
	FatG       float64 // 脂質 (g)
	CarbsG     float64 // 炭水化物 (g)
}

// Step はレシピの1ステップ（調理手順の1行）を表します。
// DBのstepsテーブルに対応します。
type Step struct {
	Position int    // 手順の順序（1始まり）
	Memo     string // 手順の説明文
}
