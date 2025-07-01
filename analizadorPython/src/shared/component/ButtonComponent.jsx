import React from "react";

function BotonComponent({ onClick, children, type = "button" }) {
  return (
    <button type={type} onClick={onClick} style={{ margin: "0 8px" }}>
      {children}
    </button>
  );
}

export default BotonComponent;