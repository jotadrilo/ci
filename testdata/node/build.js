import { mkdirSync, writeFileSync } from "node:fs";
import { add } from "./index.js";

mkdirSync("dist", { recursive: true });
writeFileSync("dist/out.txt", `${add(2, 3)}\n`);
console.log("built dist/out.txt");
