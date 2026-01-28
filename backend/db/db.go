package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	_ "github.com/go-sql-driver/mysql"
)

// データベースの接続と，接続状態を確認する関数
func Init() *sql.DB {
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	dbHost := os.Getenv("DB_HOST")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("データベース設定エラー:", err)
	}
	// 通信テスト
	if err := db.Ping(); err != nil {
		log.Fatal("データベース接続テスト失敗（接続情報を確認）:", err)
	}
	fmt.Println("データベース接続成功")
	return db
}