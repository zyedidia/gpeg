SRC = $(wildcard *.go)

test: $(SRC)
	go test

coverage.out: $(SRC)
	go test -coverprofile=coverage.out

cover: coverage.out
	go tool cover -html=coverage.out

clean:
	rm coverage.out

.PHONY: clean cover test
