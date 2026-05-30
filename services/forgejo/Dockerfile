# Forgejo Dockerfile (rootless, PostgreSQL利用)
FROM codeberg.org/forgejo/forgejo:15

# 例: 設定ファイルや初期スクリプトの追加
ARG APP_INI_PATH=/data/forgejo/conf/app.ini
COPY conf/app.ini ${APP_INI_PATH}

# 初期化スクリプトを特定のディレクトリにコピー
COPY init-admin.sh /etc/templates/init.d/init-admin.sh
# スクリプトに実行権限を付与
RUN chmod +x /etc/templates/init.d/init-admin.sh

# デフォルトCMD/ENTRYPOINTは公式イメージのまま
