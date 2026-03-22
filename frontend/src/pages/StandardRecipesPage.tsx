/**
 * @file StandardRecipesPage.tsx
 * @description 基準レシピ検索結果ページ。
 * URL /standard-recipes?query=〇〇&mode=recipe に対応します。
 * Pythonアプリの standard_recipes.html に相当します。
 */

import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { searchStandardRecipes } from '../api/standardRecipes';
import type { StandardRecipe, SearchMode } from '../types/standardRecipe';

/**
 * StandardRecipesPage は基準レシピ検索結果ページコンポーネントです。
 */
export function StandardRecipesPage() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const query = searchParams.get('query') ?? '';
    const mode = (searchParams.get('mode') ?? 'recipe') as SearchMode;

    const [recipes, setRecipes] = useState<StandardRecipe[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        if (!query) return;
        setIsLoading(true);
        searchStandardRecipes(query, mode)
            .then((res) => setRecipes(res.recipes))
            .catch((err: Error) => setError(err.message))
            .finally(() => setIsLoading(false));
    }, [query, mode]);

    const handleReSearch = () => {
        navigate('/standard-search');
    };

    return (
        <div>
            <h1>基準レシピ検索結果: "{query}"</h1>
            <p style={{ textAlign: 'center' }}>
                <Link to="/standard-search">← 検索ページに戻る</Link>
            </p>

            <div style={{ textAlign: 'center', marginBottom: '1.5em' }}>
                <button className="submit-button" onClick={handleReSearch}>
                    検索し直す
                </button>
            </div>

            {isLoading && <p style={{ textAlign: 'center' }}>検索中...</p>}
            {error && <p style={{ color: 'red', textAlign: 'center' }}>{error}</p>}

            {!isLoading && recipes.length > 0 && (
                <>
                    {recipes.map((recipe) => (
                        <div key={recipe.id} className="result-item">
                            <h2>
                                <Link to={`/standard-recipe/${recipe.id}`}>{recipe.name}</Link>
                            </h2>
                            <div className="recipe-meta">
                                <span>レシピ数: {recipe.recipe_count}件</span>
                                {recipe.cooking_time_label && (
                                    <span>調理時間: {recipe.cooking_time_label}</span>
                                )}
                                <span>平均手順数: {recipe.average_steps}</span>
                            </div>

                            {/* 材料グループの表示 */}
                            {recipe.ingredient_groups.length > 0 && (
                                <div style={{ marginTop: '1em' }}>
                                    <h3>代表的な材料</h3>
                                    {recipe.ingredient_groups
                                        .sort((a, b) => b.total_count - a.total_count)
                                        .slice(0, 3)
                                        .map((grp) => (
                                            <div key={grp.group_name} style={{ marginBottom: '0.5em' }}>
                                                <strong>{grp.group_name}</strong>:{'  '}
                                                {grp.items.slice(0, 5).map((item, i) => (
                                                    <span key={item.name}>
                                                        {item.name}({item.count}){i < Math.min(grp.items.length, 5) - 1 ? '、' : ''}
                                                    </span>
                                                ))}
                                            </div>
                                        ))}
                                </div>
                            )}
                        </div>
                    ))}
                </>
            )}

            {!isLoading && !error && query && recipes.length === 0 && (
                <p style={{ textAlign: 'center' }}>該当する基準レシピは見つかりませんでした。</p>
            )}
        </div>
    );
}
