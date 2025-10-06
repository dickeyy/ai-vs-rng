#! /bin/bash

git pull
rm -rf ./build
pm2 delete "cis-320 --debug"
go build -o build/cis-320 .
pm2 start "build/cis-320 --debug"
pm2 logs "cis-320 --debug"