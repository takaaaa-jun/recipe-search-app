/**
 * @file standardRecipes.ts
 * @description 基準レシピAPIのクライアント関数。
 * バックエンドの /api/standard-recipes/* エンドポイントへのHTTPリクエストを担当します。
 */

import type { SearchStandardRecipesResponse, StandardRecipe, SearchMode } from '../types/standardRecipe';

const API_BASE = '/api';

/**
 * searchStandardRecipes は基準レシピ検索APIを呼び出します。
 *
 * @param query - 検索クエリ文字列（スペース区切りでAND、'-' プレフィックスでNOT）
 * @param mode - 検索モード（'recipe': レシピ名検索 | 'ingredient': 材料名検索）
 * @returns 検索結果の基準レシピリスト（最大5件）
 * @throws APIエラー時はErrorをthrowします
 */
export async function searchStandardRecipes(
    query: string,
    mode: SearchMode,
): Promise<SearchStandardRecipesResponse> {
    const response = await fetch(`${API_BASE}/standard-recipes/search`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query, search_mode: mode }),
    });
    if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(error.error ?? `HTTP error: ${response.status}`);
    }
    return response.json() as Promise<SearchStandardRecipesResponse>;
}

/**
 * getStandardRecipeDetail は指定IDの基準レシピ詳細を取得します。
 *
 * @param id - 基準レシピID
 * @returns 基準レシピの完全詳細情報
 * @throws 見つからない場合や通信エラー時はErrorをthrowします
 */
export async function getStandardRecipeDetail(id: number): Promise<StandardRecipe> {
    const response = await fetch(`${API_BASE}/standard-recipes/${id}`);
    if (!response.ok) {
        if (response.status === 404) {
            throw new Error('基準レシピが見つかりませんでした');
        }
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(error.error ?? `HTTP error: ${response.status}`);
    }
    return response.json() as Promise<StandardRecipe>;
}
