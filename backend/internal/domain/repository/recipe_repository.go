// Package repository はドメイン層のリポジトリインターフェースを定義するパッケージです。
// インターフェースのみを定義し、DBへの具体的なアクセス方法（MySQL実装など）は
// infrastructure層に委ねます。これによりドメイン層はDB技術の詳細に依存しません。
package repository

import (
	"context"

	"backend/internal/domain/entity"
)

// RecipeRepository はパーソナルレシピデータへのアクセスを抽象化するインターフェースです。
// MySQL実装は internal/infrastructure/mysql/recipe_repository.go で提供します。
type RecipeRepository interface {
	// FindIDsByIngredientSynonyms は食材の同義語グループから条件に一致するレシピIDを検索します。
	//
	// パラメータ:
	//   - ctx: コンテキスト（タイムアウト・キャンセル制御）
	//   - synonymGroups: [[同義語A,同義語A'], [同義語B,同義語B']] のような2次元配列
	//                    外側の配列要素はAND条件、内側の配列要素はOR条件（同義語展開）
	//   - startID:       カーソルページネーション用の開始レシピID（ランダム開始点）
	//   - limit:         取得する最大件数（通常は10）
	//
	// 戻り値: 条件に一致したレシピIDのスライス
	FindIDsByIngredientSynonyms(ctx context.Context, synonymGroups [][]string, startID, limit int) ([]int, error)

	// FindDetailByID は指定のIDに対応するレシピの完全な詳細情報を取得します。
	// materials, steps, nutrition情報を含む完全なRecipeエンティティを返します。
	// 見つからない場合は nil, nil を返します。
	FindDetailByID(ctx context.Context, id int) (*entity.Recipe, error)

	// FindSummariesByIDs は指定されたIDリストに一致するレシピサマリーリストを取得します。
	// IDの順序は保持されます（FIELD関数使用）。
	FindSummariesByIDs(ctx context.Context, ids []int) ([]entity.RecipeSummary, error)
}

// SynonymRepository は同義語辞書（synonym_dictionaryテーブル）へのアクセスを
// 抽象化するインターフェースです。
type SynonymRepository interface {
	// GetSynonyms は指定されたキーワードの同義語を全て取得します。
	// キーワード自身も戻り値に含まれます。
	// 例: "鶏肉" → ["鶏肉", "チキン", "とりにく"]
	GetSynonyms(ctx context.Context, keyword string) ([]string, error)

	// UnifyKeywords は入力キーワードリスト内の同義語グループを統一します。
	// 同一の normalized_name を持つキーワードが複数ある場合、
	// DBでIDが最も小さいものを代表として1つに集約します。
	// 検索前のクエリ正規化に使用します。
	UnifyKeywords(ctx context.Context, keywords []string) ([]string, error)

	// GetNormalizedName はキーワードに対応する正規化名を取得します。
	// 基準レシピの材料名検索で使用します。
	GetNormalizedName(ctx context.Context, keyword string) (string, error)
}
