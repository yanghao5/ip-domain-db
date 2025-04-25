#!/bin/bash

cd /tmp

git clone --depth=1 https://github.com/yanghao5/ip-domain-db.git && cd ip-domain-db
make build
make run

mv ipdomain.db ~/file
