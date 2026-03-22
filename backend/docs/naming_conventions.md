# 命名規則・コーディングスタイルガイド

> このドキュメントはコードレビューやPRにおける基準として機能します。
> 一貫した命名規則は認知負荷を下げ、チームが素早くコードを理解することを助けます。

---

## バックエンド (Go)

### パッケージ名
- **小文字・単語区切りなし** を使用する（Goの慣習）
- 例: `entity`, `repository`, `handler`, `router`, `mysql`
- ❌ `recipeEntity`, `recipe_entity`

### ディレクトリ構造（クリーンアーキテクチャ）

```
内側（依存なし）  ←  外側（内側に依存）
domain → usecase → adapter → infrastructure
```

各層のパッケージが依存できる方向は「外から内」のみ。
`adapter/` は `usecase/` に依存できるが、`usecase/` が `adapter/` に依存してはいけない。

### ファイル名
- **スネークケース** を使用する
- 例: `recipe_repository.go`, `personal_recipe_handler.go`

### 型名・構造体名・インターフェース名
- **パスカルケース（UpperCamelCase）** を使用する
- インターフェース名は `-er` サフィックス（Go慣習）または `Repository` サフィックス
- 例: `RecipeRepository`, `SynonymRepository`, `PersonalRecipeHandler`

```go
// ✅ 良い例: パスカルケース + 役割が明確
type RecipeRepository interface { ... }
type SearchPersonalRecipeUsecase struct { ... }

// ❌ 悪い例
type recipeRepository interface { ... }
type SearchUC struct { ... }  // 略語は使わない
```

### 変数名・フィールド名
- **キャメルケース（lowerCamelCase）** を使用する
- 単一文字変数は `i`, `j`（ループインデックス）、`r`（http.Request等）のみ許可
- 短縮形は避ける（可読性を優先）

```go
// ✅ 良い例: 役割が自明
recipeID := 42
synonymGroups := [][]string{}
foundIDs := []int{}

// ❌ 悪い例: 短縮しすぎ
rID := 42
sG := [][]string{}
```

### 定数
- Go標準の **パスカルケース** を使用する
- 例: `DefaultNutritionStandards`, `SearchModeRecipe`, `SearchModeIngredient`

```go
// ✅ 良い例
const SearchModeIngredient SearchMode = "ingredient"

// ❌ 悪い例 (C/Python スタイル)
const SEARCH_MODE_INGREDIENT = "ingredient"
```

### 関数・メソッド名
- **パスカルケース（公開）** または **キャメルケース（非公開）** を使用する
- 動詞から始める。何を「する」かが明確に分かる名前をつける

| 種別 | 命名パターン | 例 |
|------|------------|-----|
| 取得 | `Find` / `Get` | `FindIDsByIngredient`, `GetSynonyms` |
| 実行 | `Execute` | `Execute(ctx, input)` |
| 変換 | `to〇〇` | `toStandardRecipeResponse` |
| 生成 | `New〇〇` | `NewRecipeRepository` |
| 分割/処理 | 動詞+目的語 | `parseQuery`, `deduplicateAndSort` |

```go
// ✅ 良い例: 動詞から始まり、何をするか明確
func (r *RecipeRepository) FindIDsByIngredientSynonyms(...) {}
func (uc *SearchPersonalRecipeUsecase) Execute(...) {}

// ❌ 悪い例
func (r *RecipeRepository) RecipeIDs(...) {}  // 動詞なし
func (r *RecipeRepository) Fetch(...) {}      // 何をFetchするか不明
```

### エラーハンドリング
- エラーは必ず呼び出し元に返す（`log.Fatal` は `main.go` 内のみ）
- エラーメッセージは日本語を避け、英語でコンテキストを付与する（`%w` でラップ）

```go
// ✅ 良い例: fmt.Errorf + %w でコンテキストを追加
if err != nil {
    return nil, fmt.Errorf("FindIDsByIngredient query failed (keyword=%s): %w", kw, err)
}

// ❌ 悪い例: コンテキストなし
if err != nil {
    return nil, err
}
```

### コメント
- 公開型・公開関数には **Godocコメント** を必ず記述する
- ファイル冒頭に `// Package 〇〇 は...` の形式でパッケージコメントを記述
- 複雑なアルゴリズムには実装の意図をコメントで説明する

```go
// ✅ 良い例: Godoc形式 + 目的の説明
// FindIDsByIngredientSynonyms は食材の同義語グループから条件に一致するレシピIDを検索します。
// アルゴリズム: 単一グループはScatter-Gather、複数グループはレアリティファースト戦略を使用。
func (r *RecipeRepository) FindIDsByIngredientSynonyms(...) {}
```

### DTOと型変換
- ドメインエンティティはAPIレスポンスに直接使わず、専用のDTO構造体に変換する
- 変換関数の命名は `to〇〇Response` とする

```go
// ✅ 良い例: エンティティとDTOを分離
type StandardRecipeResponse struct { ... }  // DTO（アダプター層）
type StandardRecipe struct { ... }          // エンティティ（ドメイン層）

func toStandardRecipeResponse(r entity.StandardRecipe) StandardRecipeResponse { ... }
```

---

## フロントエンド (TypeScript + React)

### ファイル名・コンポーネント名
- **コンポーネントファイル**: パスカルケース + `.tsx` 拡張子
- **ユーティリティ・APIファイル**: キャメルケース + `.ts` 拡張子

```
RecipeCard.tsx      ← コンポーネント
SearchForm.tsx      ← コンポーネント
recipes.ts          ← API関数
standardRecipes.ts  ← API関数
useUserId.ts        ← カスタムフック
```

### TypeScript 型・インターフェース名
- **パスカルケース** を使用する
- `I` プレフィックスは使用しない（Goと同様）

```typescript
// ✅ 良い例
interface RecipeSummary { ... }
interface NutritionInfo { ... }

// ❌ 悪い例
interface IRecipeSummary { ... }
```

### 変数・関数名
- **キャメルケース** を使用する
- 関数名は動詞から始める

```typescript
// ✅ 良い例
const searchRecipes = async (query: string): Promise<RecipeSummary[]> => { ... }
const getRecipeDetail = async (id: number): Promise<RecipeDetail> => { ... }

// ❌ 悪い例
const recipes = async () => { ... }  // 動詞なし
```

### React コンポーネントのprops型名
- `〇〇Props` の命名パターンを使用する

```typescript
// ✅ 良い例
interface SearchFormProps {
  onSearch: (query: string) => void;
  initialQuery?: string;
}
```

### APIエラーハンドリング
- API呼び出しは `try/catch` でラップし、エラー状態をReact stateで管理する
- エラーメッセージはユーザーに分かりやすい日本語で表示する

---

## 共通ルール

### 略語の扱い
よく使う略語は以下の表記で統一する（全て大文字 or 全て小文字）:

| 略語 | Goでの表記 | TypeScriptでの表記 |
|------|-----------|-------------------|
| ID | `recipeID`, `standardRecipeID` | `recipeId`, `standardRecipeId` |
| URL | `requestURL` | `requestUrl` |
| API | `APIServer` | `apiClient` |

### マジックナンバーの排除
数値リテラルは定数や変数に抽出して命名する:

```go
// ✅ 良い例
const maxSearchResults = 10
const maxScanCandidates = 10000

// ❌ 悪い例
limit := 10
scanned < 10000
```
