package vm

const (
	opChar byte = iota
	opJump
	opChoice
	opCall
	opCommit
	opReturn
	opFail
	opSet
	opAny
	opPartialCommit
	opSpan
	opBackCommit
	opFailTwice
	opTestChar
	opTestSet
	opTestAny
	opChoice2
)
