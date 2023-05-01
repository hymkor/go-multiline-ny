all:
	go fmt
	go build

demo:
	go run examples/example.go

.PHONY: all try
