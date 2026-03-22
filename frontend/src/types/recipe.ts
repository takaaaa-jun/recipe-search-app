/**
 * @file recipe.ts
 * @description パーソナルレシピ関連のTypeScript型定義。
 * バックエンドのGoエンティティおよびDTOに対応する型を定義します。
 */

// ===========================================================================
// パーソナルレシピ関連の型定義
// ===========================================================================

/**
 * RecipeSummary は検索結果一覧の1件を表す型です。
 * バックエンドの RecipeSummaryResponse DTO に対応します。
 */
export interface RecipeSummary {
  id: number;
  title: string;
  description: string; // 先頭20文字程度
  published_at: string; // ISO 8601形式
}

/**
 * SearchResponse はレシピ検索APIのレスポンス型です。
 */
export interface SearchResponse {
  recipes: RecipeSummary[];
}

/**
 * NutritionInfo はエネルギー・たんぱく質・脂質などの6項目の栄養素情報を表す型です。
 */
export interface NutritionInfo {
  energy_kcal: number;
  protein_g: number;
  fat_g: number;
  carbs_g: number;
  fiber_g: number;
  salt_g: number;
}

/**
 * NutritionRatio は各栄養素の基準充足率（%）を表す型です。
 */
export interface NutritionRatio {
  energy: number;
  protein: number;
  fat: number;
  carbs: number;
  fiber: number;
  salt: number;
}

/**
 * NutritionStandards は1食分の栄養素基準摂取量を表す型です。
 */
export interface NutritionStandards {
  energy_kcal: number;
  protein_g: number;
  fat_g: number;
  carbs_g: number;
  fiber_g: number;
  salt_g: number;
}

/**
 * Ingredient は1つの材料情報を表す型です。
 */
export interface Ingredient {
  id: number;
  name: string;
  quantity: string;       // 表示用（例: "大さじ1"）
  normalized_quantity_g: number; // グラム換算量
}

/**
 * Step は調理手順の1ステップを表す型です。
 */
export interface Step {
  position: number;
  memo: string;
}

/**
 * RecipeDetail はレシピ詳細APIのレスポンス型です。
 * バックエンドの RecipeDetailResponse DTO に対応します。
 */
export interface RecipeDetail {
  id: number;
  title: string;
  description: string;
  cooking_time: string;   // 表示用テキスト（例: "約15分"）
  serving_for: string;    // 表示用テキスト（例: "2人前"）
  serving_size: number;   // 数値（栄養計算用）
  published_at: string;   // ISO 8601形式
  ingredients: Ingredient[];
  steps: Step[];
  nutrition_totals: NutritionInfo;
  nutrition_per_serving: NutritionInfo;
  nutrition_ratios: NutritionRatio;
  standards: NutritionStandards;
}
