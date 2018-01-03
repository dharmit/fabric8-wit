#!/bin/bash

yum install -y golang epel-release docker
yum install -y glide mercurial 

export GOPATH=~/go

go get -d github.com/fabric8-services/fabric8-wit
cd $GOPATH/src/github.com/fabric8-services/fabric8-wit

systemctl enable --now docker
make
