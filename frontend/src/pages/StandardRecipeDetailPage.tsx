/**
 * @file StandardRecipeDetailPage.tsx
 * @description 基準レシピ詳細ページ。
 * URL /standard-recipe/:id に対応します。
 * Pythonアプリの standard_recipe_detail.html に相当します。
 */

import { useEffect, useState } from 'react';
import { useParams } from 'react-router-dom';
import { getStandardRecipeDetail } from '../api/standardRecipes';
import type { StandardRecipe } from '../types/standardRecipe';

/**
 * StandardRecipeDetailPage は基準レシピ詳細ページコンポーネントです。
 */
export function StandardRecipeDetailPage() {
    const { id } = useParams<{ id: string }>();
    const [recipe, setRecipe] = useState<StandardRecipe | null>(null);
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
        getStandardRecipeDetail(recipeId)
            .then((data) => setRecipe(data))
            .catch((err: Error) => setError(err.message))
            .finally(() => setIsLoading(false));
    }, [id]);

    if (isLoading) return <p style={{ textAlign: 'center', marginTop: '3em' }}>読み込み中...</p>;
    if (error) return <p style={{ color: 'red', textAlign: 'center', marginTop: '3em' }}>{error}</p>;
    if (!recipe) return <p style={{ textAlign: 'center', marginTop: '3em' }}>基準レシピが見つかりませんでした。</p>;

    return (
        <div>
            <p>
                <a href="javascript:history.back()">← 検索結果に戻る</a>
            </p>

            <div className="result-item">
                <h2>{recipe.name}</h2>

                <div className="recipe-meta">
                    <span>レシピ数: {recipe.recipe_count}件</span>
                    {recipe.cooking_time_label && <span>調理時間: {recipe.cooking_time_label}</span>}
                    <span>平均手順数: {recipe.average_steps}</span>
                </div>

                {/* 材料グループ */}
                <div className="recipe-details">
                    <div>
                        <h3>代表的な材料</h3>
                        {recipe.ingredient_groups
                            .sort((a, b) => b.total_count - a.total_count)
                            .map((grp) => (
                                <div key={grp.group_name} style={{ marginBottom: '1em' }}>
                                    <h4 style={{ margin: '0 0 0.5em 0', color: '#555' }}>{grp.group_name}</h4>
                                    <ul style={{ paddingLeft: '1.5em' }}>
                                        {grp.items
                                            .sort((a, b) => b.count - a.count)
                                            .map((item) => (
                                                <li key={item.name} className="ingredient-item">
                                                    <strong>{item.name}</strong>{' '}
                                                    <span className="small-note">({item.count}件のレシピで使用)</span>
                                                </li>
                                            ))}
                                    </ul>
                                </div>
                            ))}
                    </div>

                    {/* 代表的な手順 */}
                    <div>
                        <h3>代表的な調理手順</h3>
                        <ul>
                            {recipe.steps.map((step, index) => (
                                <li key={index}>
                                    <strong>{step.food_name}</strong> を <strong>{step.action}</strong>
                                    <span className="small-note"> ({step.count}件のレシピで使用)</span>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>
            </div>
        </div>
    );
}
