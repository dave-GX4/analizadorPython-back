import React from "react";

function InputComponent({ value, onChange, placeholder, ...props }) {
  return (
    <input
      value={value}
      onChange={onChange}
      placeholder={placeholder}
      {...props}
      style={{ padding: "4px", margin: "0 8px" }}
    />
  );
}

export default InputComponent;