// Package handler はHTTPリクエストの受け取りとレスポンス生成を担うアダプター層のパッケージです。
package handler

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// LogActionRequest はアクションログAPIのリクエストボディ構造体です。
type LogActionRequest struct {
	// Action はログに記録するアクション名（例: "search_submit", "add_condition"）
	Action string `json:"action"`
	// Details はアクションの補足情報（検索クエリなど任意のフィールドを含めることができる）
	Details map[string]interface{} `json:"details"`
	// URL はアクションが発生したページのURL
	URL string `json:"url"`
	// Timestamp はアクションが発生した時刻（ISO 8601形式）
	Timestamp string `json:"timestamp"`
}

// LogActionResponse はアクションログAPIの成功レスポンス構造体です。
type LogActionResponse struct {
	Status string `json:"status"` // "success" または "error"
}

// LogActionHandler はアクションログAPIのHTTPハンドラーを保持する構造体です。
// このハンドラーはフロントエンドからのユーザー操作ログを受け取り、サーバーログに記録します。
// 将来的にはKafkaやDBへの書き込みに拡張可能な設計にしています。
type LogActionHandler struct{}

// NewLogActionHandler はLogActionHandlerの新しいインスタンスを生成します。
func NewLogActionHandler() *LogActionHandler {
	return &LogActionHandler{}
}

// LogAction はユーザーアクションのログを記録するAPIのハンドラーです。
//
// エンドポイント: POST /api/log_action
// リクエストボディ (JSON):
//
//	{
//	  "action": "search_submit",
//	  "details": { "query": "鶏肉 トマト" },
//	  "url": "http://...",
//	  "timestamp": "2026-03-22T10:00:00Z"
//	}
//
// Pythonの api.py の log_action() 関数に相当します。
func (h *LogActionHandler) LogAction(c *gin.Context) {
	var req LogActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "リクエストが不正です"})
		return
	}

	// サーバーログにACTION_LOGとして記録（将来的にはDBやメッセージキューに送信可能）
	log.Printf("ACTION_LOG: action=%s url=%s timestamp=%s details=%v",
		req.Action, req.URL, req.Timestamp, req.Details)

	c.JSON(http.StatusOK, LogActionResponse{Status: "success"})
}
