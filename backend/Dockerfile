# ビルドステージ: Go バイナリのビルド
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 依存関係のコピーとダウンロード（Dockerキャッシュを最大限活用）
COPY go.mod go.sum ./
RUN go mod download

# ソースコードのコピーとビルド
# CGO_ENABLED=0: 静的バイナリを生成（Alpine互換）
# -trimpath: デバッグ情報からローカルパスを除去
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -o /app/server ./cmd/server

# 実行ステージ: 最小限のイメージ
FROM alpine:3.21

# X.509証明書（TLS通信のため）とタイムゾーンデータをインストール
RUN apk --no-cache add ca-certificates tzdata

# JSTタイムゾーンの設定
ENV TZ=Asia/Tokyo

WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server .

EXPOSE 8005

# アプリケーションの起動
ENTRYPOINT ["/app/server"]