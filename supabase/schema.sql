-- ============================================================
-- RukkiePulse — Supabase Schema
-- Run this in: Supabase Dashboard → SQL Editor → New query
-- ============================================================

-- ── 1. Services ─────────────────────────────────────────────
create table if not exists public.services (
  id          uuid primary key default gen_random_uuid(),
  owner_id    uuid references auth.users(id) on delete cascade not null,
  name        text not null,
  description text,
  language    text check (language in ('node', 'python', 'go', 'other')),
  created_at  timestamptz default now() not null
);

alter table public.services enable row level security;

create policy "owner can manage services"
  on public.services for all
  using  (owner_id = auth.uid())
  with check (owner_id = auth.uid());

-- ── 2. API Keys ─────────────────────────────────────────────
create table if not exists public.api_keys (
  id           uuid primary key default gen_random_uuid(),
  service_id   uuid references public.services(id) on delete cascade not null,
  label        text not null default 'default',
  key_prefix   text not null,   -- first 12 chars shown in UI  (rk_live_xxxx)
  key_hash     text not null unique, -- sha-256 of the full key, for validation
  created_at   timestamptz default now() not null,
  last_used_at timestamptz,
  revoked_at   timestamptz
);

alter table public.api_keys enable row level security;

create policy "owner can manage api_keys via service"
  on public.api_keys for all
  using (
    service_id in (
      select id from public.services where owner_id = auth.uid()
    )
  )
  with check (
    service_id in (
      select id from public.services where owner_id = auth.uid()
    )
  );

-- ── 3. Seed admin user ──────────────────────────────────────
-- Creates the rukkiecodes account if it doesn't already exist.
-- Password: Rukkie@codemonster100;
do $$
declare
  uid uuid := gen_random_uuid();
begin
  if not exists (
    select 1 from auth.users where email = 'rukkiecodes@gmail.com'
  ) then
    insert into auth.users (
      instance_id, id, aud, role,
      email, encrypted_password,
      email_confirmed_at, created_at, updated_at,
      confirmation_token, email_change, email_change_token_new, recovery_token
    ) values (
      '00000000-0000-0000-0000-000000000000',
      uid,
      'authenticated',
      'authenticated',
      'rukkiecodes@gmail.com',
      crypt('Rukkie@codemonster100;', gen_salt('bf')),
      now(), now(), now(),
      '', '', '', ''
    );
  end if;
end $$;
