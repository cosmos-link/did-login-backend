FROM golang:latest

WORKDIR /app
# 安装必要的系统库 (Ubuntu-based image uses apt)
RUN apt-get update && apt-get install -y gcc libc6-dev && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 编译应用
RUN go build -o main .

EXPOSE 8080
CMD ["./main"]
