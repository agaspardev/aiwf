// setup-safety-gates.mjs — Escribe fidelityGate, riskGate y pipelineCircuitBreaker en SQLite
// Estos campos NO están en el schema PUT del endpoint /api/settings/compression
// Se escriben directo en la tabla key_value del namespace "compression"
//
// Ejecutar: node scripts/setup-safety-gates.mjs
// Requiere: omniroute restart después de ejecutar

import { readFileSync } from "fs";
import { join, dirname } from "path";
import { fileURLToPath } from "url";
import { createRequire } from "module";

const __dirname = dirname(fileURLToPath(import.meta.url));
const require = createRequire(import.meta.url);

// better-sqlite3 viene incluido con OmniRoute
const Database = require(
  "C:\\Users\\anton\\AppData\\Roaming\\npm\\node_modules\\omniroute\\node_modules\\better-sqlite3"
);

const DB_PATH = join(
  process.env.USERPROFILE || process.env.HOME,
  ".omniroute",
  "storage.sqlite"
);

const NAMESPACE = "compression";

// Leer config desde safety-gates.json
const configPath = join(__dirname, "..", "config", "safety-gates.json");
const config = JSON.parse(readFileSync(configPath, "utf-8"));

console.log("=== Safety Gates Setup ===");
console.log("DB:", DB_PATH);

const db = new Database(DB_PATH, { readonly: false });
db.pragma("journal_mode = WAL");

const insert = db.prepare(
  "INSERT OR REPLACE INTO key_value (namespace, key, value) VALUES (?, ?, ?)"
);

const entries = [
  { key: "fidelityGate", value: config.fidelityGate },
  { key: "riskGate", value: config.riskGate },
  { key: "pipelineCircuitBreaker", value: config.pipelineCircuitBreaker },
];

const tx = db.transaction(() => {
  for (const { key, value } of entries) {
    insert.run(NAMESPACE, key, JSON.stringify(value));
    console.log(`OK: ${key} = ${JSON.stringify(value)}`);
  }
});

tx();
db.close();

console.log("");
console.log("=== Safety gates escritos exitosamente ===");
console.log("IMPORTANTE: Ejecutar 'omniroute restart' para que los cambios tomen efecto.");
