package main

// ============================================================================
// GO
// ============================================================================
// Gen and Build main
//go:generate /usr/bin/env bash -c "echo 'Building example golang binaries (CLI and Server)'"
//go:generate /usr/bin/env bash -c "mkdir -p bin-all/{server,cli,client}"
//go:generate go build -v -o bin-all/server ./main/server/main.go
//go:generate go build -v -o bin-all/cli ./main/cli/main.go

// ============================================================================
// Flutter
// ============================================================================
// GRPC & Protobuf
// go:generate /usr/bin/env bash -c "echo 'generating protobuf and grpc services for flutter/dart'"
// go:generate protoc -I./proto/v2 -I. --dart_out=grpc:./client/lib/api/v2/ ./proto/v2/authn.proto ./proto/v2/users.proto
//go:generate /usr/bin/env bash -c "echo 'building flutter binary'"
// go:generate make flu-web-build
