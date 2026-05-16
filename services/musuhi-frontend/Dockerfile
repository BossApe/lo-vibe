FROM node:20-alpine
WORKDIR /app
ARG VITE_API_BASE_URL=http://localhost:18080
ENV VITE_API_BASE_URL=${VITE_API_BASE_URL}
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build
EXPOSE 3000
ENV NODE_ENV=production
ENV PORT=3000
CMD npm run preview -- --host 0.0.0.0 --port $PORT
