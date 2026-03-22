/**
 * @file HomePage.tsx
 * @description パーソナルレシピ検索のトップページ。
 * ルート URL (/) に対応します。
 * Pythonアプリの index.html に相当します。
 */

import { useNavigate } from 'react-router-dom';
import { SearchForm } from '../components/SearchForm';

/**
 * HomePage はパーソナルレシピ検索のトップページコンポーネントです。
 * ユーザーが検索クエリを入力すると、SearchResultsPage にナビゲートします。
 */
export function HomePage() {
    const navigate = useNavigate();

    /**
     * handleSearch は検索フォームの送信時に呼ばれるコールバックです。
     * クエリをURLパラメータに含めてSearchResultsPageへ遷移します。
     */
    const handleSearch = (query: string) => {
        navigate(`/search?query=${encodeURIComponent(query)}`);
    };

    return (
        <div>
            <h1>パーソナルレシピ検索</h1>
            <p className="search-note">
                検索条件を追加して、詳細な検索ができます。<br />
                「NOT」を選択すると、その単語を含まないレシピを検索します。
            </p>
            <SearchForm onSearch={handleSearch} submitLabel="検索" />
            <div className="extra-links">
                <p><a href="/standard-search">基準レシピを検索する</a></p>
            </div>
        </div>
    );
}
