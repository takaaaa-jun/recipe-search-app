/**
 * @file recipes.ts
 * @description パーソナルレシピAPIのクライアント関数。
 * バックエンドの /api/recipes/* エンドポイントへのHTTPリクエストを担当します。
 */

import type { SearchResponse, RecipeDetail } from '../types/recipe';

/** バックエンドAPIのベースURL。Viteのプロキシ設定で /api がバックエンドに転送される */
const API_BASE = '/api';

/**
 * searchRecipes はパーソナルレシピ検索APIを呼び出します。
 *
 * @param query - 検索クエリ文字列（スペース区切りでAND、'-' プレフィックスでNOT）
 * @param startId - カーソルページネーションの開始ID（省略時はサーバーでランダム生成）
 * @returns 検索結果のレシピサマリーリスト（最大10件）
 * @throws APIエラー時はErrorをthrowします
 */
export async function searchRecipes(query: string, startId?: number): Promise<SearchResponse> {
    const params = new URLSearchParams({ query });
    if (startId !== undefined) {
        params.set('start_id', String(startId));
    }

    const response = await fetch(`${API_BASE}/recipes/search?${params.toString()}`);
    if (!response.ok) {
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(error.error ?? `HTTP error: ${response.status}`);
    }
    return response.json() as Promise<SearchResponse>;
}

/**
 * getRecipeDetail は指定IDのパーソナルレシピ詳細を取得します。
 *
 * @param id - レシピID
 * @returns レシピの完全詳細情報（材料・手順・栄養素含む）
 * @throws 見つからない場合や通信エラー時はErrorをthrowします
 */
export async function getRecipeDetail(id: number): Promise<RecipeDetail> {
    const response = await fetch(`${API_BASE}/recipes/${id}`);
    if (!response.ok) {
        if (response.status === 404) {
            throw new Error('レシピが見つかりませんでした');
        }
        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(error.error ?? `HTTP error: ${response.status}`);
    }
    return response.json() as Promise<RecipeDetail>;
}
