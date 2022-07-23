mac:
	GOOS=darwin GOARCH=amd64 go build

linux:
	GOOS=linux GOARCH=amd64 go build

arm:
	GOOS=linux GOARCH=arm go build
