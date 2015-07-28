all: deps build test lint

ci: test lint

deps:
	goimports -w .
	go get .
	go get golang.org/x/tools/cmd/goimports
	go get github.com/stretchr/testify

test:
	DBNAME=testdb go test -coverprofile=coverage.out

cover: test
	go tool cover -func=coverage.out

coverhtml: test
	go tool cover -html=coverage.out

build: clean
	go build
	strip gobananas

clean:
	-rm gobananas

lint:
	golint .
