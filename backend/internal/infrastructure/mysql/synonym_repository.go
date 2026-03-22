// Package mysql はドメイン層のリポジトリインターフェースに対するMySQL実装を提供します。
package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// SynonymRepository はSynonymRepositoryインターフェースのMySQL実装です。
// synonym_dictionaryテーブルを使って同義語の展開・統合を行います。
type SynonymRepository struct {
	db *sql.DB
}

// NewSynonymRepository はSynonymRepositoryの新しいインスタンスを生成します。
func NewSynonymRepository(db *sql.DB) *SynonymRepository {
	return &SynonymRepository{db: db}
}

// GetSynonyms は指定されたキーワードの全同義語を取得します。
// キーワード自身も戻り値に含まれます。
// Pythonの get_synonyms() 関数に相当します。
func (r *SynonymRepository) GetSynonyms(ctx context.Context, keyword string) ([]string, error) {
	synonymSet := map[string]struct{}{keyword: {}}

	// Step 1: キーワードが normalized_name である場合、その synonym を全取得
	rows1, err := r.db.QueryContext(ctx,
		"SELECT synonym FROM synonym_dictionary WHERE normalized_name = ?", keyword)
	if err != nil {
		return nil, fmt.Errorf("GetSynonyms step1 query failed (keyword=%s): %w", keyword, err)
	}
	defer rows1.Close()
	for rows1.Next() {
		var syn string
		if err := rows1.Scan(&syn); err != nil {
			return nil, err
		}
		synonymSet[syn] = struct{}{}
	}

	// Step 2: キーワードが synonym である場合、normalized_name を取得し、
	//         さらにその normalized_name に紐づく全 synonym を取得
	rows2, err := r.db.QueryContext(ctx,
		"SELECT normalized_name FROM synonym_dictionary WHERE synonym = ?", keyword)
	if err != nil {
		return nil, fmt.Errorf("GetSynonyms step2 query failed: %w", err)
	}
	defer rows2.Close()

	normalizedNames := []string{}
	for rows2.Next() {
		var norm string
		if err := rows2.Scan(&norm); err != nil {
			return nil, err
		}
		normalizedNames = append(normalizedNames, norm)
	}

	for _, norm := range normalizedNames {
		synonymSet[norm] = struct{}{}
		// その normalized_name の全 synonym も追加
		rows3, err := r.db.QueryContext(ctx,
			"SELECT synonym FROM synonym_dictionary WHERE normalized_name = ?", norm)
		if err != nil {
			return nil, fmt.Errorf("GetSynonyms step3 query failed (norm=%s): %w", norm, err)
		}
		defer rows3.Close()
		for rows3.Next() {
			var syn string
			if err := rows3.Scan(&syn); err != nil {
				return nil, err
			}
			synonymSet[syn] = struct{}{}
		}
	}

	// setをスライスに変換
	result := make([]string, 0, len(synonymSet))
	for syn := range synonymSet {
		result = append(result, syn)
	}
	return result, nil
}

// UnifyKeywords は入力キーワードリスト内の同義語グループを統合します。
// 同一の normalized_name を持つ複数のキーワードがある場合、DBでIDが最小のものを代表として採用します。
// Pythonの unify_keywords() 関数に相当します。
func (r *SynonymRepository) UnifyKeywords(ctx context.Context, keywords []string) ([]string, error) {
	if len(keywords) == 0 {
		return nil, nil
	}

	// Step 1: 各 keyword → normalized_name のマッピングを取得
	placeholders := strings.Repeat("?,", len(keywords))
	placeholders = placeholders[:len(placeholders)-1]
	args := make([]interface{}, len(keywords))
	for i, kw := range keywords {
		args[i] = kw
	}

	rows, err := r.db.QueryContext(ctx,
		fmt.Sprintf("SELECT synonym, normalized_name FROM synonym_dictionary WHERE synonym IN (%s)", placeholders),
		args...)
	if err != nil {
		return nil, fmt.Errorf("UnifyKeywords step1 query failed: %w", err)
	}
	defer rows.Close()

	// keyword → normalized_name のマップ
	kwToNorm := map[string]string{}
	for rows.Next() {
		var synonym, normalizedName string
		if err := rows.Scan(&synonym, &normalizedName); err != nil {
			return nil, err
		}
		kwToNorm[synonym] = normalizedName
	}

	// Step 2: 登場した normalized_name を収集
	seenNorms := map[string]struct{}{}
	for _, kw := range keywords {
		if norm, ok := kwToNorm[kw]; ok {
			seenNorms[norm] = struct{}{}
		}
	}

	// Step 3: 各 normalized_name についてIDが最小の synonym を取得
	normToBest := map[string]string{}
	if len(seenNorms) > 0 {
		normList := make([]string, 0, len(seenNorms))
		for norm := range seenNorms {
			normList = append(normList, norm)
		}
		normPlaceholders := strings.Repeat("?,", len(normList))
		normPlaceholders = normPlaceholders[:len(normPlaceholders)-1]
		normArgs := make([]interface{}, len(normList))
		for i, norm := range normList {
			normArgs[i] = norm
		}

		bestRows, err := r.db.QueryContext(ctx,
			fmt.Sprintf("SELECT normalized_name, synonym, id FROM synonym_dictionary WHERE normalized_name IN (%s) ORDER BY id ASC", normPlaceholders),
			normArgs...)
		if err != nil {
			return nil, fmt.Errorf("UnifyKeywords step3 query failed: %w", err)
		}
		defer bestRows.Close()

		for bestRows.Next() {
			var norm, syn string
			var id int
			if err := bestRows.Scan(&norm, &syn, &id); err != nil {
				return nil, err
			}
			// ORDER BY id ASC なので、最初に現れたもの（IDが最小）を採用
			if _, exists := normToBest[norm]; !exists {
				normToBest[norm] = syn
			}
		}
	}

	// Step 4: 入力順序を維持しつつ、統合されたキーワードリストを構築
	unifiedKeywords := []string{}
	processedNorms := map[string]struct{}{}

	for _, kw := range keywords {
		if norm, ok := kwToNorm[kw]; ok {
			// DBに存在するキーワード: normalized_name の代表に置換
			if _, processed := processedNorms[norm]; !processed {
				if best, hasBest := normToBest[norm]; hasBest {
					unifiedKeywords = append(unifiedKeywords, best)
				}
				processedNorms[norm] = struct{}{}
			}
		} else {
			// DBにないキーワードはそのまま使用
			unifiedKeywords = append(unifiedKeywords, kw)
		}
	}
	return unifiedKeywords, nil
}

// GetNormalizedName はキーワードに対応する normalized_name を取得します。
// Pythonの get_normalized_name() 関数に相当します。
func (r *SynonymRepository) GetNormalizedName(ctx context.Context, keyword string) (string, error) {
	// Step 1: キーワードが既に normalized_name として存在するか確認
	var norm string
	err := r.db.QueryRowContext(ctx,
		"SELECT normalized_name FROM synonym_dictionary WHERE normalized_name = ? LIMIT 1", keyword).
		Scan(&norm)
	if err == nil {
		return keyword, nil // 既に正規化名
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("GetNormalizedName step1 query failed: %w", err)
	}

	// Step 2: synonym として検索し、対応する normalized_name を取得
	err = r.db.QueryRowContext(ctx,
		"SELECT normalized_name FROM synonym_dictionary WHERE synonym = ? LIMIT 1", keyword).
		Scan(&norm)
	if err == sql.ErrNoRows {
		return "", nil // 見つからない場合は空文字列
	}
	if err != nil {
		return "", fmt.Errorf("GetNormalizedName step2 query failed: %w", err)
	}
	return norm, nil
}
