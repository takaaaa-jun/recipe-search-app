/**
 * @file RecipeDetailPage.tsx
 * @description パーソナルレシピ詳細ページ。
 * URL /recipe/:id に対応します。
 * Pythonアプリの recipe_detail.html に相当します。
 */

import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getRecipeDetail } from '../api/recipes';
import type { RecipeDetail } from '../types/recipe';
import { NutritionTotals, NutritionPerServing } from '../components/NutritionDisplay';

/**
 * RecipeDetailPage はパーソナルレシピ詳細ページコンポーネントです。
 * URLパスパラメータ `:id` からレシピIDを取得し、詳細情報をAPIで取得して表示します。
 */
export function RecipeDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [recipe, setRecipe] = useState<RecipeDetail | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!id) return;
        const recipeId = parseInt(id, 10);
        if (isNaN(recipeId)) {
            setError('不正なIDです');
            setIsLoading(false);
            return;
        }

        setIsLoading(true);
        getRecipeDetail(recipeId)
            .then((data) => setRecipe(data))
            .catch((err: Error) => setError(err.message))
            .finally(() => setIsLoading(false));
    }, [id]);

    if (isLoading) return <p style={{ textAlign: 'center', marginTop: '3em' }}>読み込み中...</p>;
    if (error) return <p style={{ color: 'red', textAlign: 'center', marginTop: '3em' }}>{error}</p>;
    if (!recipe) return <p style={{ textAlign: 'center', marginTop: '3em' }}>レシピが見つかりませんでした。</p>;

    return (
        <div>
            <p>
                <a href="javascript:history.back()">← 検索結果に戻る</a>
            </p>

            <div className="result-item">
                <h2>{recipe.title}</h2>

                <div className="recipe-meta">
                    {recipe.cooking_time && <span className="time-icon">{recipe.cooking_time}</span>}
                    {recipe.serving_for && <span className="serving-icon">{recipe.serving_for}</span>}
                    <span>
                        公開日: {recipe.published_at ? new Date(recipe.published_at).toLocaleDateString('ja-JP') : '不明'}
                    </span>
                </div>

                <p>{recipe.description}</p>

                {/* 栄養素：レシピ全体合計 */}
                <NutritionTotals totals={recipe.nutrition_totals} />

                {/* 栄養素：1人分 + 充足率 */}
                <NutritionPerServing
                    perServing={recipe.nutrition_per_serving}
                    ratios={recipe.nutrition_ratios}
                    standards={recipe.standards}
                    servingSize={recipe.serving_size}
                />

                {/* 材料・手順 */}
                <div className="recipe-details">
                    <div>
                        <h3>材料</h3>
                        <ul>
                            {recipe.ingredients.length > 0 ? (
                                recipe.ingredients.map((ing) => (
                                    <li key={ing.id} className="ingredient-item">
                                        <strong>{ing.name}</strong>: {ing.quantity}
                                        <div className="ingredient-nutrition">
                                            {ing.normalized_quantity_g > 0 ? (
                                                <span className="small-note">
                                                    ({Math.round(ing.normalized_quantity_g)}g として計算)
                                                </span>
                                            ) : (
                                                <span className="small-note">[材料名不明または量不明のため計算不可]</span>
                                            )}
                                        </div>
                                    </li>
                                ))
                            ) : (
                                <li>材料情報が登録されていません。</li>
                            )}
                        </ul>
                    </div>

                    <div>
                        <h3>作り方</h3>
                        <ol>
                            {recipe.steps.length > 0 ? (
                                recipe.steps.map((step) => (
                                    <li key={step.position}>{step.memo}</li>
                                ))
                            ) : (
                                <li>手順情報が登録されていません。</li>
                            )}
                        </ol>
                    </div>
                </div>
            </div>
        </div>
    );
}
