.PHONY: build run clean

build:
	git clone --depth=1 --branch sing https://github.com/MetaCubeX/meta-rules-dat.git
	rm -rf meta-rules-dat/.git
	find meta-rules-dat -type f -name "*.srs" -exec rm {} +
	python3 build/preprocess/preprocessor.py
	go mod tidy
	CGO_ENABLED=1 go build -ldflags="-s -w" -o main

run:
	./main

clean:
	rm -rf meta-rules-dat main ipdomain.db ipdomain.db-shm ipdomain.db-wal