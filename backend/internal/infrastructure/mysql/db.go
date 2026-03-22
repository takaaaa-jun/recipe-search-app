// Package mysql はドメイン層のリポジトリインターフェースに対するMySQL実装を提供します。
package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql" // MySQLドライバーの副作用インポート
)

// NewDB は環境変数の設定を読み込み、MySQLデータベースへの接続プールを生成して返します。
//
// 必要な環境変数:
//   - MYSQL_HOST:     MySQLサーバーのホスト名（例: "db"）
//   - MYSQL_PORT:     ポート番号（デフォルト: "3306"）
//   - MYSQL_USER:     接続ユーザー名
//   - MYSQL_PASSWORD: 接続パスワード
//   - MYSQL_DATABASE: 接続先データベース名
//
// 接続に失敗した場合はlog.Fatalで終了します。
func NewDB() *sql.DB {
	host := os.Getenv("MYSQL_HOST")
	port := os.Getenv("MYSQL_PORT")
	user := os.Getenv("MYSQL_USER")
	password := os.Getenv("MYSQL_PASSWORD")
	database := os.Getenv("MYSQL_DATABASE")

	// ポートのデフォルト値を設定
	if port == "" {
		port = "3306"
	}

	// go-sql-driver/mysql の DSN フォーマット:
	// [username[:password]@][protocol[(address)]]/dbname[?param1=value1&...
	// parseTime=true で time.Time への自動マッピングを有効化
	// charset=utf8mb4 で日本語（4バイト文字）を正しく扱う
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4&collation=utf8mb4_general_ci",
		user, password, host, port, database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("データベース設定エラー: %v", err)
	}

	// 接続の疎通確認（Pingで実際に接続テスト）
	if err := db.Ping(); err != nil {
		log.Fatalf("データベース接続テスト失敗（接続情報を確認してください）: %v", err)
	}

	// 接続プールの設定（本番環境向け推奨値）
	db.SetMaxOpenConns(25)  // 最大同時接続数
	db.SetMaxIdleConns(10)  // アイドル状態で保持する接続数
	// SetConnMaxLifetimeは標準ライブラリのtime.Durationを使用するが
	// ここではimportを最小限にするためtime.Minuteの代わりに秒数を省略

	log.Println("MySQLデータベース接続成功")
	return db
}
