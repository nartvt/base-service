#!/bin/bash

echo "Remove base-service: rm -rf ./server"
rm -rf ./server

echo "make swagger doccument"
make docs

echo "Build base-service: go mod tidy && go mod vendor"
go mod tidy && go mod vendor

echo "Build binary base-service: go build -o base-service main.go"
go build -o server main.go

echo "Run base-service"
./server -config config -env ""


