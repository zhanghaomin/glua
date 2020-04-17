package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type tokenType int

const EOF = -1

const (
	TAnd tokenType = iota
	TBreak
	TDo
	TElse
	TElseif
	TEnd
	TFalse
	TFor
	TFunction
	TGoto
	TIf
	TIn
	TLocal
	TNil
	TNot
	TOr
	TRepeat
	TRequire
	TReturn
	TThen
	TTrue
	TUntil
	TWhile
	TId          tokenType = 1<<8 + iota
	TAssign                // =
	TEq                    // ==
	TNe                    // ~=
	TGt                    // >
	TLt                    // <
	TGte                   // >=
	TLte                   // <=
	TMinus                 // -
	TMinusAssign           // -=
	TPlus                  // +
	TPlusAssign            // +=
	TNumber
	TStr
	TLeftParent   // (
	TRightParent  // )
	TComma        // ,
	TDot          // .
	T2Dot         // .
	TColon        // :
	TOpenBrace    // {
	TCloseBrace   // }
	TLeftBracket  // [
	TRightBracket // ]
	TPound        // #
)

var (
	keywordsStr2Token, tokenName = map[string]tokenType{
		"and":      TAnd,
		"break":    TBreak,
		"do":       TDo,
		"else":     TElse,
		"elseif":   TElseif,
		"end":      TEnd,
		"false":    TFalse,
		"for":      TFor,
		"function": TFunction,
		"goto":     TGoto,
		"if":       TIf,
		"in":       TIn,
		"local":    TLocal,
		"nil":      TNil,
		"not":      TNot,
		"or":       TOr,
		"repeat":   TRepeat,
		"require":  TRequire,
		"return":   TReturn,
		"then":     TThen,
		"true":     TTrue,
		"until":    TUntil,
		"while":    TWhile,
	}, map[tokenType]string{
		TId:           "id",
		TAssign:       "assign",      // =
		TEq:           "eq",          // ==
		TNe:           "ne",          // ~=
		TGt:           "gt",          // >
		TLt:           "lt",          // <
		TGte:          "gte",         // >=
		TLte:          "lte",         // <=
		TMinus:        "minus",       // -
		TMinusAssign:  "minusAssign", // -=
		TPlus:         "plus",        // +
		TPlusAssign:   "plusAssign",  // +
		TNumber:       "number",
		TStr:          "str",
		TLeftParent:   "leftParenthesis",  // (
		TRightParent:  "rightParenthesis", // )
		TComma:        "comma",            // ,
		TDot:          "dot",              // .
		T2Dot:         "2Dot",             // ..
		TColon:        "colon",            // :
		TOpenBrace:    "openBrace",        // {
		TCloseBrace:   "closeBrace",       // }
		TLeftBracket:  "leftBracket",      // [
		TRightBracket: "rightBracket",     // ]
		TPound:        "pound",            // #
	}
)

type Error struct {
	pos Position
	msg string
}

type Position struct {
	line     int
	column   int
	fileName string
}

type Token struct {
	pos Position
	typ tokenType
	val string
}

type Lexer struct {
	pos          Position
	src          *bufio.Reader
	prevToken    *Token
	currentToken *Token
}

func (e *Error) String() string {
	return fmt.Sprintf("file: %s line: %d(column: %d) %s", e.pos.fileName, e.pos.line, e.pos.column, e.msg)
}

func (l *Lexer) Scan() (*Token, *Error) {
	l.prevToken = l.currentToken
retry:
	c := l.readNext()

	switch c {
	case '-':
		if l.peek() == '-' {
			l.readNext()
			l.skipComment(c)
			goto retry
		} else if l.peek() == '=' {
			l.readNext()
			l.currentToken = l.makeToken(TMinusAssign, "", 2)
		} else {
			l.currentToken = l.makeToken(TMinus, "", 1)
		}
	case '=':
		if l.peek() == '=' {
			l.readNext()
			l.currentToken = l.makeToken(TEq, "", 2)
		} else {
			l.currentToken = l.makeToken(TAssign, "", 1)
		}
	case '>':
		if l.peek() == '=' {
			l.readNext()
			l.currentToken = l.makeToken(TGte, "", 2)
		} else {
			l.currentToken = l.makeToken(TGt, "", 1)
		}
	case '<':
		if l.peek() == '=' {
			l.readNext()
			l.currentToken = l.makeToken(TLte, "", 2)
		} else {
			l.currentToken = l.makeToken(TLt, "", 1)
		}
	case '(':
		l.currentToken = l.makeToken(TLeftParent, "", 1)
	case ')':
		l.currentToken = l.makeToken(TRightParent, "", 1)
	case '+':
		if l.peek() == '=' {
			l.readNext()
			l.currentToken = l.makeToken(TPlusAssign, "", 2)
		} else {
			l.currentToken = l.makeToken(TPlus, "", 1)
		}
	case ',':
		l.currentToken = l.makeToken(TComma, "", 1)
	case '\'':
		fallthrough
	case '"':
		l.matchString(c)
	case '.':
		if l.peek() == '.' {
			l.readNext()
			l.currentToken = l.makeToken(T2Dot, "", 2)
		} else {
			l.currentToken = l.makeToken(TDot, "", 1)
		}
	case ':':
		l.currentToken = l.makeToken(TColon, "", 1)
	case '{':
		l.currentToken = l.makeToken(TOpenBrace, "", 1)
	case '}':
		l.currentToken = l.makeToken(TCloseBrace, "", 1)
	case '[':
		if l.peek() == '[' || l.peek() == '=' {
			return l.matchString(c)
		} else {
			l.currentToken = l.makeToken(TLeftBracket, "", 1)
		}
	case ']':
		l.currentToken = l.makeToken(TRightBracket, "", 1)
	case '#':
		l.currentToken = l.makeToken(TPound, "", 1)
	case '~':
		if l.peek() != '=' {
			goto err
		}
		l.currentToken = l.makeToken(TNe, "", 2)
	case '\n':
		l.newLine()
		fallthrough
	case ' ', '\t', '\r':
		goto retry
	case EOF:
		goto eof
	default:
		switch {
		case unicode.IsLetter(rune(c)) || c == '_':
			l.keywordOrId(c)
		case unicode.IsNumber(rune(c)):
			l.matchNumber(c)
		default:
			goto err
		}
	}

	return l.currentToken, nil
eof:
	return nil, &Error{
		pos: Position{
			line:     l.pos.line,
			column:   l.pos.column - 1,
			fileName: l.pos.fileName,
		},
		msg: "reach end",
	}
err:
	return nil, &Error{
		pos: Position{
			line:     l.pos.line,
			column:   l.pos.column - 1,
			fileName: l.pos.fileName,
		},
		msg: "unknown token " + string(c),
	}
}

func (l *Lexer) skipComment(first int) {
	var c int

	if l.peek() == '[' {
		// 找到第二个[
		c = l.readNext()
		openTag := string(c)
		str := ""
		var c int

		for c = l.peek(); c == '='; c = l.peek() {
			l.readNext()
			openTag += string('=')
		}

		if c != '[' {
			goto common
		}

		openTag += string('[')                            // [=*[
		closeTag := strings.ReplaceAll(openTag, "[", "]") // ]=*]

		// 寻找close ]=]==]
		for c = l.peek(); c != EOF && !strings.Contains(str, closeTag); c = l.peek() {
			l.readNext()
			if c == '\n' {
				l.newLine()
			}
			str += string(c)
		}
	}
common:
	for c = l.readNext(); c != EOF && c != '\n'; c = l.readNext() {
		// 跳过所有字符，直到换行符
	}

	if c == '\n' {
		l.newLine()
	}
}

func (l *Lexer) matchString(first int) (*Token, *Error) {
	if first == '\'' || first == '"' {
		escape := false
		str := ""

		// 找到下一个同类字符
		for c := l.readNext(); c != EOF; c = l.readNext() {
			if escape {
				escape = false

				switch c {
				case '\\':
					str += "\\"
				case 'a':
					str += "\a"
				case 'b':
					str += "\b"
				case 'f':
					str += "\f"
				case 'n':
					str += "\n"
				case 'r':
					str += "\r"
				case 't':
					str += "\t"
				case 'v':
					str += "\v"
				case '0':
					str += "\x00"
				case c:
					str += string(c)
				case '\n':
					str += "\n"
					l.newLine()
				default:
					return nil, &Error{
						pos: Position{
							l.pos.line,
							l.pos.column - 1,
							l.pos.fileName,
						},
						msg: "invalid escape sequence",
					}
				}

				continue
			}

			if c == '\\' {
				escape = true
			} else if c == first {
				l.currentToken = l.makeToken(TStr, str, 0) // 跨行token位置以结束位置为准，不然不好算
				break
			} else if c == '\n' {
				return nil, &Error{
					pos: l.pos,
					msg: "字符串不能跨行",
				}
			} else {
				str += string(c)
			}
		}
	} else { // [[ ]]  [===[ ]===]
		// 找到第二个[
		openTag := string(first)
		str := ""
		var c int

		for c = l.readNext(); c == '='; c = l.readNext() {
			openTag += string('=')
		}

		if c != '[' {
			return nil, &Error{
				pos: l.pos,
				msg: "字符串不合法",
			}
		}

		openTag += string('[')                            // [=*[
		closeTag := strings.ReplaceAll(openTag, "[", "]") // ]=*]

		// 如果后面紧跟一个换行，忽略这个换行符
		if l.peek() == '\n' {
			l.readNext()
			l.newLine()
		}

		// 寻找close ]=]==]
		for c = l.peek(); c != EOF && !strings.Contains(str, closeTag); c = l.peek() {
			l.readNext()
			if c == '\n' {
				l.newLine()
			}
			str += string(c)
		}

		if !strings.Contains(str, closeTag) {
			return nil, &Error{
				pos: l.pos,
				msg: "reach end",
			}
		}

		str = strings.TrimSuffix(str, closeTag)
		l.currentToken = l.makeToken(TStr, str, 0)
	}

	return l.currentToken, nil
}

func (l *Lexer) matchNumber(first int) {
	str := string(first)

	for c := l.peek(); unicode.IsNumber(rune(c)); c = l.peek() {
		l.readNext()
		str += string(c)
	}

	l.currentToken = l.makeToken(TStr, str, len(str))
}

func (l *Lexer) keywordOrId(first int) {
	str := string(first)

	for c := l.peek(); unicode.IsLetter(rune(c)); c = l.peek() {
		l.readNext()
		str += string(c)
	}

	if typ, ok := keywordsStr2Token[str]; ok {
		l.currentToken = l.makeToken(typ, "", len(str))
	} else {
		l.currentToken = l.makeToken(TId, str, len(str))
	}
}

func (l *Lexer) makeToken(typ tokenType, val string, tokenLen int) *Token {
	return &Token{
		pos: Position{
			line:     l.pos.line,
			column:   l.pos.column - tokenLen,
			fileName: l.pos.fileName,
		},
		typ: typ,
		val: val,
	}
}

func InitLexer(src *bufio.Reader, fileName string) *Lexer {
	l := Lexer{}
	l.src = src
	l.pos = Position{
		line:     1,
		column:   1,
		fileName: fileName,
	}
	return &l
}

func (l *Lexer) peek() int {
	if c, err := l.src.ReadByte(); err != io.EOF {
		_ = l.src.UnreadByte()
		return int(c)
	}

	return EOF
}

func (l *Lexer) newLine() {
	l.pos.column = 1
	l.pos.line++
}

func (l *Lexer) readNext() int {
	if c, err := l.src.ReadByte(); err != io.EOF {
		l.pos.column++
		return int(c)
	}

	return EOF
}
