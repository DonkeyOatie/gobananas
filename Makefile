build: clean
	goimports -w .
	go get .
	go build
	strip golang_blog

clean:
	-rm golang_blog
