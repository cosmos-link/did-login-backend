FROM golang:latest

ARG APP_PORT=0

WORKDIR /app

# 设置数据库环境变量（默认值）
ENV DB_HOST=47.84.96.59
ENV DB_PORT=3308
ENV DB_USER=root
ENV DB_PASSWORD=ykt123456
ENV DB_NAME=ykt_db
ENV REDIS_HOST=47.84.96.59
ENV REDIS_PORT=6379
ENV REDIS_PASSWORD=123456
ENV JWT_SECRET=ykt-did-platform-secret-key-2024

# 安装必要的系统库 (Ubuntu-based image uses apt)
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/ ./
# 编译应用
RUN go build -o main .

EXPOSE ${APP_PORT}

CMD ["./main"]
