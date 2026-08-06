const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const src = process.argv[2];
const dest = process.argv[3];

let Database;
try {
  Database = require("C:/Users/anton/AppData/Roaming/npm/node_modules/omniroute/node_modules/better-sqlite3");
} catch {
  Database = require("better-sqlite3");
}

if (fs.existsSync(dest)) fs.unlinkSync(dest);

if (typeof Database.prototype.backup === "function") {
  const db = new Database(src, { readonly: true });
  db.backup(dest)
    .then(() => { db.close(); console.log("BACKUP_OK", dest); })
    .catch(e => { db.close(); console.error("BACKUP_FAIL", e.message); process.exit(1); });
} else {
  // fallback: VACUUM INTO (consistent snapshot, online)
  const db = new Database(src, { readonly: true });
  db.pragma(`journal_mode = wal`);
  db.exec(`VACUUM INTO '${dest.replace(/'/g, "''")}'`);
  db.close();
  console.log("BACKUP_OK_VACUUM", dest);
}
