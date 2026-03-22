/**
 * @file App.tsx
 * @description アプリケーションのルートコンポーネント。
 * React Routerを使用してページルーティングを設定します。
 */

import { BrowserRouter, Routes, Route } from 'react-router-dom';
import { HomePage } from './pages/HomePage';
import { SearchResultsPage } from './pages/SearchResultsPage';
import { RecipeDetailPage } from './pages/RecipeDetailPage';
import { StandardSearchPage } from './pages/StandardSearchPage';
import { StandardRecipesPage } from './pages/StandardRecipesPage';
import { StandardRecipeDetailPage } from './pages/StandardRecipeDetailPage';

/**
 * App はReactアプリケーションのルートコンポーネントです。
 * 全ページのルーティング定義をここで一元管理します。
 *
 * ルーティング一覧:
 *   /                                  トップページ（パーソナルレシピ検索）
 *   /search?query=〇〇               パーソナルレシピ検索結果
 *   /recipe/:id                       パーソナルレシピ詳細
 *   /standard-search                  基準レシピ検索
 *   /standard-recipes?query=〇〇     基準レシピ検索結果
 *   /standard-recipe/:id             基準レシピ詳細
 */
function App() {
  return (
    <BrowserRouter basename="/recipe-search-app">
      {/* ページ全体のコンテナ（max-widthとパディングはindex.cssで設定） */}
      <div className="page-container">
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/search" element={<SearchResultsPage />} />
          <Route path="/recipe/:id" element={<RecipeDetailPage />} />
          <Route path="/standard-search" element={<StandardSearchPage />} />
          <Route path="/standard-recipes" element={<StandardRecipesPage />} />
          <Route path="/standard-recipe/:id" element={<StandardRecipeDetailPage />} />
        </Routes>
      </div>
    </BrowserRouter>
  );
}

export default App;
