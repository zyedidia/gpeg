# GPeg

[![Documentation](https://godoc.org/github.com/zyedidia/gpeg?status.svg)](http://godoc.org/github.com/zyedidia/gpeg)
[![Go Report Card](https://goreportcard.com/badge/github.com/zyedidia/gpeg)](https://goreportcard.com/report/github.com/zyedidia/gpeg)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/zyedidia/gpeg/blob/master/LICENSE)

GPeg is a tool for working with parsing expression grammars (PEGs). It is
built with three primary goals in mind:

* Efficient parsing for two use-cases.
    * Language grammars with AST construction (where PEGs serve as a CFG
      alternative)
    * Patterns (where PEGs serve as a regex alternative).
* Incremental parsing.
* Support for dynamically loading grammars (meaning parsers can be generated
  and used at runtime).

GPeg uses the same general parsing techniques as Lua's LPeg library and is
heavily inspired by LPeg.

# Features

* Fast incremental parsing.
* Parsing virtual machine (parsers can be dynamically generated).
* Pattern compiler with optimizations.
* Support for the original PEG syntax with some extensions.
* Parse more complex string data structures (via ReaderAt interface).
* Support for back-references (context-sensitivity).
* Can convert most Go regular expressions to PEGs (see the `rxconv` package).
* Basic error recovery.
* Syntax highlighting library ([zyedidia/flare](https://github.com/zyedidia/flare)).
* Tools for visualizing grammars, ASTs, and memo tables ([zyedidia/gpeg-extra](https://github.com/zyedidia/gpeg-extra)).

# Publications

* Zachary Yedidia and Stephen Chong. "Fast Incremental PEG Parsing." Proceedings of the 14th ACM SIGPLAN International Conference on Software Language Engineering (SLE), October 2021. [Link](https://zyedidia.github.io/preprints/gpeg_sle21.pdf).
* Zachary Yedidia. "Incremental PEG Parsing." Bachelor's thesis. [Link](https://zyedidia.github.io/notes/yedidia_thesis.pdf).

# Related work

* Ford, Bryan. "Parsing expression grammars: a recognition-based syntactic foundation." Proceedings of the 31st ACM SIGPLAN-SIGACT symposium on Principles of programming languages. 2004. [Link](https://bford.info/pub/lang/peg.pdf).
* [LPeg](http://www.inf.puc-rio.br/~roberto/lpeg/).
    * Ierusalimschy, Roberto. "A text pattern‐matching tool based on Parsing
      Expression Grammars." Software: Practice and Experience 39.3 (2009):
      221-258. [Link](http://www.inf.puc-rio.br/~roberto/docs/peg.pdf).
    * Medeiros, Sérgio, and Fabio Mascarenhas. "Syntax error recovery in
      parsing expression grammars." Proceedings of the 33rd Annual ACM
      Symposium on Applied Computing. 2018.
      [Link](https://arxiv.org/pdf/1806.11150.pdf).
    * Medeiros, Sérgio, Fabio Mascarenhas, and Roberto Ierusalimschy. "Left
      recursion in parsing expression grammars." Science of Computer
      Programming 96 (2014): 177-190.
      [Link](https://arxiv.org/pdf/1207.0443.pdf).
* [NPeg](https://github.com/zevv/npeg).
* [Papa Carlo](https://lakhin.com/projects/papa-carlo/).
* Dubroy, Patrick, and Alessandro Warth. "Incremental packrat parsing."
  Proceedings of the 10th ACM SIGPLAN International Conference on Software
  Language Engineering. 2017.
* Marcelo Oikawa, Roberto Ierusalimschy, Ana Lucia de Moura. "Converting regexes to Parsing Expression Grammars." [Link](http://www.inf.puc-rio.br/~roberto/docs/ry10-01.pdf).
  [Link](https://ohmlang.github.io/pubs/sle2017/incremental-packrat-parsing.pdf).
* [Tree Sitter](https://tree-sitter.github.io/tree-sitter/).

