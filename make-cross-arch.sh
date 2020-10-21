#!/bin/bash

# Reference:
# https://github.com/golang/go/blob/master/src/go/build/syslist.go

# don't set GOPATH as it conflicts with the module compilation

os_archs=(
    linux/amd64
    plan9/amd64
)

os_archs_64=()

for os_arch in "${os_archs[@]}"
do
    goos=${os_arch%/*}
    goarch=${os_arch#*/}
    mkdir -p ./bin/${goos}/${goarch}

    buildTime=$(date +'%Y-%m-%dT%T')
    sha1ver=$(git rev-parse HEAD)
    GOOS=${goos} GOARCH=${goarch} go build -ldflags "-X main.sha1ver=${sha1ver} -X main.buildTime=${buildTime}" -o ./bin/${GOOS}/${GOARCH}
    #GOOS="plan9" GOARCH="amd64" go build -gcflags="all=-N -l -dwarf" -o ./bin/${GOOS}/${GOARCH}
    if [ $? -eq 0 ]
    then
        os_archs_64+=(${os_arch})
    fi
done

echo "64-bit:"
for os_arch in "${os_archs_64[@]}"
do
    printf "\t%s\n" "${os_arch}"
done
echo
