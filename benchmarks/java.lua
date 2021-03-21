local re = require"re"
local lpeg = require"lpeg"

local m = re.compile([[

CompilationUnit <- Spacing PackageDeclaration? ImportDeclaration* TypeDeclaration* EOT
PackageDeclaration <- Annotation* PACKAGE QualifiedIdentifier SEMI
ImportDeclaration <- IMPORT STATIC? QualifiedIdentifier (DOT STAR)? SEMI

TypeDeclaration <- Modifier* (ClassDeclaration
			     / EnumDeclaration
			     / InterfaceDeclaration
			     / AnnotationTypeDeclaration)
		 / SEMI

ClassDeclaration <- CLASS Identifier TypeParameters? (EXTENDS ClassType)? (IMPLEMENTS ClassTypeList)? ClassBody

ClassBody <- LWING ClassBodyDeclaration* RWING

ClassBodyDeclaration
   <- SEMI
    / STATIC? Block                              
    / Modifier* MemberDecl                        

MemberDecl
   <- TypeParameters GenericMethodOrConstructorRest
    / Type Identifier MethodDeclaratorRest         
    / Type VariableDeclarators SEMI              
    / VOID Identifier VoidMethodDeclaratorRest    
    / Identifier ConstructorDeclaratorRest         
    / InterfaceDeclaration                          
    / ClassDeclaration                               
    / EnumDeclaration                                 
    / AnnotationTypeDeclaration                        

GenericMethodOrConstructorRest
    <- (Type / VOID) Identifier MethodDeclaratorRest
    / Identifier ConstructorDeclaratorRest

MethodDeclaratorRest
    <- FormalParameters Dim* (THROWS ClassTypeList)? (MethodBody / SEMI)

VoidMethodDeclaratorRest
    <- FormalParameters (THROWS ClassTypeList)? (MethodBody / SEMI)

ConstructorDeclaratorRest
    <- FormalParameters (THROWS ClassTypeList)? MethodBody

MethodBody
    <- Block

InterfaceDeclaration
    <- INTERFACE Identifier TypeParameters? (EXTENDS ClassTypeList)? InterfaceBody

InterfaceBody
    <- LWING InterfaceBodyDeclaration* RWING

InterfaceBodyDeclaration
    <- Modifier* InterfaceMemberDecl
    / SEMI

InterfaceMemberDecl
    <- InterfaceMethodOrFieldDecl
    / InterfaceGenericMethodDecl
    / VOID Identifier VoidInterfaceMethodDeclaratorRest
    / InterfaceDeclaration
    / AnnotationTypeDeclaration
    / ClassDeclaration
    / EnumDeclaration

InterfaceMethodOrFieldDecl
    <- Type Identifier InterfaceMethodOrFieldRest

InterfaceMethodOrFieldRest
    <- ConstantDeclaratorsRest SEMI
    / InterfaceMethodDeclaratorRest

InterfaceMethodDeclaratorRest
    <- FormalParameters Dim* (THROWS ClassTypeList)? SEMI

InterfaceGenericMethodDecl
    <- TypeParameters (Type / VOID) Identifier InterfaceMethodDeclaratorRest

VoidInterfaceMethodDeclaratorRest
    <- FormalParameters (THROWS ClassTypeList)? SEMI

ConstantDeclaratorsRest
    <- ConstantDeclaratorRest (COMMA ConstantDeclarator)*

ConstantDeclarator
    <- Identifier ConstantDeclaratorRest

ConstantDeclaratorRest
    <- Dim* EQU VariableInitializer

EnumDeclaration
    <- ENUM Identifier (IMPLEMENTS ClassTypeList)? EnumBody

EnumBody
    <- LWING EnumConstants? COMMA? EnumBodyDeclarations? RWING

EnumConstants
    <- EnumConstant (COMMA EnumConstant)*

EnumConstant
    <- Annotation* Identifier Arguments? ClassBody?

EnumBodyDeclarations
    <- SEMI ClassBodyDeclaration*

LocalVariableDeclarationStatement
    <- (FINAL / Annotation)* Type VariableDeclarators SEMI

VariableDeclarators
    <- VariableDeclarator (COMMA VariableDeclarator)*

VariableDeclarator
    <- Identifier Dim* (EQU VariableInitializer)?

FormalParameters
    <- LPAR FormalParameterList? RPAR

FormalParameter
    <- (FINAL / Annotation)* Type VariableDeclaratorId

LastFormalParameter
    <- (FINAL / Annotation)* Type ELLIPSIS VariableDeclaratorId

FormalParameterList
    <- FormalParameter (COMMA FormalParameter)* (COMMA LastFormalParameter)?
    / LastFormalParameter

VariableDeclaratorId
    <- Identifier Dim*

Block
    <- LWING BlockStatements RWING

BlockStatements
    <- BlockStatement*

BlockStatement
    <- LocalVariableDeclarationStatement
    / Modifier*
      ( ClassDeclaration
      / EnumDeclaration
      )
    / Statement

Statement
    <- Block
    / ASSERT Expression (COLON Expression)? SEMI
    / IF ParExpression Statement (ELSE Statement)?
    / FOR LPAR ForInit? SEMI Expression? SEMI ForUpdate? RPAR Statement
    / FOR LPAR FormalParameter COLON Expression RPAR Statement
    / WHILE ParExpression Statement
    / DO Statement WHILE ParExpression   SEMI
    / TRY LPAR Resource (SEMI Resource)* SEMI? RPAR Block Catch* Finally?
    / TRY Block (Catch+ Finally? / Finally)
    / SWITCH ParExpression LWING SwitchBlockStatementGroups RWING
    / SYNCHRONIZED ParExpression Block
    / RETURN Expression? SEMI
    / THROW Expression   SEMI
    / BREAK Identifier? SEMI
    / CONTINUE Identifier? SEMI
    / SEMI
    / StatementExpression SEMI
    / Identifier COLON Statement

Resource
    <- Modifier* Type VariableDeclaratorId EQU Expression

Catch
    <- CATCH LPAR (FINAL / Annotation)* Type (OR Type)* VariableDeclaratorId RPAR Block

Finally
    <- FINALLY Block

SwitchBlockStatementGroups
    <- SwitchBlockStatementGroup*

SwitchBlockStatementGroup
    <- SwitchLabel BlockStatements

SwitchLabel
    <- CASE ConstantExpression COLON
    / CASE EnumConstantName COLON
    / DEFAULT COLON

ForInit
    <- (FINAL / Annotation)* Type VariableDeclarators
    / StatementExpression (COMMA StatementExpression)*

ForUpdate
    <- StatementExpression (COMMA StatementExpression)*

EnumConstantName
    <- Identifier

StatementExpression
    <- Expression


ConstantExpression
    <- Expression

Expression
    <- ConditionalExpression (AssignmentOperator ConditionalExpression)*


AssignmentOperator
    <- EQU
    / PLUSEQU
    / MINUSEQU
    / STAREQU
    / DIVEQU
    / ANDEQU
    / OREQU
    / HATEQU
    / MODEQU
    / SLEQU
    / SREQU
    / BSREQU

ConditionalExpression
    <- ConditionalOrExpression (QUERY Expression COLON ConditionalOrExpression)*

ConditionalOrExpression
    <- ConditionalAndExpression (OROR ConditionalAndExpression)*

ConditionalAndExpression
    <- InclusiveOrExpression (ANDAND InclusiveOrExpression)*

InclusiveOrExpression
    <- ExclusiveOrExpression (OR ExclusiveOrExpression)*

ExclusiveOrExpression
    <- AndExpression (HAT AndExpression)*

AndExpression
    <- EqualityExpression (AND EqualityExpression)*

EqualityExpression
    <- RelationalExpression ((EQUAL /  NOTEQUAL) RelationalExpression)*

RelationalExpression
    <- ShiftExpression ((LE / GE / LT / GT) ShiftExpression / INSTANCEOF ReferenceType)*

ShiftExpression
    <- AdditiveExpression ((SL / SR / BSR) AdditiveExpression)*

AdditiveExpression
    <- MultiplicativeExpression ((PLUS / MINUS) MultiplicativeExpression)*

MultiplicativeExpression
    <- UnaryExpression ((STAR / DIV / MOD) UnaryExpression)*

UnaryExpression
    <- PrefixOp UnaryExpression
    / LPAR Type RPAR UnaryExpression
    / Primary (Selector)* (PostfixOp)*

Primary
    <- ParExpression
    / NonWildcardTypeArguments (ExplicitGenericInvocationSuffix / THIS Arguments)
    / THIS Arguments?
    / SUPER SuperSuffix
    / Literal
    / NEW Creator
    / QualifiedIdentifier IdentifierSuffix?
    / BasicType Dim* DOT CLASS
    / VOID DOT CLASS

IdentifierSuffix
    <- LBRK ( RBRK Dim* DOT CLASS / Expression RBRK)
    / Arguments
    / DOT
      ( CLASS
      / ExplicitGenericInvocation
      / THIS
      / SUPER Arguments
      / NEW NonWildcardTypeArguments? InnerCreator
      )

ExplicitGenericInvocation
    <- NonWildcardTypeArguments ExplicitGenericInvocationSuffix

NonWildcardTypeArguments
    <- LPOINT ReferenceType (COMMA ReferenceType)* RPOINT

ExplicitGenericInvocationSuffix
    <- SUPER SuperSuffix
    / Identifier Arguments

PrefixOp
    <- INC
    / DEC
    / BANG
    / TILDA
    / PLUS
    / MINUS

PostfixOp
    <- INC
    / DEC

Selector
    <- DOT Identifier Arguments?
    / DOT ExplicitGenericInvocation
    / DOT THIS
    / DOT SUPER SuperSuffix
    / DOT NEW NonWildcardTypeArguments? InnerCreator
    / DimExpr

SuperSuffix
    <- Arguments
    / DOT Identifier Arguments?

BasicType
    <- ( 'byte'
      / 'short'
      / 'char'
      / 'int'
      / 'long'
      / 'float'
      / 'double'
      / 'boolean'
      ) !LetterOrDigit Spacing

Arguments
    <- LPAR (Expression (COMMA Expression)*)? RPAR

Creator
    <- NonWildcardTypeArguments? CreatedName ClassCreatorRest
    / NonWildcardTypeArguments? (ClassType / BasicType) ArrayCreatorRest

CreatedName
    <- Identifier NonWildcardTypeArguments? (DOT Identifier NonWildcardTypeArguments?)*

InnerCreator
    <- Identifier ClassCreatorRest

ArrayCreatorRest
    <- LBRK ( RBRK Dim* ArrayInitializer / Expression RBRK DimExpr* Dim* )


ClassCreatorRest
    <- Diamond? Arguments ClassBody?

Diamond
    <- LPOINT RPOINT

ArrayInitializer
    <- LWING (VariableInitializer (COMMA VariableInitializer)*)? COMMA?  RWING

VariableInitializer
    <- ArrayInitializer
    / Expression

ParExpression
    <- LPAR Expression RPAR

QualifiedIdentifier
    <- Identifier (DOT Identifier)*

Dim
    <- LBRK RBRK

DimExpr
    <- LBRK Expression RBRK

Type
    <- (BasicType / ClassType) Dim*

ReferenceType
    <- BasicType Dim+
    / ClassType Dim*

ClassType
    <- Identifier TypeArguments? (DOT Identifier TypeArguments?)*

ClassTypeList
    <- ClassType (COMMA ClassType)*

TypeArguments
    <- LPOINT TypeArgument (COMMA TypeArgument)* RPOINT

TypeArgument
    <- ReferenceType
    / QUERY ((EXTENDS / SUPER) ReferenceType)?

TypeParameters
    <- LPOINT TypeParameter (COMMA TypeParameter)* RPOINT

TypeParameter
    <- Identifier (EXTENDS Bound)?

Bound
    <- ClassType (AND ClassType)*

Modifier
    <- Annotation
    / ( 'public'
      / 'protected'
      / 'private'
      / 'static'
      / 'abstract'
      / 'final'
      / 'native'
      / 'synchronized'
      / 'transient'
      / 'volatile'
      / 'strictfp'
      ) !LetterOrDigit Spacing

AnnotationTypeDeclaration
    <- AT INTERFACE Identifier AnnotationTypeBody

AnnotationTypeBody
    <- LWING AnnotationTypeElementDeclaration* RWING

AnnotationTypeElementDeclaration
    <- Modifier* AnnotationTypeElementRest
    / SEMI

AnnotationTypeElementRest
    <- Type AnnotationMethodOrConstantRest SEMI
    / ClassDeclaration
    / EnumDeclaration
    / InterfaceDeclaration
    / AnnotationTypeDeclaration

AnnotationMethodOrConstantRest
    <- AnnotationMethodRest
    / AnnotationConstantRest

AnnotationMethodRest
    <- Identifier LPAR RPAR DefaultValue?

AnnotationConstantRest
    <- VariableDeclarators

DefaultValue
    <- DEFAULT ElementValue

Annotation
    <- NormalAnnotation
    / SingleElementAnnotation
    / MarkerAnnotation

NormalAnnotation
    <- AT QualifiedIdentifier LPAR ElementValuePairs? RPAR

SingleElementAnnotation
    <- AT QualifiedIdentifier LPAR ElementValue RPAR

MarkerAnnotation
    <- AT QualifiedIdentifier

ElementValuePairs
    <- ElementValuePair (COMMA ElementValuePair)*

ElementValuePair
    <- Identifier EQU ElementValue

ElementValue
    <- ConditionalExpression
    / Annotation
    / ElementValueArrayInitializer

ElementValueArrayInitializer
    <- LWING ElementValues? COMMA? RWING

ElementValues
    <- ElementValue (COMMA ElementValue)*

Spacing
     <- ( [ %nl]+            
      / '/*' (!'*/' .)* '*/'    
      / '//' (![%nl] .)* [%nl] 
      )*

Identifier <- !Keyword Letter LetterOrDigit* Spacing

Letter <- [a-z] / [A-Z] / [_$]

LetterOrDigit <- [a-z] / [A-Z] / [0-9] / [_$]

Keyword
   <- ( 'abstract'
      / 'assert'
      / 'boolean'
      / 'break'
      / 'byte'
      / 'case'
      / 'catch'
      / 'char'
      / 'class'
      / 'const'
      / 'continue'
      / 'default'
      / 'double'
      / 'do'
      / 'else'
      / 'enum'
      / 'extends'
      / 'false'
      / 'finally'
      / 'final'
      / 'float'
      / 'for'
      / 'goto'
      / 'if'
      / 'implements'
      / 'import'
      / 'interface'
      / 'int'
      / 'instanceof'
      / 'long'
      / 'native'
      / 'new'
      / 'null'
      / 'package'
      / 'private'
      / 'protected'
      / 'public'
      / 'return'
      / 'short'
      / 'static'
      / 'strictfp'
      / 'super'
      / 'switch'
      / 'synchronized'
      / 'this'
      / 'throws'
      / 'throw'
      / 'transient'
      / 'true'
      / 'try'
      / 'void'
      / 'volatile'
      / 'while'
      ) !LetterOrDigit

ASSERT       <- 'assert'       !LetterOrDigit Spacing
BREAK        <- 'break'        !LetterOrDigit Spacing
CASE         <- 'case'         !LetterOrDigit Spacing
CATCH        <- 'catch'        !LetterOrDigit Spacing
CLASS        <- 'class'        !LetterOrDigit Spacing
CONTINUE     <- 'continue'     !LetterOrDigit Spacing
DEFAULT      <- 'default'      !LetterOrDigit Spacing
DO           <- 'do'           !LetterOrDigit Spacing
ELSE         <- 'else'         !LetterOrDigit Spacing
ENUM         <- 'enum'         !LetterOrDigit Spacing
EXTENDS      <- 'extends'      !LetterOrDigit Spacing
FINALLY      <- 'finally'      !LetterOrDigit Spacing
FINAL        <- 'final'        !LetterOrDigit Spacing
FOR          <- 'for'          !LetterOrDigit Spacing
IF           <- 'if'           !LetterOrDigit Spacing
IMPLEMENTS   <- 'implements'   !LetterOrDigit Spacing
IMPORT       <- 'import'       !LetterOrDigit Spacing
INTERFACE    <- 'interface'    !LetterOrDigit Spacing
INSTANCEOF   <- 'instanceof'   !LetterOrDigit Spacing
NEW          <- 'new'          !LetterOrDigit Spacing
PACKAGE      <- 'package'      !LetterOrDigit Spacing
RETURN       <- 'return'       !LetterOrDigit Spacing
STATIC       <- 'static'       !LetterOrDigit Spacing
SUPER        <- 'super'        !LetterOrDigit Spacing
SWITCH       <- 'switch'       !LetterOrDigit Spacing
SYNCHRONIZED <- 'synchronized' !LetterOrDigit Spacing
THIS         <- 'this'         !LetterOrDigit Spacing
THROWS       <- 'throws'       !LetterOrDigit Spacing
THROW        <- 'throw'        !LetterOrDigit Spacing
TRY          <- 'try'          !LetterOrDigit Spacing
VOID         <- 'void'         !LetterOrDigit Spacing
WHILE        <- 'while'        !LetterOrDigit Spacing

Literal
   <- ( FloatLiteral
      / IntegerLiteral          
      / CharLiteral
      / StringLiteral
      / 'true'  !LetterOrDigit
      / 'false' !LetterOrDigit
      / 'null'  !LetterOrDigit
      ) Spacing

IntegerLiteral
   <- ( HexNumeral
      / BinaryNumeral
      / OctalNumeral            
      / DecimalNumeral          
      ) [lL]?

DecimalNumeral <- '0' / [1-9] ([_]* [0-9])*

HexNumeral     <- ('0x' / '0X') HexDigits

BinaryNumeral  <- ('0b' / '0B') [01] ([_]* [01])*

OctalNumeral   <- '0' ([_]* [0-7])+

FloatLiteral   <- HexFloat / DecimalFloat

DecimalFloat
   <- Digits '.' Digits?  Exponent? [fFdD]?
    / '.' Digits Exponent? [fFdD]?
    / Digits Exponent [fFdD]?
    / Digits Exponent? [fFdD]

Exponent <- [eE] [-+]? Digits

HexFloat <- HexSignificand BinaryExponent [fFdD]?

HexSignificand
   <- ('0x' / '0X') HexDigits? '.' HexDigits
    / HexNumeral '.'?                           

BinaryExponent <- [pP] [-+]? Digits

Digits <- [0-9]([_]*[0-9])*

HexDigits <- HexDigit ([_]*HexDigit)*

HexDigit <- [a-f] / [A-F] / [0-9]

CharLiteral <- ['] (Escape / !['\] .) [']

StringLiteral <- '"' (Escape / !["\%nl] .)* '"'

Escape <- '\' ([btnfr"'\] / OctalEscape / UnicodeEscape)

OctalEscape
   <- [0-3][0-7][0-7]
    / [0-7][0-7]
    / [0-7]

UnicodeEscape
   <- 'u'+ HexDigit HexDigit HexDigit HexDigit

AT              <-   '@'       Spacing
AND             <-   '&'![=&]  Spacing
ANDAND          <-   '&&'      Spacing
ANDEQU          <-   '&='      Spacing
BANG            <-   '!' !'='  Spacing
BSR             <-   '>>>' !'=' Spacing
BSREQU          <-   '>>>='    Spacing
COLON           <-   ':'       Spacing
COMMA           <-   ','       Spacing
DEC             <-   '--'      Spacing
DIV             <-   '/' !'='  Spacing
DIVEQU          <-   '/='      Spacing
DOT             <-   '.'       Spacing
ELLIPSIS        <-   '...'     Spacing
EQU             <-   '=' !'='  Spacing
EQUAL           <-   '=='      Spacing
GE              <-   '>='      Spacing
GT              <-   '>'![=>]  Spacing
HAT             <-   '^' !'='   Spacing
HATEQU          <-   '^='      Spacing
INC             <-   '++'      Spacing
LBRK            <-   '['       Spacing
LE              <-   '<='      Spacing
LPAR            <-   '('       Spacing
LPOINT          <-   '<'       Spacing
LT              <-   '<' ![=<]  Spacing
LWING           <-   '{'       Spacing
MINUS           <-   '-' ![-=] Spacing
MINUSEQU        <-   '-='      Spacing
MOD             <-   '%' !'='   Spacing
MODEQU          <-   '%='      Spacing
NOTEQUAL        <-   '!='      Spacing
OR              <-   '|' ![=|]  Spacing
OREQU           <-   '|='      Spacing
OROR            <-   '||'      Spacing
PLUS            <-   '+' ![=+]  Spacing
PLUSEQU         <-   '+='      Spacing
QUERY           <-   '?'       Spacing
RBRK            <-   ']'       Spacing
RPAR            <-   ')'       Spacing
RPOINT          <-   '>'       Spacing
RWING           <-   '}'       Spacing
SEMI            <-   ';'       Spacing
SL              <-   '<<' !'='  Spacing
SLEQU           <-   '<<='     Spacing
SR              <-   '>>' ![=>] Spacing
SREQU           <-   '>>='     Spacing
STAR            <-   '*' !'='   Spacing
STAREQU         <-   '*='      Spacing
TILDA           <-   '~'       Spacing

EOT <- !.
]])

local open = io.open

local function read_file(path)
    local file = open(path, "rb") -- r read mode and b binary mode
    if not file then return nil end
    local content = file:read "*a" -- *a or *all reads the whole file
    file:close()
    return content
end


local input = read_file(arg[1])
local sz = m:match(input)
print(sz)
-- print(string.format("elapsed time: %.5f\n", os.clock() - x))

