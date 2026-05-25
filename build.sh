#!/bin/bash

set -e

# Получитьб информацию о версии
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "1.0.0")
BUILD_DATE=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "🔨 Building GophKeeper..."
echo "   Version:    $VERSION"
echo "   Build Date: $BUILD_DATE"
echo "   Git Commit: $GIT_COMMIT"
echo ""


mkdir -p dist

# Функция для сборки под конкретную платформу
build_for_platform() {
    local os=$1
    local arch=$2
    local output=$3
    
    echo "📦 Building for $os/$arch..."
    GOOS=$os GOARCH=$arch go build \
        -ldflags "-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE -X main.GitCommit=$GIT_COMMIT" \
        -o "dist/$output" \
        ./cmd/client/
}

# Сборка под разные платформы
build_for_platform "linux" "amd64" "gophkeeper_linux_amd64"
build_for_platform "linux" "arm64" "gophkeeper_linux_arm64"
build_for_platform "darwin" "amd64" "gophkeeper_darwin_amd64"
build_for_platform "darwin" "arm64" "gophkeeper_darwin_arm64"
build_for_platform "windows" "amd64" "gophkeeper_windows_amd64.exe"
build_for_platform "windows" "386" "gophkeeper_windows_386.exe"

echo ""
echo "✅ Build complete! Check the 'dist' directory:"
ls -la dist/