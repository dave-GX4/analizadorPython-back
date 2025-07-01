package parser

import (
	"examencorte2/src/lexer"
	"fmt"
	"strings"
)

type ASTNode struct {
	Type     string     `json:"type"`
	Value    string     `json:"value,omitempty"`
	Line     int        `json:"line"`
	Children []*ASTNode `json:"children,omitempty"`
}

type SyntaxResult struct {
	AST       *ASTNode `json:"ast"`
	Errors    []string `json:"errors"`
	Success   bool     `json:"success"`
	ErrorLine int      `json:"error_line,omitempty"`
}

type Parser struct {
	tokens   []lexer.Token
	current  int
	errors   []string
	indent   int
}

func Analyze(tokens []lexer.Token) SyntaxResult {
	// Filtrar tokens de espacios en blanco para el análisis sintáctico
	filteredTokens := filterTokens(tokens)
	
	parser := &Parser{
		tokens:  filteredTokens,
		current: 0,
		errors:  []string{},
		indent:  0,
	}
	
	ast := parser.parseProgram()
	
	return SyntaxResult{
		AST:     ast,
		Errors:  parser.errors,
		Success: len(parser.errors) == 0,
		ErrorLine: parser.getErrorLine(),
	}
}

func filterTokens(tokens []lexer.Token) []lexer.Token {
	var filtered []lexer.Token
	for _, token := range tokens {
		if token.Type != lexer.WHITESPACE && token.Type != lexer.NEWLINE {
			filtered = append(filtered, token)
		}
	}
	return filtered
}

func (p *Parser) parseProgram() *ASTNode {
	program := &ASTNode{
		Type:     "Program",
		Children: []*ASTNode{},
		Line:     1,
	}
	
	for !p.isAtEnd() {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Children = append(program.Children, stmt)
		}
	}
	
	return program
}

func (p *Parser) parseStatement() *ASTNode {
	if p.match("def") {
		return p.parseFunctionDef()
	}
	
	if p.match("if") {
		return p.parseIfStatement()
	}
	
	if p.check("print") {
		return p.parseExpressionStatement()
	}
	
	if p.checkType(lexer.IDENTIFIER) {
		return p.parseAssignmentOrExpression()
	}
	
	return p.parseExpressionStatement()
}

func (p *Parser) parseFunctionDef() *ASTNode {
	line := p.previous().Line
	
	if !p.checkType(lexer.IDENTIFIER) {
		p.error("Se esperaba nombre de función")
		return nil
	}
	
	name := p.advance().Value
	
	if !p.match("(") {
		p.error("Se esperaba '(' después del nombre de función")
		return nil
	}
	
	params := []*ASTNode{}
	if !p.check(")") {
		for {
			if !p.checkType(lexer.IDENTIFIER) {
				p.error("Se esperaba nombre de parámetro")
				break
			}
			param := &ASTNode{
				Type:  "Parameter",
				Value: p.advance().Value,
				Line:  p.previous().Line,
			}
			params = append(params, param)
			
			if !p.match(",") {
				break
			}
		}
	}
	
	if !p.match(")") {
		p.error("Se esperaba ')' después de los parámetros")
		return nil
	}
	
	if !p.match(":") {
		p.error("Se esperaba ':' después de la definición de función")
		return nil
	}
	
	body := p.parseBlock()
	
	return &ASTNode{
		Type:  "FunctionDef",
		Value: name,
		Line:  line,
		Children: append(params, body),
	}
}

func (p *Parser) parseIfStatement() *ASTNode {
	line := p.previous().Line
	
	condition := p.parseExpression()
	if condition == nil {
		return nil
	}
	
	if !p.match(":") {
		p.error("Se esperaba ':' después de la condición if")
		return nil
	}
	
	thenBranch := p.parseBlock()
	
	ifNode := &ASTNode{
		Type:     "IfStatement",
		Line:     line,
		Children: []*ASTNode{condition, thenBranch},
	}
	
	return ifNode
}

func (p *Parser) parseBlock() *ASTNode {
    block := &ASTNode{
        Type:     "Block",
        Children: []*ASTNode{},
        Line:     p.peek().Line,
    }

    for !p.isAtEnd() && !p.check("def") && !p.check("if") &&
        !p.checkNext("def") && !p.checkNext("if") {
        stmt := p.parseStatement()
        if stmt != nil {
            block.Children = append(block.Children, stmt)
        } else {
            // Avanza para evitar ciclo infinito si stmt es nil
            p.advance()
        }
        if p.current >= len(p.tokens)-1 {
            break
        }
    }

    return block
}

func (p *Parser) parseAssignmentOrExpression() *ASTNode {
	if p.current + 1 < len(p.tokens) && p.tokens[p.current + 1].Value == "=" {
		return p.parseAssignment()
	}
	return p.parseExpressionStatement()
}

func (p *Parser) parseAssignment() *ASTNode {
	line := p.peek().Line
	
	if !p.checkType(lexer.IDENTIFIER) {
		p.error("Se esperaba identificador en asignación")
		return nil
	}
	
	name := p.advance().Value
	
	if !p.match("=") {
		p.error("Se esperaba '=' en asignación")
		return nil
	}
	
	value := p.parseExpression()
	if value == nil {
		return nil
	}
	
	return &ASTNode{
		Type:  "Assignment",
		Value: name,
		Line:  line,
		Children: []*ASTNode{value},
	}
}

func (p *Parser) parseExpressionStatement() *ASTNode {
	expr := p.parseExpression()
	if expr == nil {
		return nil
	}
	
	return &ASTNode{
		Type:     "ExpressionStatement",
		Line:     expr.Line,
		Children: []*ASTNode{expr},
	}
}

func (p *Parser) parseExpression() *ASTNode {
	return p.parseComparison()
}

func (p *Parser) parseComparison() *ASTNode {
	expr := p.parseTerm()
	
	for p.match(">", "<", ">=", "<=", "==", "!=") {
		operator := p.previous().Value
		right := p.parseTerm()
		expr = &ASTNode{
			Type:     "BinaryOp",
			Value:    operator,
			Line:     expr.Line,
			Children: []*ASTNode{expr, right},
		}
	}
	
	return expr
}

func (p *Parser) parseTerm() *ASTNode {
	expr := p.parseFactor()
	
	for p.match("+", "-") {
		operator := p.previous().Value
		right := p.parseFactor()
		expr = &ASTNode{
			Type:     "BinaryOp",
			Value:    operator,
			Line:     expr.Line,
			Children: []*ASTNode{expr, right},
		}
	}
	
	return expr
}

func (p *Parser) parseFactor() *ASTNode {
	if p.match("(") {
		expr := p.parseExpression()
		if !p.match(")") {
			p.error("Se esperaba ')' después de la expresión")
		}
		return expr
	}
	
	if p.checkType(lexer.NUMBER) {
		return &ASTNode{
			Type:  "Number",
			Value: p.advance().Value,
			Line:  p.previous().Line,
		}
	}
	
	if p.checkType(lexer.STRING) {
		return &ASTNode{
			Type:  "String",
			Value: p.advance().Value,
			Line:  p.previous().Line,
		}
	}
	
	if p.checkType(lexer.IDENTIFIER) {
		name := p.advance().Value
		
		// Verificar si es una llamada a función
		if p.match("(") {
			args := []*ASTNode{}
			if !p.check(")") {
				for {
					arg := p.parseExpression()
					if arg != nil {
						args = append(args, arg)
					}
					if !p.match(",") {
						break
					}
				}
			}
			
			if !p.match(")") {
				p.error("Se esperaba ')' después de los argumentos")
			}
			
			return &ASTNode{
				Type:     "FunctionCall",
				Value:    name,
				Line:     p.previous().Line,
				Children: args,
			}
		}
		
		// Verificar acceso a atributo/método
		if p.match(".") {
			if !p.checkType(lexer.IDENTIFIER) {
				p.error("Se esperaba nombre de método después de '.'")
				return nil
			}
			
			method := p.advance().Value
			
			if p.match("(") {
				args := []*ASTNode{}
				if !p.check(")") {
					for {
						arg := p.parseExpression()
						if arg != nil {
							args = append(args, arg)
						}
						if !p.match(",") {
							break
						}
					}
				}
				
				if !p.match(")") {
					p.error("Se esperaba ')' después de los argumentos del método")
				}
				
				return &ASTNode{
					Type:  "MethodCall",
					Value: fmt.Sprintf("%s.%s", name, method),
					Line:  p.previous().Line,
					Children: args,
				}
			}
		}
		
		return &ASTNode{
			Type:  "Identifier",
			Value: name,
			Line:  p.previous().Line,
		}
	}
	
	p.error("Se esperaba expresión")
	return nil
}

// Métodos auxiliares
func (p *Parser) match(types ...string) bool {
	for _, t := range types {
		if p.check(t) {
			p.advance()
			return true
		}
	}
	return false
}

func (p *Parser) check(tokenValue string) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Value == tokenValue
}

func (p *Parser) checkType(tokenType lexer.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == tokenType
}

func (p *Parser) checkNext(tokenValue string) bool {
	if p.current + 1 >= len(p.tokens) {
		return false
	}
	return p.tokens[p.current + 1].Value == tokenValue
}

func (p *Parser) advance() lexer.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens)
}

func (p *Parser) peek() lexer.Token {
	if p.isAtEnd() {
		return lexer.Token{}
	}
	return p.tokens[p.current]
}

func (p *Parser) previous() lexer.Token {
	if p.current == 0 {
		return lexer.Token{}
	}
	return p.tokens[p.current-1]
}

func (p *Parser) error(message string) {
	line := 1
	if !p.isAtEnd() {
		line = p.peek().Line
	}
	p.errors = append(p.errors, fmt.Sprintf("Error en línea %d: %s", line, message))
}

func (p *Parser) getErrorLine() int {
	if len(p.errors) == 0 {
		return 0
	}
	// Extraer número de línea del primer error
	errorMsg := p.errors[0]
	if strings.Contains(errorMsg, "línea ") {
		var line int
		fmt.Sscanf(errorMsg, "Error en línea %d:", &line)
		return line
	}
	return 0
}