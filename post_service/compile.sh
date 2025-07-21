#!/usr/bin/env bash
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
 protoc --go_out=. --go-grpc_out=. post.proto


# For python, uncomment all lines below:
#
#mkdir -p messenger/proto
#ln -s $(pwd)/messenger.proto ./messenger/proto/messenger.proto || true
#python3 -m grpc_tools.protoc -I../proto --python_out=../../ --pyi_out=../../ --grpc_python_out=../../ messenger/proto/messenger.proto