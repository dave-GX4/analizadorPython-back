import { useState } from "react";
import TablaTokens from "./shared/component/TablaTokens";
import BotonComponent from "./shared/component/ButtonComponent";
import SpanComponent from "./shared/component/SpanComponent";
import "./App.css";

function App() {
  const [code, setCode] = useState("");
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);

  const analizar = async () => {
    setLoading(true);
    setResult(null);
    try {
      const res = await fetch("http://localhost:8080/analyze", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
      });
      const data = await res.json();
      setResult(data);
    } catch{
      setResult({ error: "Error de conexión con el backend" });
    }
    setLoading(false);
  };

  return (
    <div className="app-container" style={{ padding: 24, maxWidth: 900 }}>
      <h2>Analizador Python</h2>
      <SpanComponent style={{ fontWeight: "bold" }}>Código a analizar:</SpanComponent>
      <br />
      <textarea
        rows={8}
        cols={70}
        value={code}
        onChange={e => setCode(e.target.value)}
        placeholder="Escribe tu código Python aquí..."
        style={{ margin: "8px 0", fontFamily: "monospace" }}
      />
      <br />
      <BotonComponent onClick={analizar} children={'analizar'}/>
      <BotonComponent onClick={() => setCode("")} children='Limpiar'/>
      {loading && <SpanComponent style={{ color: "blue" }} children={'Analizando...'}/>}
      {result && (
        <div style={{ marginTop: 24 }}>
          {result.error && (
            <SpanComponent style={{ color: "red" }}>{result.error}</SpanComponent>
          )}
          {result.lexical_analysis && (
            <>
              <h3>Tokens</h3>
              <TablaTokens tokens={result.lexical_analysis.tokens} />
              <h4>Estadísticas</h4>
              <ul>
                <li>Palabras reservadas: {result.lexical_analysis.statistics.keywords}</li>
                <li>Identificadores: {result.lexical_analysis.statistics.identifiers}</li>
                <li>Números: {result.lexical_analysis.statistics.numbers}</li>
                <li>Strings: {result.lexical_analysis.statistics.strings}</li>
                <li>Errores: {result.lexical_analysis.statistics.errors}</li>
              </ul>
            </>
          )}
          {result.syntax_analysis && (
            <>
              <h3>Análisis Sintáctico</h3>
              <SpanComponent>
                {result.syntax_analysis.errors && result.syntax_analysis.errors.length > 0
                  ? result.syntax_analysis.errors.join(", ")
                  : "Sin errores de sintaxis"}
              </SpanComponent>
            </>
          )}
          {result.semantic_analysis && (
            <>
              <h3>Análisis Semántico</h3>
              <SpanComponent>
                {result.semantic_analysis.errors && result.semantic_analysis.errors.length > 0
                  ? result.semantic_analysis.errors.join(", ")
                  : "Sin errores semánticos"}
              </SpanComponent>
            </>
          )}
        </div>
      )}
    </div>
  );
}

export default App;
