#!/bin/sh

## Forgejoの初期管理ユーザーとアクセストークンを作成するスクリプト

# 1. データベースの起動完了を少し待つ
sleep 5

# 2. ユーザー 'musuhi' が存在しない場合のみ作成
if ! forgejo admin user list | grep -q "musuhi"; then
  echo "--- Creating admin user: musuhi ---"
  forgejo admin user create \
    --admin \
    --username musuhi \
    --password musuhi \
    --email musuhi@example.com \
    --must-change-password=false
else
  echo "--- Admin user musuhi already exists ---"
fi

# 3. トークンを自動発行して、共有ボリューム（/shared/token.txt）に保存する
# (すでに生成されていても上書き保存またはエラーを無視する設定)
echo "Generating access token..."
mkdir -p /shared
forgejo admin user generate-access-token \
  --username musuhi \
  --token-name "msuhi-api-token" \
  --scopes all \
  --raw > /shared/token.txt

echo "Token successfully saved to shared volume."
