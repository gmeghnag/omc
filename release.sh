#!/bin/bash

TAG=$(git tag --points-at HEAD)
if [[ -z $TAG ]]; then
    TAG='Unknown version'
fi
if  [[ $(git diff-index --quiet HEAD --)  ]] ; then
    HASH=$(git log -n1 --pretty=format:%h)
else
    HASH=$(git log -n1 --pretty=format:%h)-dirty
fi

GOOS=linux GOARCH=amd64 go build -ldflags "-X 'omc/vars.OMCVersionTag=${TAG}' -X omc/vars.OMCVersionHash=${HASH}" -o omc-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -ldflags "-X 'omc/vars.OMCVersionTag=${TAG}' -X omc/vars.OMCVersionHash=${HASH}" -o omc-win-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -ldflags "-X 'omc/vars.OMCVersionTag=${TAG}' -X omc/vars.OMCVersionHash=${HASH}" -o omc-darwin-amd64 main.go