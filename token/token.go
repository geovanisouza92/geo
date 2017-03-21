package token

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Col     int
}

//go:generate stringer -type=TokenType

type TokenType byte

const (
	Error TokenType = iota
	EOF

	// Identifiers + literals
	Id
	Number
	String

	// Operators
	Assign // =
	Plus   // +
	Minus  // -
	Mul    // *
	Div    // /
	Not    // !
	Eq     // ==
	Neq    // !=
	Gt     // >
	Ge     // >=
	Lt     // <
	Le     // <=
	Pipe   // |
	And    // &&
	Or     // ||

	// Delimiters
	EOL      // ;
	Comma    // ,
	Colon    // :
	LParen   // (
	RParen   // )
	LBrace   // {
	RBrace   // }
	LBracket // [
	RBracket // ]

	// Keywords
	Fn
	Let
	Return
	True
	False
	If
	Else
)

var keywords = map[string]TokenType{
	"fn":     Fn,
	"let":    Let,
	"return": Return,
	"true":   True,
	"false":  False,
	"if":     If,
	"else":   Else,
}

func LookupId(id string) TokenType {
	if t, ok := keywords[id]; ok {
		return t
	}
	return Id
}
