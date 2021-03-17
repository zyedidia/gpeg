encoding:
	go run encoding.go -grammar ../grammars/java.peg > java.dat
	go run encoding.go -grammar ../grammars/c.peg > c.dat
	go run encoding.go -grammar ../grammars/peg.peg > peg.dat
	go run encoding.go -ins -grammar ../grammars/java.peg > java_ins.dat
	go run encoding.go -ins -grammar ../grammars/c.peg > c_ins.dat
	go run encoding.go -ins -grammar ../grammars/peg.peg > peg_ins.dat

.PHONY: encoding
