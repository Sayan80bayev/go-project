#!/bin/bash

PROJECT_ROOT="$(dirname "$(realpath "$0")")/.."
cd "$PROJECT_ROOT" || exit 1
protoc --go_out=. --go-grpc_out=. pkg/proto/*.proto