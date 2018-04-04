#!/usr/bin/env bash

set +e

go get -t -v ./...
gometalinter -D gocyclo --deadline=120s --skip bindata --exclude 'bindata/' --exclude 'should have comment or be unexported' ./...