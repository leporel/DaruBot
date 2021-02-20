#!/bin/sh

protoc --proto_path=./schema/ --go_out=./schema/  --go-grpc_out=./schema/ ./schema/*.proto