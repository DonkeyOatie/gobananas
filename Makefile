all: build test

test:
	go get github.com/stretchr/testify
	DBNAME=testdb go test -coverprofile=coverage.out

cover: test
	go tool cover -func=coverage.out

coverhtml: test
	go tool cover -html=coverage.out

build: clean
	goimports -w .
	go get .
	go build
	strip gobananas

clean:
	-rm gobananas
