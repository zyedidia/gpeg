# GPeg

**Warning:** this library is currently in alpha and the API is subject to
change. Please wait for a stable release before using it in a project. Examples
and documentation will be provided with the first release.

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

# Roadmap

* [x] Parsing virtual machine.
* [x] Pattern compiler with optimizations.
* [x] Support for original PEG syntax.
* [x] AST generation/captures.
* [x] Parse non-string data structures (reader interface).
    * [ ] Support for custom offset types (e.g. line/col instead of index).
* [x] Incremental matching.
* [x] Incremental AST/captures.
* [ ] Support for an extended PEG syntax (TBD).
* [ ] Compilation of VM code to native code (JIT or static compilation TBD).
* [ ] Documentation and examples.
* [ ] Creation of a large bundle of PEGs for common programming languages.
* [ ] Additional niceties (TBD).
    * [x] Grammar tree visualizer.
    * [x] AST visualizer.
    * [ ] Error recovery.
    * [ ] Support for left recursion.

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
  [Link](https://ohmlang.github.io/pubs/sle2017/incremental-packrat-parsing.pdf).
* [Tree Sitter](https://tree-sitter.github.io/tree-sitter/).
