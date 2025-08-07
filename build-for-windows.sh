#!/bin/bash
CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC="zig cc -target x86_64-windows" CXX="zig cc -target x86_64-windows" go build -tags desktop,production -ldflags "-w -s -H windowsgui" -o tella.exe
