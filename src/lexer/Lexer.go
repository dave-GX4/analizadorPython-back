package lexer

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	KEYWORD TokenType = iota
	IDENTIFIER
	NUMBER
	STRING
	SYMBOL
	WHITESPACE
	NEWLINE
	ERROR
)

type Token struct {
	Type    TokenType `json:"type"`
	Value   string    `json:"value"`
	Line    int       `json:"line"`
	Column  int       `json:"column"`
}

type LexicalResult struct {
	Tokens          []Token             `json:"tokens"`
	Table          map[string][]string `json:"table"`
	Statistics     TokenStatistics     `json:"statistics"`
	Errors         []string            `json:"errors"`
	ReservedWords  int                 `json:"reserved_words"`
}

type TokenStatistics struct {
	Keywords    int `json:"keywords"`
	Identifiers int `json:"identifiers"`
	Numbers     int `json:"numbers"`
	Strings     int `json:"strings"`
	Symbols     int `json:"symbols"`
	Errors      int `json:"errors"`
}

var pythonKeywords = map[string]bool{
    "def": true, "if": true, "else": true, "elif": true, "while": true,
    "for": true, "in": true, "try": true, "except": true, "finally": true,
    "with": true, "as": true, "pass": true, "break": true, "continue": true,
    "return": true, "yield": true, "import": true, "from": true, "class": true,
    "and": true, "or": true, "not": true, "is": true, "lambda": true,
    "None": true, "True": true, "False": true, "print": true,
}

var pythonSymbols = []string{
	"==", "!=", "<=", ">=", ">>", "<<", "**", "//", "+=", "-=", "*=", "/=",
	"=", "+", "-", "*", "/", "%", "<", ">", "(", ")", "[", "]", "{", "}",
	":", ";", ",", ".", "&", "|", "^", "~", "!", "@", "#", "$", "?",
}

func Analyze(code string) LexicalResult {
	result := LexicalResult{
		Tokens: []Token{},
		Table: map[string][]string{
			"PR":      {},
			"ID":      {},
			"Numeros": {},
			"Simbolos": {},
			"Error":   {},
		},
		Statistics: TokenStatistics{},
		Errors:     []string{},
	}

	lines := strings.Split(code, "\n")
	
	for lineNum, line := range lines {
		result.processLine(line, lineNum+1)
	}

	result.ReservedWords = result.Statistics.Keywords
	return result
}

func (r *LexicalResult) processLine(line string, lineNum int) {
	i := 0
	column := 1
	
	for i < len(line) {
		char := rune(line[i])
		
		// Espacios en blanco
		if unicode.IsSpace(char) {
			i++
			column++
			continue
		}
		
		// Comentarios
		if char == '#' {
			break
		}
		
		// Strings
		if char == '"' || char == '\'' {
			token, length := r.processString(line[i:], lineNum, column, char)
			r.addToken(token)
			i += length
			column += length
			continue
		}
		
		// Números
		if unicode.IsDigit(char) {
			token, length := r.processNumber(line[i:], lineNum, column)
			r.addToken(token)
			i += length
			column += length
			continue
		}
		
		// Identificadores y palabras reservadas
		if unicode.IsLetter(char) || char == '_' {
			token, length := r.processIdentifier(line[i:], lineNum, column)
			r.addToken(token)
			i += length
			column += length
			continue
		}
		
		// Símbolos
		token, length := r.processSymbol(line[i:], lineNum, column)
		if token.Type == ERROR {
			r.Errors = append(r.Errors, 
				fmt.Sprintf("Carácter no reconocido '%c' en línea %d, columna %d", 
					char, lineNum, column))
		}
		r.addToken(token)
		i += length
		column += length
	}
}

func (r *LexicalResult) processString(text string, line, column int, quote rune) (Token, int) {
	i := 1
	for i < len(text) && rune(text[i]) != quote {
		if text[i] == '\\' && i+1 < len(text) {
			i += 2
		} else {
			i++
		}
	}
	
	if i >= len(text) {
		return Token{
			Type:   ERROR,
			Value:  text,
			Line:   line,
			Column: column,
		}, len(text)
	}
	
	return Token{
		Type:   STRING,
		Value:  text[:i+1],
		Line:   line,
		Column: column,
	}, i + 1
}

func (r *LexicalResult) processNumber(text string, line, column int) (Token, int) {
	i := 0
	hasDecimal := false
	
	for i < len(text) && (unicode.IsDigit(rune(text[i])) || text[i] == '.') {
		if text[i] == '.' {
			if hasDecimal {
				break
			}
			hasDecimal = true
		}
		i++
	}
	
	return Token{
		Type:   NUMBER,
		Value:  text[:i],
		Line:   line,
		Column: column,
	}, i
}

func (r *LexicalResult) processIdentifier(text string, line, column int) (Token, int) {
	i := 0
	for i < len(text) && (unicode.IsLetter(rune(text[i])) || 
		                 unicode.IsDigit(rune(text[i])) || 
						 text[i] == '_') {
		i++
	}
	
	value := text[:i]
	tokenType := IDENTIFIER
	
	if pythonKeywords[value] {
		tokenType = KEYWORD
	}
	
	return Token{
		Type:   tokenType,
		Value:  value,
		Line:   line,
		Column: column,
	}, i
}

func (r *LexicalResult) processSymbol(text string, line, column int) (Token, int) {
	// Verificar símbolos de dos caracteres primero
	if len(text) >= 2 {
		twoChar := text[:2]
		for _, symbol := range pythonSymbols {
			if symbol == twoChar {
				return Token{
					Type:   SYMBOL,
					Value:  twoChar,
					Line:   line,
					Column: column,
				}, 2
			}
		}
	}
	
	// Verificar símbolos de un carácter
	oneChar := string(text[0])
	for _, symbol := range pythonSymbols {
		if symbol == oneChar {
			return Token{
				Type:   SYMBOL,
				Value:  oneChar,
				Line:   line,
				Column: column,
			}, 1
		}
	}
	
	// Carácter no reconocido
	return Token{
		Type:   ERROR,
		Value:  oneChar,
		Line:   line,
		Column: column,
	}, 1
}

func (r *LexicalResult) addToken(token Token) {
	r.Tokens = append(r.Tokens, token)
	
	switch token.Type {
	case KEYWORD:
		r.Table["PR"] = append(r.Table["PR"], token.Value)
		r.Statistics.Keywords++
	case IDENTIFIER:
		r.Table["ID"] = append(r.Table["ID"], token.Value)
		r.Statistics.Identifiers++
	case NUMBER:
		r.Table["Numeros"] = append(r.Table["Numeros"], token.Value)
		r.Statistics.Numbers++
	case STRING:
		// Los strings no se incluyen en la tabla como en tu ejemplo
		r.Statistics.Strings++
	case SYMBOL:
		r.Table["Simbolos"] = append(r.Table["Simbolos"], token.Value)
		r.Statistics.Symbols++
	case ERROR:
		r.Table["Error"] = append(r.Table["Error"], token.Value)
		r.Statistics.Errors++
	}
}