// Code generated by "stringer -type token -linecomment"; DO NOT EDIT.

package syntax

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[_EOF-1]
	_ = x[_Name-2]
	_ = x[_Literal-3]
	_ = x[_Op-4]
	_ = x[_Semi-5]
	_ = x[_Comma-6]
	_ = x[_Colon-7]
	_ = x[_Dot-8]
	_ = x[_Arrow-9]
	_ = x[_Assign-10]
	_ = x[_Ref-11]
	_ = x[_Lbrace-12]
	_ = x[_Rbrace-13]
	_ = x[_Lparen-14]
	_ = x[_Rparen-15]
	_ = x[_Lbrack-16]
	_ = x[_Rbrack-17]
	_ = x[_Break-18]
	_ = x[_Continue-19]
	_ = x[_If-20]
	_ = x[_Else-21]
	_ = x[_For-22]
	_ = x[_Match-23]
	_ = x[_Range-24]
}

const _token_name = "eofnameliteralopsemi or newline,:.->=${}))[]breakcontinueifelseformatchrange"

var _token_index = [...]uint8{0, 3, 7, 14, 16, 31, 32, 33, 34, 36, 37, 38, 39, 40, 41, 42, 43, 44, 49, 57, 59, 63, 66, 71, 76}

func (i token) String() string {
	i -= 1
	if i >= token(len(_token_index)-1) {
		return "token(" + strconv.FormatInt(int64(i+1), 10) + ")"
	}
	return _token_name[_token_index[i]:_token_index[i+1]]
}
