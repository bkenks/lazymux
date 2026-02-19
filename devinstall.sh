#!/usr/bin/env bash

TEST_DIR="./testing"
mkdir -p $TEST_DIR
go build -o $TEST_DIR
cp -f $TEST_DIR/lazymux ~/.local/bin
