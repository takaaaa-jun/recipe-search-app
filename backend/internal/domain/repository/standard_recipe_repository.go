// Package repository はドメイン層のリポジトリインターフェースを定義するパッケージです。
package repository

import (
	"context"

	"backend/internal/domain/entity"
)

// StandardRecipeRepository は基準レシピデータへのアクセスを抽象化するインターフェースです。
// MySQL実装は internal/infrastructure/mysql/standard_recipe_repository.go で提供します。
type StandardRecipeRepository interface {
	// FindIDsByName はカテゴリ名（category_medium）でレシピIDを検索します（レシピ名モード）。
	// AND/NOT条件をサポートし、人気順（recipe_count降順）で最大5件を返します。
	//
	// パラメータ:
	//   - ctx: コンテキスト
	//   - inclusions: 含むキーワードのリスト（LIKE検索）
	//   - exclusions: 含まないキーワードのリスト（NOT LIKE検索）
	FindIDsByName(ctx context.Context, inclusions, exclusions []string) ([]int, error)

	// FindIDsByIngredient は材料名でレシピIDを検索します（材料名モード）。
	// AND交差 + スコアリング戦略を使用して最大5件を返します。
	//
	// パラメータ:
	//   - ctx: コンテキスト
	//   - inclusions: 含む材料名のリスト (normalized_name または LIKE検索)
	//   - exclusions: 除外する材料名のリスト
	//   - normalizedNameFunc: キーワード → normalized_name 変換関数（SynonymRepositoryから渡す）
	FindIDsByIngredient(ctx context.Context, inclusions, exclusions []string) ([]int, error)

	// FindDetailByIDs は指定されたIDリストに一致する完全な基準レシピ情報を取得します。
	// 材料グループ・手順を含む完全なStandardRecipeエンティティを返します。
	// IDの順序は保持されます。
	FindDetailByIDs(ctx context.Context, ids []int) ([]entity.StandardRecipe, error)

	// FindDetailByID は指定IDの基準レシピ1件の完全詳細情報を取得します。
	// 見つからない場合は nil, nil を返します。
	FindDetailByID(ctx context.Context, id int) (*entity.StandardRecipe, error)
}
