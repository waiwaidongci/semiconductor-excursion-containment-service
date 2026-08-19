import { cpSync, mkdirSync } from "node:fs";
mkdirSync("dist", { recursive: true });
cpSync("src/index.html", "dist/index.html");
cpSync("src/app.js", "dist/app.js");
cpSync("src/styles.css", "dist/styles.css");
