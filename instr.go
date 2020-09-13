package gpeg

import (
	"fmt"
	"log"
)

type result int

type instr interface {
	exec(vm *vmstate)
}

// basic instructions
type iChar struct {
	n rune
}

func (i iChar) String() string {
	return fmt.Sprintf("Char %v", string(i.n))
}

type iJump struct {
	label int
}

func (i iJump) String() string {
	return fmt.Sprintf("Jump %v", i.label)
}

type iChoice struct {
	label int
}

func (i iChoice) String() string {
	return fmt.Sprintf("Choice %v", i.label)
}

type iCall struct {
	label int
}

func (i iCall) String() string {
	return fmt.Sprintf("Call %v", i.label)
}

type iOpenCall struct {
	name string
}

func (i iOpenCall) String() string {
	return fmt.Sprintf("OpenCall %v", i.name)
}

type iCommit struct {
	label int
}

func (i iCommit) String() string {
	return fmt.Sprintf("Commit %v", i.label)
}

// TODO: capture instructions

type iReturn struct{}

func (i iReturn) String() string {
	return fmt.Sprintf("Return")
}

type iFail struct{}

func (i iFail) String() string {
	return fmt.Sprintf("Fail")
}

// extra instructions for general optimization
type iCharset struct {
	set charset
}

func (i iCharset) String() string {
	return fmt.Sprintf("Charset %v", i.set)
}

type iAny struct {
	n int
}

func (i iAny) String() string {
	return fmt.Sprintf("Any %v", i.n)
}

type iPartialCommit struct {
	label int
}

func (i iPartialCommit) String() string {
	return fmt.Sprintf("PartialCommit %v", i.label)
}

type iSpan struct {
	set charset
}

func (i iSpan) String() string {
	return fmt.Sprintf("Span %v", i.set)
}

type iBackCommit struct {
	label int
}

func (i iBackCommit) String() string {
	return fmt.Sprintf("BackCommit %v", i.label)
}

type iFailTwice struct{}

func (i iFailTwice) String() string {
	return fmt.Sprintf("FailTwice")
}

// extra instructions for head-fail optimization
type iTestChar struct {
	x     rune
	label int
}

func (i iTestChar) String() string {
	return fmt.Sprintf("TestChar %v %v", string(i.x), i.label)
}

type iTestCharset struct {
	set   charset
	label int
}

func (i iTestCharset) String() string {
	return fmt.Sprintf("TestCharset %v %v", i.set, i.label)
}

type iTestAny struct {
	n     int
	label int
}

func (i iTestAny) String() string {
	return fmt.Sprintf("TestAny %v %v", i.n, i.label)
}

type iChoice2 struct {
	label int
	back  int
}

func (i iChoice2) String() string {
	return fmt.Sprintf("Choice2 %v %v", i.label, i.back)
}

func (i iChar) exec(vm *vmstate) {
	r, size := vm.input.PeekRune()
	if r == i.n {
		vm.input.SeekBytes(size, SeekCurrent)
		vm.ip++
		return
	}
	vm.ip = ipFail
}

func (i iJump) exec(vm *vmstate) {
	vm.ip += i.label
}

func (i iChoice) exec(vm *vmstate) {
	vm.stack.Push(vm.stack.BacktrackEntry(vm.ip+i.label, vm.input.Offset(), vm.caplist))
	vm.ip++
}

func (i iCall) exec(vm *vmstate) {
	vm.stack.Push(vm.stack.ReturnAddressEntry(vm.ip + 1))
	vm.ip += i.label
}

func (i iOpenCall) exec(vm *vmstate) {
	log.Fatal("OpenCall cannot executed")
}

func (i iReturn) exec(vm *vmstate) {
	entry, ok := vm.stack.Pop()
	if ok && entry.ReturnAddress() {
		vm.ip = entry.raddress
		return
	}

	log.Fatal("Return statement failed")
}

func (i iCommit) exec(vm *vmstate) {
	vm.stack.Pop()
	vm.ip += i.label
}

func (i iFail) exec(vm *vmstate) {
	vm.ip = ipFail
}

func (i iCharset) exec(vm *vmstate) {
	r, size := vm.input.PeekRune()
	if i.set.Has(r) {
		vm.input.SeekBytes(size, SeekCurrent)
		vm.ip++
		return
	}
	vm.ip = ipFail
}

func (i iAny) exec(vm *vmstate) {
	if vm.input.Offset()+i.n < vm.input.Len() {
		for j := 0; j < i.n; j++ {
			_, size := vm.input.PeekRune()
			vm.input.SeekBytes(size, SeekCurrent)
		}
		vm.ip++
		return
	}

	vm.ip = ipFail
}

func (i iPartialCommit) exec(vm *vmstate) {
	ent := vm.stack.Peek()
	if ent == nil || ent.ReturnAddress() {
		log.Fatal("partial commit failed")
	}

	ent.btrack.off = vm.input.Offset()
	ent.btrack.caplist = vm.caplist
	vm.ip += i.label
}

func (i iSpan) exec(vm *vmstate) {
	r, size := vm.input.PeekRune()
	if i.set.Has(r) {
		vm.input.SeekBytes(size, SeekCurrent)
		return
	}

	vm.ip++
}

func (i iFailTwice) exec(vm *vmstate) {
	vm.stack.Pop()
	vm.ip = ipFail
}

func (i iBackCommit) exec(vm *vmstate) {
	ent, ok := vm.stack.Pop()
	if ok && !ent.ReturnAddress() {
		vm.input.SeekBytes(ent.btrack.off, SeekStart)
		vm.caplist = ent.btrack.caplist
		vm.ip += i.label
		return
	}
	log.Fatal("back commit failed")
}

func (i iTestChar) exec(vm *vmstate) {
	r, size := vm.input.PeekRune()
	if r == i.x {
		vm.input.SeekBytes(size, SeekCurrent)
		vm.ip++
		return
	}

	vm.ip += i.label
}

func (i iTestCharset) exec(vm *vmstate) {
	r, size := vm.input.PeekRune()
	if i.set.Has(r) {
		vm.input.SeekBytes(size, SeekCurrent)
		vm.ip++
		return
	}

	vm.ip += i.label
}

func (i iTestAny) exec(vm *vmstate) {
	if vm.input.Offset()+i.n < vm.input.Len() {
		for j := 0; j < i.n; j++ {
			_, size := vm.input.PeekRune()
			vm.input.SeekBytes(size, SeekCurrent)
		}
		vm.ip++
		return
	}

	vm.ip += i.label
}

func (i iChoice2) exec(vm *vmstate) {
	// TODO: i.back needs to support unicode
	vm.stack.Push(vm.stack.BacktrackEntry(vm.ip+i.label, vm.input.Offset()-i.back, vm.caplist))
	vm.ip++
}
