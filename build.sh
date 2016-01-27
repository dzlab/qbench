#!/bin/sh

export GOPATH=$(pwd)
echo "GOPATH is $GOPATH"
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
#mkdir /tmp/integration-endpoint
#mv ./* /tmp/integration-endpoint
#mkdir -p src/github.com/adomik/integration-endpoint
#cp -ap -R /tmp/integration-endpoint/* src/github.com/adomik/integration-endpoint
#cd src/github.com/adomik/integration-endpoint
go get
#go build application.go
go build
cd $GOPATH
echo "Moving application to $APP_DEPLOY_DIR"
#mkdir bin && mv src/github.com/adomik/integration-endpoint/application bin
mkdir bin
mv bin/staging bin/application
