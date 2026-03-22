/**
 * @file standardRecipe.ts
 * @description 基準レシピ関連のTypeScript型定義。
 * バックエンドのGoエンティティおよびDTOに対応する型を定義します。
 */

// ===========================================================================
// 基準レシピ関連の型定義
// ===========================================================================

/**
 * StandardIngredientItem は基準レシピの1つの材料を表す型です。
 */
export interface StandardIngredientItem {
    name: string;  // 材料名（正規化名）
    count: number; // 使用レシピ件数
}

/**
 * StandardIngredientGroup は1つのカテゴリ（グループ）に属する材料の集合を表す型です。
 */
export interface StandardIngredientGroup {
    group_name: string;               // グループ名（例: "野菜", "肉類"）
    total_count: number;              // グループ全体の合計使用件数
    items: StandardIngredientItem[];  // グループ内の材料リスト
}

/**
 * StandardStep は基準レシピの代表的な1手順を表す型です。
 */
export interface StandardStep {
    food_name: string; // 対象食材名
    action: string;    // 調理アクション（例: "切る", "炒める"）
    count: number;     // この手順を実施しているレシピ件数
}

/**
 * StandardRecipe は基準レシピの完全な情報を表す型です。
 * バックエンドの StandardRecipeResponse DTO に対応します。
 */
export interface StandardRecipe {
    id: number;
    name: string;             // カテゴリ名（例: "野菜炒め"）
    recipe_count: number;     // 個人レシピ件数
    cooking_time_label: string; // 調理時間テキスト（例: "約30分"）
    average_steps: number;    // 平均手順数
    ingredient_groups: StandardIngredientGroup[]; // 材料グループリスト
    steps: StandardStep[];    // 代表的な手順リスト（件数順）
}

/**
 * SearchStandardRecipesResponse は基準レシピ検索APIのレスポンス型です。
 */
export interface SearchStandardRecipesResponse {
    recipes: StandardRecipe[];
}

/**
 * SearchMode は基準レシピの検索モードを表す型です。
 */
export type SearchMode = 'recipe' | 'ingredient';
