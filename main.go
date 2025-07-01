package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"examencorte2/src/lexer"
	"examencorte2/src/parser"
	"examencorte2/src/semantico"
)

type AnalysisRequest struct {
	Code string `json:"code"`
}

type AnalysisResponse struct {
	LexicalAnalysis  lexer.LexicalResult    `json:"lexical_analysis"`
	SyntaxAnalysis   parser.SyntaxResult    `json:"syntax_analysis"`
	SemanticAnalysis semantico.SemanticResult `json:"semantic_analysis"`
	Success         bool                    `json:"success"`
	Error           string                  `json:"error,omitempty"`
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

func analyzeCode(w http.ResponseWriter, r *http.Request) {
    enableCORS(w)

    if r.Method == "OPTIONS" {
        w.WriteHeader(http.StatusOK)
        return
    }

    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var req AnalysisRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Léxico: Convierte el código fuente en tokens.
    lexicalResult := lexer.Analyze(req.Code)

    // Sintáctico: Verifica que los tokens sigan una estructura gramática válida.
    syntaxResult := parser.Analyze(lexicalResult.Tokens)

    // Semántico: Verifica el significado: tipos correctos, operaciones válidas, etc.
    semanticResult := semantico.Analyze(lexicalResult.Tokens, syntaxResult.AST)

    response := AnalysisResponse{
        LexicalAnalysis:  lexicalResult,
        SyntaxAnalysis:   syntaxResult,
        SemanticAnalysis: semanticResult,
        Success:         len(syntaxResult.Errors) == 0 && len(semanticResult.Errors) == 0,
    }

    if len(syntaxResult.Errors) > 0 {
        response.Error = fmt.Sprintf("Errores de sintaxis: %v", syntaxResult.Errors)
    } else if len(semanticResult.Errors) > 0 {
        response.Error = fmt.Sprintf("Errores semánticos: %v", semanticResult.Errors)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/analyze", analyzeCode)
	
	fmt.Println("Servidor iniciado en puerto 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}