import React from "react";

function TablaTokens({ tokens }) {
  if (!tokens || tokens.length === 0) return <div>No hay tokens</div>;

  return (
    <table border="1" cellPadding="4" style={{ marginTop: 16 }}>
      <thead>
        <tr>
          <th>Tipo</th>
          <th>Valor</th>
          <th>Línea</th>
          <th>Columna</th>
        </tr>
      </thead>
      <tbody>
        {tokens.map((tok, idx) => (
          <tr key={idx}>
            <td>{tok.type}</td>
            <td>{tok.value}</td>
            <td>{tok.line}</td>
            <td>{tok.column}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}

export default TablaTokens;