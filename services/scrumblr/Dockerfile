
# Scrumblr Dockerfile
# 公式イメージがないため、Node.js+gitでビルド
FROM node:20-slim

ARG REDIS_URL=redis://redis:6379
ENV REDIS_URL=${REDIS_URL}

WORKDIR /app

# 必要パッケージのインストールとScrumblrクローン
RUN apt-get update && \
    apt-get install -y git && \
    git clone https://github.com/aliasaria/scrumblr.git . && \
    npm install && \
    sed -i "s#        url: 'redis://127.0.0.1:6379',#        url: process.env.REDIS_URL || 'redis://redis:6379',#" server.js && \
    grep 'url:' server.js && \
    cat server.js | head -50 && \
    apt-get purge -y git && \
    apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

EXPOSE 8080

CMD ["npm", "start"]
