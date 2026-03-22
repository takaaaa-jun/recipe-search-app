// Package entity はアプリケーションのドメインエンティティを定義するパッケージです。
package entity

// ===========================================================================
// 栄養素関連エンティティ
// ===========================================================================

// NutritionInfo はエネルギー、たんぱく質、脂質、炭水化物、食物繊維、食塩相当量の
// 6項目を持つ汎用的な栄養素情報構造体です。
// レシピ全体の合計値、1人分の計算値のいずれにも使用します。
type NutritionInfo struct {
	EnergyKcal float64 // エネルギー (kcal)
	ProteinG   float64 // たんぱく質 (g)
	FatG       float64 // 脂質 (g)
	CarbsG     float64 // 炭水化物 (g)
	FiberG     float64 // 食物繊維 (g)
	SaltG      float64 // 食塩相当量 (g)
}

// NutritionRatio は各栄養素の基準充足率（パーセンテージ）です。
// 例えば Energy=75.0 は「エネルギーが基準値の75%」を意味します。
// フロントエンドでの栄養バーの進捗表示などに使用します。
type NutritionRatio struct {
	Energy  float64 // エネルギー充足率 (%)
	Protein float64 // たんぱく質充足率 (%)
	Fat     float64 // 脂質充足率 (%)
	Carbs   float64 // 炭水化物充足率 (%)
	Fiber   float64 // 食物繊維充足率 (%)
	Salt    float64 // 食塩相当量充足率 (%)
}

// NutritionStandards は1食分の基準摂取量（固定値）です。
// Pythonコードの STANDARDS 定数に相当します。
// フロントエンドに渡して「基準: 734kcal」のような表示に使用します。
type NutritionStandards struct {
	EnergyKcal float64 // エネルギー基準量 (kcal)
	ProteinG   float64 // たんぱく質基準量 (g)
	FatG       float64 // 脂質基準量 (g)
	CarbsG     float64 // 炭水化物基準量 (g)
	FiberG     float64 // 食物繊維基準量 (g)
	SaltG      float64 // 食塩相当量基準量 (g)
}

// DefaultNutritionStandards は食ideathonアプリで定義された基準摂取量の定数値です。
// Pythonコードの STANDARDS 定数と同じ値を使用します。
var DefaultNutritionStandards = NutritionStandards{
	EnergyKcal: 734,
	ProteinG:   31,
	FatG:       21,
	CarbsG:     106,
	FiberG:     7,
	SaltG:      2.5,
}
