const { Client } = require('pg');
const fs = require('fs');
const path = require('path');

// Load .env
const envPath = path.join(__dirname, '..', '.env');
const env = fs.readFileSync(envPath, 'utf8');
const vars = {};
env.split('\n').forEach(line => {
  line = line.trim();
  if (!line || line.startsWith('#')) return;
  const idx = line.indexOf('=');
  if (idx > -1) vars[line.slice(0, idx).trim()] = line.slice(idx + 1).trim();
});

const client = new Client({
  host: vars.SUPERBASE_POOL_HOST,
  port: parseInt(vars.SUPERBASE_POOL_PORT),
  database: vars.SUPERBASE_POOL_DATABASE,
  user: vars.SUPERBASE_POOL_USER,
  password: vars.SUPERBASE_DB_PASSWORD,
  ssl: { rejectUnauthorized: false },
});

const sql = fs.readFileSync(path.join(__dirname, '..', 'supabase', 'schema.sql'), 'utf8');

async function run() {
  console.log('Connecting to Supabase...');
  await client.connect();
  console.log('Connected. Running migration...');
  try {
    await client.query(sql);
    console.log('Migration complete.');
  } catch (err) {
    console.error('Migration error:', err.message);
    process.exit(1);
  } finally {
    await client.end();
  }
}

run();
