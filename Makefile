# よく使うコマンドをショートカット化しました

# 1. 環境のセットアップ（ライブラリの整理・インストール）
# 使い方: make tidy
tidy:
	docker compose run --rm backend go mod tidy

# 2. 新しいライブラリを追加する
# 使い方: make add pkg=ライブラリ名
add:
	docker compose run --rm backend go get $(pkg)

# 3. サーバーを起動する
# 使い方: make up
up:
	docker compose up -d

# 4. サーバーを停止する
# 使い方: make down
down:
	docker compose down

# 5. ログを見る
# 使い方: make logs
logs:
	docker compose logs -f backend

# 6. コンテナの中に入る（デバッグ用）
# 使い方: make sh
sh:
	docker compose run --rm backend sh