// Code generated by "stringer -type=TokenType"; DO NOT EDIT

package token

import "fmt"

const _TokenType_name = "ErrorEOFIdNumberStringAssignPlusMinusMulDivNotEqNeqGtGeLtLePipeAndOrEOLCommaColonLParenRParenLBraceRBraceLBracketRBracketFnLetReturnTrueFalseIfElse"

var _TokenType_index = [...]uint8{0, 5, 8, 10, 16, 22, 28, 32, 37, 40, 43, 46, 48, 51, 53, 55, 57, 59, 63, 66, 68, 71, 76, 81, 87, 93, 99, 105, 113, 121, 123, 126, 132, 136, 141, 143, 147}

func (i TokenType) String() string {
	if i >= TokenType(len(_TokenType_index)-1) {
		return fmt.Sprintf("TokenType(%d)", i)
	}
	return _TokenType_name[_TokenType_index[i]:_TokenType_index[i+1]]
}
