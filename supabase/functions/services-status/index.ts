import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, content-type, x-rukkie-cli",
};

// Simple pre-shared secret so only the CLI can call this
const CLI_SECRET = "rukkie-cli-v1-xqmjdjjwprnqogokoejz";

Deno.serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  const cliHeader = req.headers.get("x-rukkie-cli") ?? "";
  if (cliHeader !== CLI_SECRET) {
    return new Response(JSON.stringify({ error: "Unauthorized" }), {
      status: 401,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }

  const supabase = createClient(
    Deno.env.get("SUPABASE_URL")!,
    Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
  );

  // Fetch all services with their most-recent api_key last_used_at
  const { data: services, error } = await supabase
    .from("services")
    .select(
      "id, name, language, description, created_at, api_keys(id, label, key_prefix, last_used_at, revoked_at)"
    )
    .order("created_at", { ascending: true });

  if (error) {
    return new Response(JSON.stringify({ error: error.message }), {
      status: 500,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }

  // For each service, find the most recently active (non-revoked) api key
  const result = (services ?? []).map((svc: any) => {
    const keys: any[] = Array.isArray(svc.api_keys) ? svc.api_keys : [];
    const activeKeys = keys.filter((k) => !k.revoked_at);
    // pick key with the latest last_used_at
    let lastUsedAt: string | null = null;
    for (const k of activeKeys) {
      if (!lastUsedAt || (k.last_used_at && k.last_used_at > lastUsedAt)) {
        lastUsedAt = k.last_used_at;
      }
    }
    return {
      id: svc.id,
      name: svc.name,
      language: svc.language ?? "other",
      description: svc.description ?? "",
      activeKeys: activeKeys.length,
      lastUsedAt,
    };
  });

  return new Response(JSON.stringify({ services: result }), {
    status: 200,
    headers: { ...corsHeaders, "Content-Type": "application/json" },
  });
});
