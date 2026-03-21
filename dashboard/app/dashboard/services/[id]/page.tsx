"use client";
import { useEffect, useState, use } from "react";
import { useRouter } from "next/navigation";
import { supabase, type Service, type ApiKey } from "@/lib/supabase";
import { generateApiKey } from "@/lib/apikeys";

export default function ServicePage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = use(params);
  const router = useRouter();

  const [service, setService] = useState<Service | null>(null);
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [creating, setCreating] = useState(false);
  const [newKeyLabel, setNewKeyLabel] = useState("default");
  const [showNewKeyForm, setShowNewKeyForm] = useState(false);
  const [revealedKey, setRevealedKey] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    loadData();
  }, [id]);

  async function loadData() {
    const [{ data: svc }, { data: apiKeys }] = await Promise.all([
      supabase.from("services").select("*").eq("id", id).single(),
      supabase.from("api_keys").select("*").eq("service_id", id).order("created_at", { ascending: false }),
    ]);
    if (!svc) { router.replace("/dashboard"); return; }
    setService(svc);
    setKeys(apiKeys ?? []);
    setLoading(false);
  }

  async function createKey() {
    setCreating(true);
    const { fullKey, prefix, hash } = await generateApiKey();

    const { error } = await supabase.from("api_keys").insert({
      service_id: id,
      label: newKeyLabel || "default",
      key_prefix: prefix,
      key_hash: hash,
    });

    if (error) {
      alert(error.message);
      setCreating(false);
      return;
    }

    setRevealedKey(fullKey);
    setShowNewKeyForm(false);
    setNewKeyLabel("default");
    setCreating(false);
    loadData();
  }

  async function revokeKey(keyId: string) {
    if (!confirm("Revoke this API key? Requests using it will fail immediately.")) return;
    await supabase
      .from("api_keys")
      .update({ revoked_at: new Date().toISOString() })
      .eq("id", keyId);
    loadData();
  }

  async function deleteKey(keyId: string) {
    if (!confirm("Permanently delete this API key?")) return;
    await supabase.from("api_keys").delete().eq("id", keyId);
    setKeys((prev) => prev.filter((k) => k.id !== keyId));
  }

  function copyKey() {
    if (!revealedKey) return;
    navigator.clipboard.writeText(revealedKey);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  }

  if (loading) return null;
  if (!service) return null;

  const activeKeys = keys.filter((k) => !k.revoked_at);
  const revokedKeys = keys.filter((k) => k.revoked_at);

  const snippet = {
    node: `import { initRukkie } from 'rukkie-agent'

initRukkie({
  serviceName: '${service.name}',
  apiKey: 'YOUR_API_KEY',
})`,
    python: `from rukkie_agent import init_rukkie

init_rukkie(
    service_name="${service.name}",
    api_key="YOUR_API_KEY",
)`,
    go: `// Go agent coming soon — use the REST API directly`,
    other: `// POST /api/v1/heartbeat
// Authorization: Bearer YOUR_API_KEY`,
  };

  return (
    <div className="page">
      {/* Revealed key modal */}
      {revealedKey && (
        <div className="modal-overlay">
          <div className="modal">
            <h2 style={{ color: "var(--green)" }}>API Key Created</h2>
            <p style={{ color: "var(--muted)", fontSize: "13px", marginBottom: "16px" }}>
              Copy this key now — it will never be shown again.
            </p>
            <div className="key-display">
              <span style={{ flex: 1, wordBreak: "break-all", color: "var(--green)" }}>
                {revealedKey}
              </span>
              <button className="btn btn-secondary" style={{ flexShrink: 0 }} onClick={copyKey}>
                {copied ? "Copied!" : "Copy"}
              </button>
            </div>
            <button
              className="btn btn-primary"
              style={{ width: "100%", justifyContent: "center", marginTop: "20px" }}
              onClick={() => setRevealedKey(null)}
            >
              I've saved it →
            </button>
          </div>
        </div>
      )}

      {/* Header */}
      <div className="page-header">
        <div>
          <p style={{ color: "var(--muted)", fontSize: "12px", marginBottom: "4px" }}>
            <a href="/dashboard">← Services</a>
          </p>
          <h1 className="page-title">{service.name}</h1>
          {service.description && (
            <p style={{ color: "var(--muted)", fontSize: "13px", marginTop: "4px" }}>
              {service.description}
            </p>
          )}
        </div>
        <button
          className="btn btn-primary"
          onClick={() => setShowNewKeyForm(true)}
        >
          + Generate API Key
        </button>
      </div>

      {/* New key form */}
      {showNewKeyForm && (
        <div className="card" style={{ marginBottom: "24px" }}>
          <h3 style={{ marginBottom: "14px", fontSize: "14px" }}>New API Key</h3>
          <div style={{ display: "flex", gap: "10px", alignItems: "flex-end" }}>
            <div style={{ flex: 1 }}>
              <label>Label</label>
              <input
                value={newKeyLabel}
                onChange={(e) => setNewKeyLabel(e.target.value)}
                placeholder="production"
                autoFocus
              />
            </div>
            <button className="btn btn-primary" onClick={createKey} disabled={creating}>
              {creating ? "Generating…" : "Generate"}
            </button>
            <button className="btn btn-ghost" onClick={() => setShowNewKeyForm(false)}>
              Cancel
            </button>
          </div>
        </div>
      )}

      {/* Active keys */}
      <div className="card" style={{ marginBottom: "24px", padding: 0 }}>
        <div style={{ padding: "16px 20px", borderBottom: "1px solid var(--border)", display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <h2 style={{ fontSize: "14px", fontWeight: 600 }}>
            Active Keys <span style={{ color: "var(--muted)", fontWeight: 400 }}>({activeKeys.length})</span>
          </h2>
        </div>
        {activeKeys.length === 0 ? (
          <div className="empty">
            <p>No active keys. Generate one above.</p>
          </div>
        ) : (
          <table className="table">
            <thead>
              <tr>
                <th>Label</th>
                <th>Key (prefix)</th>
                <th>Created</th>
                <th>Last used</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {activeKeys.map((key) => (
                <tr key={key.id}>
                  <td style={{ fontWeight: 600 }}>{key.label}</td>
                  <td style={{ fontFamily: "monospace", color: "var(--green)", fontSize: "13px" }}>
                    {key.key_prefix}…
                  </td>
                  <td style={{ color: "var(--muted)", fontSize: "12px" }}>
                    {new Date(key.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ color: "var(--muted)", fontSize: "12px" }}>
                    {key.last_used_at ? new Date(key.last_used_at).toLocaleDateString() : "Never"}
                  </td>
                  <td style={{ textAlign: "right" }}>
                    <button
                      className="btn btn-ghost"
                      style={{ fontSize: "12px", marginRight: "6px" }}
                      onClick={() => revokeKey(key.id)}
                    >
                      Revoke
                    </button>
                    <button
                      className="btn btn-danger"
                      style={{ fontSize: "12px" }}
                      onClick={() => deleteKey(key.id)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>

      {/* Code snippet */}
      <div className="card" style={{ marginBottom: "24px" }}>
        <h2 style={{ fontSize: "14px", fontWeight: 600, marginBottom: "12px" }}>Integration Snippet</h2>
        <pre className="snippet">
          {snippet[service.language ?? "other"] ?? snippet.other}
        </pre>
        <p style={{ color: "var(--muted)", fontSize: "12px", marginTop: "10px" }}>
          Replace <code style={{ color: "var(--orange)" }}>YOUR_API_KEY</code> with an active key above.
        </p>
      </div>

      {/* Revoked keys */}
      {revokedKeys.length > 0 && (
        <div className="card" style={{ padding: 0 }}>
          <div style={{ padding: "16px 20px", borderBottom: "1px solid var(--border)" }}>
            <h2 style={{ fontSize: "14px", fontWeight: 600, color: "var(--muted)" }}>
              Revoked Keys ({revokedKeys.length})
            </h2>
          </div>
          <table className="table">
            <thead>
              <tr>
                <th>Label</th>
                <th>Key (prefix)</th>
                <th>Revoked</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {revokedKeys.map((key) => (
                <tr key={key.id}>
                  <td style={{ color: "var(--muted)" }}>{key.label}</td>
                  <td style={{ fontFamily: "monospace", color: "var(--muted)", fontSize: "13px" }}>
                    {key.key_prefix}…
                  </td>
                  <td style={{ color: "var(--muted)", fontSize: "12px" }}>
                    {key.revoked_at ? new Date(key.revoked_at).toLocaleDateString() : "—"}
                  </td>
                  <td style={{ textAlign: "right" }}>
                    <button
                      className="btn btn-danger"
                      style={{ fontSize: "12px" }}
                      onClick={() => deleteKey(key.id)}
                    >
                      Delete
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
