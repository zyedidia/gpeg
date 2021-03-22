local re = require"re"
local lpeg = require"lpeg"

local m = re.compile([[
doc           <- JSON !.
JSON          <- S_ (Number / Object / Array / String / True / False / Null) S_
Object        <- '{' (String ':' JSON (',' String ':' JSON)* / S_) '}'
Array         <- '[' (JSON (',' JSON)* / S_) ']'
StringBody    <- Escape? ((!["\] .)+ Escape*)*
String        <- S_ '"' StringBody '"' S_
Escape        <- '\' (["{|\bfnrt] / UnicodeEscape)
UnicodeEscape <- 'u' [0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f][0-9A-Fa-f]
Number        <- Minus? IntPart FractPart? ExpPart?
Minus         <- '-'
IntPart       <- '0' / [1-9][0-9]*
FractPart     <- '.' [0-9]+
ExpPart       <- [eE] ('+' / '-')? [0-9]+
True          <- 'true'
False         <- 'false'
Null          <- 'null'
S_            <- (%nl / ' ')*
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

