import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const corsHeaders = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers": "authorization, content-type",
};

Deno.serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response("ok", { headers: corsHeaders });
  }

  const authHeader = req.headers.get("authorization") ?? "";
  const apiKey = authHeader.replace(/^Bearer\s+/i, "").trim();

  if (!apiKey) {
    return new Response(JSON.stringify({ error: "Missing API key" }), {
      status: 401,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }

  // Hash the incoming key with SHA-256
  const encoder = new TextEncoder();
  const keyData = encoder.encode(apiKey);
  const hashBuf = await crypto.subtle.digest("SHA-256", keyData);
  const hashArr = Array.from(new Uint8Array(hashBuf));
  const keyHash = hashArr.map((b) => b.toString(16).padStart(2, "0")).join("");

  const supabase = createClient(
    Deno.env.get("SUPABASE_URL")!,
    Deno.env.get("SUPABASE_SERVICE_ROLE_KEY")!
  );

  // Find the key and mark last_used_at
  const { data: keyRow, error } = await supabase
    .from("api_keys")
    .select("id, service_id, revoked_at, services(name, language)")
    .eq("key_hash", keyHash)
    .single();

  if (error || !keyRow) {
    return new Response(JSON.stringify({ error: "Invalid API key" }), {
      status: 403,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }

  if (keyRow.revoked_at) {
    return new Response(JSON.stringify({ error: "API key has been revoked" }), {
      status: 403,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    });
  }

  // Update last_used_at
  await supabase
    .from("api_keys")
    .update({ last_used_at: new Date().toISOString() })
    .eq("id", keyRow.id);

  const service = Array.isArray(keyRow.services)
    ? keyRow.services[0]
    : keyRow.services;

  return new Response(
    JSON.stringify({
      ok: true,
      service: service?.name ?? "unknown",
      language: service?.language ?? "other",
    }),
    {
      status: 200,
      headers: { ...corsHeaders, "Content-Type": "application/json" },
    }
  );
});
