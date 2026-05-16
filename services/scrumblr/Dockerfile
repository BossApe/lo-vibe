# Scrumblr Dockerfile
# 公式イメージがないため、Node.js+gitでビルド
FROM node:20-slim

WORKDIR /app

# 必要パッケージのインストールとScrumblrクローン
RUN apt-get update && \
    apt-get install -y git && \
    git clone https://github.com/aliasaria/scrumblr.git . && \
    npm install && \
    apt-get purge -y git && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 3000

CMD ["npm", "start"]
