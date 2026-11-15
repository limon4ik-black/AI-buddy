import ReactDOM from "react-dom/client";
import { App } from "./App";
import "./style.css";
import "./workspace.css";

// Excalidraw по умолчанию тянет шрифты с CDN, а CSP фронта разрешает только свой origin:
// файлы лежат в public/fonts, поэтому просим забирать их оттуда.
(
  window as typeof window & { EXCALIDRAW_ASSET_PATH?: string }
).EXCALIDRAW_ASSET_PATH = "/";

ReactDOM.createRoot(document.getElementById("root")!).render(<App />);
