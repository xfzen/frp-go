#!/bin/sh
rm -rf bin/frpgo*

#当前版本号,每次更新服务时都必须更新版本号
CurrentVersion=1.0.0

rm -rf bin/*

#项目名
Project=frpgo

Path=$Project"/version"
GitCommit=$(git rev-parse --short HEAD || echo unsupported)
BuildTime=`date "+%Y%m%d%H%M"`
Suffix="cyg-znjsadmin_$BuildTime"

CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o frpgo -ldflags "-s -w" \
  -ldflags "-X $Path.Version=$CurrentVersion   \
  -X '$Path.BuildTime=`date "+%Y-%m-%d %H:%M:%S"`'    \
  -X '$Path.GoVersion=`go version`' -X $Path.GitCommit=$GitCommit"  \
  api/frpgo.go

# run upx
# upx frpgo

echo "build finish !!"
echo "Version:" $CurrentVersion
echo "Git commit:" $GitCommit
echo "Go version:" `go version`

