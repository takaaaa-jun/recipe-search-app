// Package router はGinエンジンの設定とルーティング定義を集約するパッケージです。
// 全エンドポイントの登録をここに集約することで、ルーティングの全体像を一箇所で把握できます。
// 新しいエンドポイントを追加する際はこのファイルを修正します。
package router

import (
	"net/http"

	"backend/internal/adapter/handler"

	"github.com/gin-gonic/gin"
)

// Dependencies はルーターが必要とするハンドラーの依存関係を保持する構造体です。
// cmd/server/main.go で組み立てた依存関係をここに渡します。
type Dependencies struct {
	// PersonalRecipeHandler はパーソナルレシピ系のHTTPハンドラー
	PersonalRecipeHandler *handler.PersonalRecipeHandler
	// StandardRecipeHandler は基準レシピ系のHTTPハンドラー
	StandardRecipeHandler *handler.StandardRecipeHandler
	// LogActionHandler はアクションログのHTTPハンドラー
	LogActionHandler *handler.LogActionHandler
}

// New はGinエンジンをセットアップし、全エンドポイントを登録したルーターを返します。
// この関数から返された *gin.Engine を main.go でサーバー起動に使用します。
func New(deps Dependencies) *gin.Engine {
	r := gin.Default()

	// ===========================================================================
	// ミドルウェア設定
	// ===========================================================================

	// CORSミドルウェア: フロントエンド（別オリジン）からのリクエストを許可する
	// 本番環境では許可するオリジンを環境変数から読み込むことを推奨
	r.Use(corsMiddleware())

	// ===========================================================================
	// ヘルスチェックエンドポイント
	// ===========================================================================

	// GET /health はサーバーの稼働確認に使用します（Kubernetes liveness probe 等）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// ===========================================================================
	// APIエンドポイント定義
	// ------------------------------------------------------------------
	// エンドポイント一覧:
	//   GET  /api/recipes/search       パーソナルレシピ検索
	//   GET  /api/recipes/:id          パーソナルレシピ詳細
	//   POST /api/standard-recipes/search  基準レシピ検索
	//   GET  /api/standard-recipes/:id     基準レシピ詳細
	//   POST /api/log_action           ユーザーアクションログ
	// ===========================================================================

	api := r.Group("/api")
	{
		// --- パーソナルレシピ ---
		recipes := api.Group("/recipes")
		{
			// GET /api/recipes/search?query=鶏肉&start_id=100
			recipes.GET("/search", deps.PersonalRecipeHandler.SearchRecipes)

			// GET /api/recipes/:id
			recipes.GET("/:id", deps.PersonalRecipeHandler.GetRecipeDetail)
		}

		// --- 基準レシピ ---
		standardRecipes := api.Group("/standard-recipes")
		{
			// POST /api/standard-recipes/search
			// Body: { "query": "野菜炒め", "search_mode": "recipe" }
			standardRecipes.POST("/search", deps.StandardRecipeHandler.SearchStandardRecipes)

			// GET /api/standard-recipes/:id
			standardRecipes.GET("/:id", deps.StandardRecipeHandler.GetStandardRecipeDetail)
		}

		// --- ユーティリティ ---
		// POST /api/log_action
		api.POST("/log_action", deps.LogActionHandler.LogAction)
	}

	return r
}

// corsMiddleware はCORSヘッダーを設定するGinミドルウェアを返します。
// OPTIONSリクエスト（プリフライト）にも対応しています。
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, X-CSRF-Token, Authorization")

		// OPTIONSリクエスト（CORSプリフライト）は204で即時返却
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
