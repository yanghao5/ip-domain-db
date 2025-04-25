# ip-domain-db

```
git clone --depth=1 --branch sing git@github.com:MetaCubeX/meta-rules-dat.git
rm -rf meta-rules-dat/.git
find meta-rules-dat -type f -name "*.srs" -exec rm {} +
go build -ldflags="-s -w" -o main
./main
rm -rf meta-rules-dat main
```
