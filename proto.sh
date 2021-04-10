#!/usr/bin/env bash

export GO_PATH=~/go
export PATH=$PATH:/$GO_PATH/bin

go install google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/protobuf/cmd/protoc-gen-go
go get google.golang.org/grpc/cmd/protoc-gen-go-grpc


SRC_DIR="proto"
DST_DIR="gen"

rm -rf $DST_DIR
mkdir $DST_DIR

protoc -I=$SRC_DIR --go_out=$DST_DIR --go_opt=module=backend/gen $SRC_DIR/base/*

protoc -I=$SRC_DIR --go_out=$DST_DIR --go-grpc_out=$DST_DIR --go_opt=module=backend/gen --go-grpc_opt=module=backend/gen $SRC_DIR/service/*