const { Client } = require('pg');
const fs = require('fs');
const path = require('path');

const env = fs.readFileSync(path.join(__dirname, '..', '.env'), 'utf8');
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

async function run() {
  await client.connect();

  // Check if user exists
  const check = await client.query(
    `SELECT id, email FROM auth.users WHERE email = $1`,
    ['rukkiecodes@gmail.com']
  );

  if (check.rows.length > 0) {
    console.log('User already exists:', check.rows[0].email);
    // Update password in case it changed
    await client.query(`
      UPDATE auth.users
      SET encrypted_password = crypt($1, gen_salt('bf')),
          updated_at = now()
      WHERE email = $2
    `, ['Rukkie@codemonster100;', 'rukkiecodes@gmail.com']);
    console.log('Password updated.');
  } else {
    await client.query(`
      INSERT INTO auth.users (
        instance_id, id, aud, role,
        email, encrypted_password,
        email_confirmed_at, created_at, updated_at,
        confirmation_token, email_change, email_change_token_new, recovery_token
      ) VALUES (
        '00000000-0000-0000-0000-000000000000',
        gen_random_uuid(),
        'authenticated', 'authenticated',
        'rukkiecodes@gmail.com',
        crypt('Rukkie@codemonster100;', gen_salt('bf')),
        now(), now(), now(),
        '', '', '', ''
      )
    `);
    console.log('User created: rukkiecodes@gmail.com');
  }

  await client.end();
}

run().catch(err => { console.error(err.message); process.exit(1); });
