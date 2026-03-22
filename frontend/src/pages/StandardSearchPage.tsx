/**
 * @file StandardSearchPage.tsx
 * @description 基準レシピ検索のトップページ。
 * URL /standard-search に対応します。
 * Pythonアプリの standard_search_home.html に相当します。
 */

import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { SearchForm } from '../components/SearchForm';
import type { SearchMode } from '../types/standardRecipe';

/**
 * StandardSearchPage は基準レシピ検索のトップページコンポーネントです。
 * レシピ名検索と材料名検索のモード切り替えをサポートします。
 */
export function StandardSearchPage() {
    const navigate = useNavigate();
    const [searchMode, setSearchMode] = useState<SearchMode>('recipe');

    /**
     * handleSearch は検索フォームの送信時に呼ばれます。
     * 検索モードもURLパラメータに含めてStandardRecipesPageへ遷移します。
     */
    const handleSearch = (query: string) => {
        navigate(
            `/standard-recipes?query=${encodeURIComponent(query)}&mode=${searchMode}`,
        );
    };

    return (
        <div>
            <h1>基準レシピ検索</h1>
            <p style={{ textAlign: 'center' }}>
                <Link to="/">← トップページに戻る</Link>
            </p>

            {/* 検索モード切り替え */}
            <div className="search-options">
                <label>
                    <input
                        type="radio"
                        name="search_mode"
                        value="recipe"
                        checked={searchMode === 'recipe'}
                        onChange={() => setSearchMode('recipe')}
                    />{' '}
                    レシピ名で検索
                </label>
                <label>
                    <input
                        type="radio"
                        name="search_mode"
                        value="ingredient"
                        checked={searchMode === 'ingredient'}
                        onChange={() => setSearchMode('ingredient')}
                    />{' '}
                    材料名で検索
                </label>
            </div>

            <p className="search-note">
                検索条件を追加して、詳細な検索ができます。<br />
                「NOT」を選択すると、その単語を含まないレシピを検索します。
            </p>

            <SearchForm onSearch={handleSearch} submitLabel="検索" />
        </div>
    );
}
