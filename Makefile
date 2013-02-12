all:
	gofmt -e -s -w .
	go vet .
	go test -v ./... -gocheck.v
