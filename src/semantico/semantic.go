package semantico

import (
	"examencorte2/src/lexer"
	"examencorte2/src/parser"
	"fmt"
	"strings"
)

type VarType int

const (
	IntType VarType = iota
	StringType
	BoolType
	UnknownType
)

type Variable struct {
	Name string
	Type VarType
	Line int
}

type SemanticResult struct {
	Errors           []string              `json:"errors"`
	Variables        map[string]Variable   `json:"variables"`
	TypeMismatches   []string              `json:"type_mismatches"`
	Success          bool                  `json:"success"`
}

type SemanticAnalyzer struct {
	variables map[string]Variable
	errors    []string
	tokens    []lexer.Token
}

func Analyze(tokens []lexer.Token, ast *parser.ASTNode) SemanticResult {
	analyzer := &SemanticAnalyzer{
		variables: make(map[string]Variable),
		errors:    []string{},
		tokens:    tokens,
	}
	
	if ast != nil {
		analyzer.analyzeNode(ast)
	}
	
	return SemanticResult{
		Errors:         analyzer.errors,
		Variables:      analyzer.variables,
		TypeMismatches: analyzer.getTypeMismatches(),
		Success:        len(analyzer.errors) == 0,
	}
}

func (sa *SemanticAnalyzer) analyzeNode(node *parser.ASTNode) {
	if node == nil {
		return
	}
	
	switch node.Type {
	case "Program":
		for _, child := range node.Children {
			sa.analyzeNode(child)
		}
		
	case "FunctionDef":
		// Analizar parámetros y cuerpo de función
		for _, child := range node.Children {
			sa.analyzeNode(child)
		}
		
	case "Assignment":
		sa.analyzeAssignment(node)
		
	case "IfStatement":
		sa.analyzeIfStatement(node)
		
	case "Block":
		for _, child := range node.Children {
			sa.analyzeNode(child)
		}
		
	case "ExpressionStatement":
		for _, child := range node.Children {
			sa.analyzeNode(child)
		}
		
	case "BinaryOp":
		sa.analyzeBinaryOperation(node)
		
	case "FunctionCall", "MethodCall":
		sa.analyzeFunctionCall(node)
		
	default:
		// Analizar hijos por defecto
		for _, child := range node.Children {
			sa.analyzeNode(child)
		}
	}
}

func (sa *SemanticAnalyzer) analyzeAssignment(node *parser.ASTNode) {
	varName := node.Value
	
	if len(node.Children) == 0 {
		sa.addError(node.Line, "Asignación sin valor")
		return
	}
	
	valueNode := node.Children[0]
	varType := sa.inferType(valueNode)
	
	// Registrar o actualizar variable
	sa.variables[varName] = Variable{
		Name: varName,
		Type: varType,
		Line: node.Line,
	}
	
	sa.analyzeNode(valueNode)
}

func (sa *SemanticAnalyzer) analyzeIfStatement(node *parser.ASTNode) {
	if len(node.Children) < 1 {
		sa.addError(node.Line, "Declaración if sin condición")
		return
	}
	
	condition := node.Children[0]
	sa.analyzeCondition(condition)
	
	// Analizar el resto de los hijos (bloque then, etc.)
	for _, child := range node.Children {
		sa.analyzeNode(child)
	}
}

func (sa *SemanticAnalyzer) analyzeCondition(node *parser.ASTNode) {
	if node == nil {
		return
	}
	
	if node.Type == "BinaryOp" {
		sa.analyzeBinaryOperation(node)
	} else {
		sa.analyzeNode(node)
	}
}

func (sa *SemanticAnalyzer) analyzeBinaryOperation(node *parser.ASTNode) {
	if len(node.Children) < 2 {
		sa.addError(node.Line, "Operación binaria incompleta")
		return
	}
	
	leftNode := node.Children[0]
	rightNode := node.Children[1]
	operator := node.Value
	
	leftType := sa.inferType(leftNode)
	rightType := sa.inferType(rightNode)
	
	// Verificar compatibilidad de tipos según el operador
	switch operator {
	case ">", "<", ">=", "<=":
		// Operadores de comparación numérica
		if leftType == StringType && rightType == IntType {
			sa.addError(node.Line, 
				fmt.Sprintf("No se puede comparar string con número usando '%s'", operator))
		} else if leftType == IntType && rightType == StringType {
			sa.addError(node.Line, 
				fmt.Sprintf("No se puede comparar número con string usando '%s'", operator))
		}
		
	case "==", "!=":
		// Operadores de igualdad (más permisivos pero aún verificamos algunos casos)
		if leftType == StringType && rightType == IntType {
			sa.addError(node.Line, 
				fmt.Sprintf("Comparación entre tipos incompatibles: string y número"))
		} else if leftType == IntType && rightType == StringType {
			sa.addError(node.Line, 
				fmt.Sprintf("Comparación entre tipos incompatibles: número y string"))
		}
		
	case "+", "-", "*", "/":
		// Operadores aritméticos
		if leftType == StringType || rightType == StringType {
			if operator != "+" { // + puede ser concatenación
				sa.addError(node.Line, 
					fmt.Sprintf("Operador '%s' no válido para strings", operator))
			}
		}
	}
	
	// Analizar recursivamente los nodos hijos
	sa.analyzeNode(leftNode)
	sa.analyzeNode(rightNode)
}

func (sa *SemanticAnalyzer) analyzeFunctionCall(node *parser.ASTNode) {
	// Verificar llamadas a funciones conocidas
	funcName := node.Value
	
	if strings.Contains(funcName, ".") {
		// Es una llamada a método
		parts := strings.Split(funcName, ".")
		if len(parts) == 2 {
			objectName := parts[0]
			methodName := parts[1]
			
			// Verificar si el objeto está definido
			if variable, exists := sa.variables[objectName]; exists {
				// Verificar métodos específicos según el tipo
				if variable.Type == StringType && methodName == "lower" {
					// Método válido para strings
				} else if variable.Type != StringType && methodName == "lower" {
					sa.addError(node.Line, 
						fmt.Sprintf("El método 'lower()' no está disponible para el tipo de '%s'", objectName))
				}
			} else {
				sa.addError(node.Line, 
					fmt.Sprintf("Variable '%s' no está definida", objectName))
			}
		}
	} else if funcName == "print" {
		// Verificar argumentos de print
		for _, arg := range node.Children {
			sa.analyzeNode(arg)
		}
	}
	
	// Analizar argumentos
	for _, child := range node.Children {
		sa.analyzeNode(child)
	}
}

func (sa *SemanticAnalyzer) inferType(node *parser.ASTNode) VarType {
	if node == nil {
		return UnknownType
	}
	
	switch node.Type {
	case "Number":
		return IntType
	case "String":
		return StringType
	case "Identifier":
		if variable, exists := sa.variables[node.Value]; exists {
			return variable.Type
		}
		return UnknownType
	case "BinaryOp":
		// El tipo depende del operador y operandos
		operator := node.Value
		if operator == ">" || operator == "<" || operator == ">=" || 
		   operator == "<=" || operator == "==" || operator == "!=" {
			return BoolType
		}
		// Para operadores aritméticos, inferir del contexto
		if len(node.Children) >= 2 {
			leftType := sa.inferType(node.Children[0])
			rightType := sa.inferType(node.Children[1])
			if leftType == IntType && rightType == IntType {
				return IntType
			}
			if leftType == StringType || rightType == StringType {
				return StringType
			}
		}
		return UnknownType
	case "MethodCall":
		// Inferir tipo basado en el método
		if strings.Contains(node.Value, ".lower") {
			return StringType
		}
		return UnknownType
	case "FunctionCall":
		// print no retorna valor útil para comparaciones
		return UnknownType
	default:
		return UnknownType
	}
}

func (sa *SemanticAnalyzer) addError(line int, message string) {
	sa.errors = append(sa.errors, fmt.Sprintf("Error semántico en línea %d: %s", line, message))
}

func (sa *SemanticAnalyzer) getTypeMismatches() []string {
	var mismatches []string
	
	// Buscar patrones específicos de incompatibilidad de tipos
	for _, err := range sa.errors {
		if strings.Contains(err, "comparar") || strings.Contains(err, "Comparación") {
			mismatches = append(mismatches, err)
		}
	}
	
	return mismatches
}

func (vt VarType) String() string {
	switch vt {
	case IntType:
		return "int"
	case StringType:
		return "string"
	case BoolType:
		return "bool"
	default:
		return "unknown"
	}
}