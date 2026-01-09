#!/bin/sh

echo "=========================================="
echo "Starting Go Backend Service"
echo "=========================================="

# 陷阱函数：当脚本退出时处理清理
cleanup() {
    echo "Received signal, shutting down..."
    if [ -n "$GO_PID" ]; then
        echo "Stopping Go server (PID: $GO_PID)..."
        kill -TERM $GO_PID 2>/dev/null
    fi
    # 等待进程结束
    wait $GO_PID 2>/dev/null
    echo "Service stopped."
    exit 0
}

# 注册信号处理
trap cleanup SIGTERM SIGINT

echo "Configuration:"
echo "  Go Service Port: 60208"
echo ""

# 设置 Go 应用环境变量
export PORT=60208
export DB_HOST=${DB_HOST:-47.84.96.59}
export DB_PORT=${DB_PORT:-3308}
export DB_USER=${DB_USER:-root}
export DB_PASSWORD=${DB_PASSWORD:-ykt123456}
export DB_NAME=${DB_NAME:-ykt_db}
export REDIS_HOST=${REDIS_HOST:-47.84.96.59}
export REDIS_PORT=${REDIS_PORT:-6379}
export REDIS_PASSWORD=${REDIS_PASSWORD:-123456}
export JWT_SECRET=${JWT_SECRET:-ykt-did-platform-secret-key-2024}

# 切换到 Go 应用工作目录
cd /app

# 验证静态文件存在
echo "Checking static files..."
if [ ! -d "static" ]; then
  echo "ERROR: static directory not found in /app"
  exit 1
fi
echo "✓ Static files found:"
ls -la static/

echo "Starting Go backend service..."
./did-server &
GO_PID=$!

# 等待 Go 应用启动
sleep 2
if ! kill -0 $GO_PID 2>/dev/null; then
    echo "ERROR: Go server failed to start!"
    exit 1
fi
echo "✓ Go server started successfully (PID: $GO_PID)"

echo ""
echo "=========================================="
echo "Go Service is running!"
echo "=========================================="
echo "Go Server PID: $GO_PID"
echo "Listening on port: 60208"
echo ""

# 等待Go进程退出
wait $GO_PID
EXIT_CODE=$?

echo ""
echo "Service exited with code: $EXIT_CODE"
echo "=========================================="

# 清理
cleanup
