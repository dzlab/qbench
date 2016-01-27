#!/bin/sh

export GOPATH=$(pwd)
echo "GOPATH is $GOPATH"
export GOBIN=$GOPATH/bin
export PATH=$PATH:$GOBIN
go get
cd $GOPATH
echo "Moving application to $APP_DEPLOY_DIR"
mkdir bin
mv bin/staging bin/application
