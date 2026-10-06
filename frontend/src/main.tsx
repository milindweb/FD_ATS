import React from "react";
import { createRoot } from "react-dom/client";
import "./styles/index.css";
import { initTheme } from "./lib/theme";
import App from "./App";

initTheme();

const container = document.getElementById("root");

const root = createRoot(container!);

root.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
