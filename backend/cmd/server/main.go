// Package main はアプリケーションのエントリポイントです。
// このファイルは「コンポジションルート」として機能し、全ての依存関係を構築（DI）し、
// サーバーを起動する責務のみを持ちます。
//
// 依存関係の構築順序:
//   1. インフラ層: DBコネクション
//   2. インフラ層: リポジトリ実装（MySQL）
//   3. ユースケース層: ユースケースにリポジトリを注入
//   4. アダプター層: ハンドラーにユースケースを注入
//   5. アダプター層: ルーターにハンドラーを注入
//   6. サーバー起動
package main

import (
	"log"
	"os"

	"backend/internal/adapter/handler"
	"backend/internal/adapter/router"
	"backend/internal/infrastructure/mysql"
	recipeUsecase "backend/internal/usecase/recipe"
	standardUsecase "backend/internal/usecase/standard_recipe"
)

func main() {
	// ===========================================================================
	// Step 1: インフラ層 - DBコネクション
	// ===========================================================================

	db := mysql.NewDB()
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("DBクローズ中にエラー: %v", err)
		}
	}()

	// ===========================================================================
	// Step 2: インフラ層 - リポジトリ実装のインスタンス化
	// ===========================================================================

	recipeRepo := mysql.NewRecipeRepository(db)
	standardRecipeRepo := mysql.NewStandardRecipeRepository(db)
	synonymRepo := mysql.NewSynonymRepository(db)

	// ===========================================================================
	// Step 3: ユースケース層 - リポジトリを注入してユースケースを構築
	// ===========================================================================

	searchRecipeUsecase := recipeUsecase.NewSearchPersonalRecipeUsecase(recipeRepo, synonymRepo)
	getRecipeDetailUsecase := recipeUsecase.NewGetRecipeDetailUsecase(recipeRepo)

	searchStdUsecase := standardUsecase.NewSearchStandardRecipeUsecase(standardRecipeRepo, synonymRepo)
	getStdDetailUsecase := standardUsecase.NewGetStandardRecipeDetailUsecase(standardRecipeRepo)

	// ===========================================================================
	// Step 4: アダプター層 - ユースケースを注入してハンドラーを構築
	// ===========================================================================

	personalHandler := handler.NewPersonalRecipeHandler(searchRecipeUsecase, getRecipeDetailUsecase)
	standardHandler := handler.NewStandardRecipeHandler(searchStdUsecase, getStdDetailUsecase)
	logHandler := handler.NewLogActionHandler()

	// ===========================================================================
	// Step 5: アダプター層 - ハンドラーを注入してルーターを構築
	// ===========================================================================

	r := router.New(router.Dependencies{
		PersonalRecipeHandler: personalHandler,
		StandardRecipeHandler: standardHandler,
		LogActionHandler:      logHandler,
	})

	// ===========================================================================
	// Step 6: サーバー起動
	// ===========================================================================

	port := os.Getenv("PORT")
	if port == "" {
		port = "8005"
	}

	log.Printf("サーバーを起動します: :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
