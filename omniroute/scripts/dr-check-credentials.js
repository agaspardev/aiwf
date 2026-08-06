// dr-check-credentials.js - valida que el .env restaurado descifra las credenciales de storage.sqlite
// Uso: node dr-check-credentials.js <storage.sqlite> <.env>
// Exit 0 = todas las credenciales encriptadas descifran OK; exit 1 = fallo
const path = require("path");
const crypto = require("crypto");
const fs = require("fs");

const [dbPath, envPath] = process.argv.slice(2);
if (!dbPath || !envPath) {
  console.error("Uso: node dr-check-credentials.js <storage.sqlite> <.env>");
  process.exit(2);
}

function loadEnv(p) {
  const env = {};
  for (const line of fs.readFileSync(p, "utf8").split(/\r?\n/)) {
    const m = line.match(/^([A-Za-z_][A-Za-z0-9_]*)=(.*)$/);
    if (m) env[m[1]] = m[2].trim();
  }
  return env;
}

let Database;
try {
  Database = require("C:/Users/anton/AppData/Roaming/npm/node_modules/omniroute/node_modules/better-sqlite3");
} catch {
  Database = require("better-sqlite3");
}

const env = loadEnv(envPath);
if (!env.STORAGE_ENCRYPTION_KEY) {
  console.log("NO_ENCRYPTION_KEY");
  process.exit(0);
}
const aesKey = crypto.scryptSync(env.STORAGE_ENCRYPTION_KEY, "omniroute-field-encryption-v1", 32);

const db = new Database(dbPath, { readonly: true });
const cols = ["api_key", "access_token", "refresh_token", "id_token"];
let rows = [];
try {
  rows = db.prepare("SELECT name, " + cols.join(",") + " FROM provider_connections").all();
} catch (e) {
  console.error("DB_ERROR", e.message);
  process.exit(1);
}

function decrypt(v) {
  const parts = v.slice("enc:v1:".length).split(":");
  if (parts.length !== 3) return { kind: "malformed" };
  try {
    const d = crypto.createDecipheriv("aes-256-gcm", aesKey, Buffer.from(parts[0], "hex"), { authTagLength: 16 });
    d.setAuthTag(Buffer.from(parts[2], "hex"));
    let out = d.update(parts[1], "hex", "utf8");
    out += d.final("utf8");
    return { kind: "ok", val: out };
  } catch (e) {
    return { kind: "fail", err: e.message };
  }
}

let enc = 0, ok = 0, fail = 0;
for (const r of rows) {
  for (const c of cols) {
    const v = r[c];
    if (!v || !String(v).startsWith("enc:v1:")) continue;
    enc++;
    const res = decrypt(String(v));
    if (res.kind === "ok") ok++;
    else { fail++; console.log("FAIL", r.name, c, res.err || "malformed"); }
  }
}

// provider_specific_data (objetos JSON con credenciales dentro)
try {
  const psd = db.prepare("SELECT name, provider_specific_data FROM provider_connections WHERE provider_specific_data IS NOT NULL AND provider_specific_data != ''").all();
  const walk = (node, who) => {
    if (!node || typeof node !== "object") return;
    for (const [k, v] of Object.entries(node)) {
      if (typeof v === "string" && v.startsWith("enc:v1:")) {
        enc++;
        const res = decrypt(v);
        if (res.kind === "ok") ok++;
        else { fail++; console.log("FAIL", who, k, res.err || "malformed"); }
      } else if (typeof v === "object") walk(v, who);
    }
  };
  for (const r of psd) walk(JSON.parse(r.provider_specific_data), r.name);
} catch (e) {}

db.close();
console.log(`CREDENTIALS enc=${enc} OK=${ok} FAIL=${fail}`);
process.exit(fail > 0 ? 1 : 0);
