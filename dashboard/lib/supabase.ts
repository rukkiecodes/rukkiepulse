import { createClient } from "@supabase/supabase-js";

const url = process.env.NEXT_PUBLIC_SUPABASE_URL!;
const key = process.env.NEXT_PUBLIC_SUPABASE_ANON_KEY!;

export const supabase = createClient(url, key);

export type Service = {
  id: string;
  owner_id: string;
  name: string;
  description: string | null;
  language: "node" | "python" | "go" | "other" | null;
  created_at: string;
};

export type ApiKey = {
  id: string;
  service_id: string;
  label: string;
  key_prefix: string;
  key_hash: string;
  created_at: string;
  last_used_at: string | null;
  revoked_at: string | null;
};
