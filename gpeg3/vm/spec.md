# Virtual Machine Specification

## Encoding

Instructions in the gpeg virtual machine are variable-length and encoded with
an 8-bit opcode followed by any possible arguments. Labels are encoded as
16-bit relative offsets. Charsets are encoded as 16-bit values to look up in a
separate table that stores the 128-bit sets. Alignments for 16-bit values must
be respected which results in padding bytes being inserted for some
instructions.

The encoding of each instruction is given below:

* `Char Byte`: `| 8-bit opcode | 8-bit Byte |`
* `Jump Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `Choice Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `Call Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `Commit Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `Return`: `| 8-bit opcode | 8-bit padding |`
* `Fail`: `| 8-bit opcode | 8-bit padding |`
* `Set Chars`: `| 8-bit opcode | 8-bit padding | 16-bit index of Chars |`
* `Any Byte`: `| 8-bit opcode | 8-bit Byte |`
* `PartialCommit Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `Span Chars`: `| 8-bit opcode | 8-bit padding | 16-bit index of Chars |`
* `BackCommit Lbl`: `| 8-bit opcode | 8-bit padding | 16-bit Lbl |`
* `FailTwice`: `| 8-bit opcode | 8-bit padding |`
* `TestChar Byte Lbl`: `| 8-bit opcode | 8-bit Byte | 16-bit Lbl |`
* `TestSet Chars Lbl`: `| 8-bit opcode | 8-bit Padding | 16-bit Lbl | 16-bit index of Chars |`
* `TestAny N Lbl`: `| 8-bit opcode | 8-bit N | 16-bit Lbl |`
* `End`: `| 8-bit opcode | 8-bit padding |`
* `Nop`: can be removed and not encoded at all.
* `MemoOpen Lbl Id`: `| 8-bit opcode | 8-bit Padding | 16-bit Lbl | 16-bit Id |`
* `MemoClose`: `| 8-bit opcode | 8-bit padding |`
* `CaptureBegin`: `| 8-bit opcode | 8-bit padding | 16-bit capture Id |`
* `CaptureLate`: `| 8-bit opcode | 8-bit N |`
* `CaptureEnd`: `| 8-bit opcode | 8-bit padding | 16-bit capture Id |`
* `CaptureFull`: `| 8-bit opcode | 8-bit N | 16-bit capture Id |`

Note that the machine stores a separate table of character sets which are
looked up whenever there is an instruction involving a character set (`Set`,
`TestSet`, `Span`).

The largest instruction is 48 bits and the smallest instruction is 16 bits. All
instructions have a size that is a multiple of 16 bits.

## Operation

The virtual machine operation consists of four pieces: the instruction data, an
instruction pointer, an explicit stack, and a memoization table.
