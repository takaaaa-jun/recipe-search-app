/**
 * @file SearchResultsPage.tsx
 * @description パーソナルレシピ検索結果ページ。
 * URL /search?query=〇〇 に対応します。
 * Pythonアプリの results.html に相当します。
 */

import { useEffect, useState } from 'react';
import { useSearchParams, useNavigate, Link } from 'react-router-dom';
import { searchRecipes } from '../api/recipes';
import type { RecipeSummary } from '../types/recipe';
import { SearchForm } from '../components/SearchForm';

/**
 * SearchResultsPage はパーソナルレシピ検索結果ページコンポーネントです。
 * URLクエリパラメータからクエリを読み取り、APIを呼び出して結果を表示します。
 */
export function SearchResultsPage() {
    const [searchParams] = useSearchParams();
    const navigate = useNavigate();
    const query = searchParams.get('query') ?? '';

    const [recipes, setRecipes] = useState<RecipeSummary[]>([]);
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);

    /** クエリが変わるたびにAPIを呼び出す */
    useEffect(() => {
        if (!query) return;

        setIsLoading(true);
        setError(null);

        searchRecipes(query)
            .then((res) => setRecipes(res.recipes))
            .catch((err: Error) => setError(err.message))
            .finally(() => setIsLoading(false));
    }, [query]);

    /**
     * handleSearch は再検索フォームの送信時に呼ばれます。
     * 新しいクエリでURLを更新しnavigate します。
     */
    const handleSearch = (newQuery: string) => {
        navigate(`/search?query=${encodeURIComponent(newQuery)}`);
    };

    return (
        <div>
            <h1>検索結果: "{query}"</h1>
            <p style={{ textAlign: 'center' }}>
                <Link to="/">← トップページに戻る</Link>
            </p>

            {/* 再検索フォーム（既存クエリをタグ表示） */}
            <SearchForm
                onSearch={handleSearch}
                initialQuery={query}
                submitLabel="再検索"
            />

            {/* ローディング表示 */}
            {isLoading && <p style={{ textAlign: 'center' }}>検索中...</p>}

            {/* エラー表示 */}
            {error && <p style={{ color: 'red', textAlign: 'center' }}>{error}</p>}

            {/* 検索結果 */}
            {!isLoading && recipes.length > 0 && (
                <>
                    <p style={{ textAlign: 'center' }}>検索結果 (ランダム開始: 10件表示)</p>
                    {recipes.map((recipe) => (
                        <div key={recipe.id} className="result-item">
                            <h2>
                                <Link to={`/recipe/${recipe.id}`}>{recipe.title}</Link>
                            </h2>
                            <div className="recipe-meta">
                                <span>公開日: {recipe.published_at ? new Date(recipe.published_at).toLocaleDateString('ja-JP') : '不明'}</span>
                            </div>
                            <p>{recipe.description}</p>
                        </div>
                    ))}
                </>
            )}

            {/* 結果なし */}
            {!isLoading && !error && query && recipes.length === 0 && (
                <p style={{ textAlign: 'center' }}>該当するレシピは見つかりませんでした。</p>
            )}
        </div>
    );
}
