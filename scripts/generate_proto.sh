#!/bin/bash

# Proto 代码生成脚本

set -e

echo "开始生成 Proto Go 代码..."

# 检查 protoc 是否安装
if ! command -v protoc &> /dev/null; then
    echo "错误: protoc 未安装，请先安装 Protocol Buffers compiler"
    exit 1
fi

# 检查 protoc-gen-go 是否安装
if ! command -v protoc-gen-go &> /dev/null; then
    echo "错误: protoc-gen-go 未安装，请运行: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

# 项目根目录
PROJECT_ROOT=$(pwd)

# Proto 源目录
PROTO_SOURCE_DIR="internal/rpc/proto"

# Proto 生成目录
PROTO_OUTPUT_DIR="internal/rpc/generated"

# 清理旧的生成文件
echo "清理旧的生成文件..."
rm -rf $PROTO_OUTPUT_DIR
mkdir -p $PROTO_OUTPUT_DIR/{auth,user,common}

# 生成 common proto
echo "生成 common proto..."
protoc \
    --proto_path=$PROTO_SOURCE_DIR \
    --go_out=$PROTO_OUTPUT_DIR \
    --go_opt=paths=source_relative \
    $PROTO_SOURCE_DIR/common/common.proto

# 生成 auth proto
echo "生成 auth proto..."
protoc \
    --proto_path=$PROTO_SOURCE_DIR \
    --go_out=$PROTO_OUTPUT_DIR \
    --go_opt=paths=source_relative \
    $PROTO_SOURCE_DIR/auth/auth.proto

# 生成 user proto
echo "生成 user proto..."
protoc \
    --proto_path=$PROTO_SOURCE_DIR \
    --go_out=$PROTO_OUTPUT_DIR \
    --go_opt=paths=source_relative \
    $PROTO_SOURCE_DIR/user/user.proto

echo "Proto Go 代码生成完成！"