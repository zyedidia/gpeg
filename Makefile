rwildcard=$(foreach d,$(wildcard $(1:=/*)),$(call rwildcard,$d,$2) $(filter $(subst *,%,$2),$d))

SRC=$(call rwildcard,.,*.go)

testdata/bible.txt:
	mkdir -p testdata
	curl --url http://www.gutenberg.org/cache/epub/10/pg10.txt --output testdata/bible.txt

test: $(SRC)
	go test

bench: testdata/bible.txt
	go test -run=none -bench=.

bench-pretty: testdata/bible.txt
	go test -run=none -bench=. | prettybench

coverage.out: $(SRC)
	go test -coverpkg=./... -coverprofile=coverage.out .

cover: coverage.out
	go tool cover -html=coverage.out

clean:
	rm coverage.out

.PHONY: clean cover test bench bench-pretty
