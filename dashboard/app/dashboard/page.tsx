"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supabase, type Service } from "@/lib/supabase";

const LANG_BADGE: Record<string, string> = {
  node:   "badge-green",
  python: "badge-orange",
  go:     "badge-green",
  other:  "badge-gray",
};

type ServiceWithStatus = Service & { connectionDot: string; connectionLabel: string; connectionCls: string };

function enrichWithStatus(services: Service[]): ServiceWithStatus[] {
  return services.map((svc) => {
    // We don't have last_used_at on services directly — will be shown per key on service page
    return { ...svc, connectionDot: "", connectionLabel: "", connectionCls: "" };
  });
}

export default function DashboardPage() {
  const router = useRouter();
  const [services, setServices] = useState<ServiceWithStatus[]>([]);
  const [keyStatus, setKeyStatus] = useState<Record<string, { dot: string; label: string; cls: string }>>({});
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadServices();
    const t = setInterval(loadServices, 30_000);
    return () => clearInterval(t);
  }, []);

  async function loadServices() {
    const { data: svcs } = await supabase
      .from("services")
      .select("*")
      .order("created_at", { ascending: false });

    if (!svcs) { setLoading(false); return; }

    // Fetch most-recent active key per service to determine connection status
    const { data: keys } = await supabase
      .from("api_keys")
      .select("service_id, last_used_at")
      .is("revoked_at", null)
      .order("last_used_at", { ascending: false });

    // Map service_id → most recent last_used_at
    const latestByService: Record<string, string | null> = {};
    for (const key of keys ?? []) {
      if (!latestByService[key.service_id]) {
        latestByService[key.service_id] = key.last_used_at;
      }
    }

    const statusMap: Record<string, { dot: string; label: string; cls: string }> = {};
    for (const svc of svcs) {
      statusMap[svc.id] = calcStatus(latestByService[svc.id] ?? null);
    }

    setKeyStatus(statusMap);
    setServices(enrichWithStatus(svcs));
    setLoading(false);
  }

  async function deleteService(id: string) {
    if (!confirm("Delete this service and all its API keys?")) return;
    await supabase.from("services").delete().eq("id", id);
    setServices((prev) => prev.filter((s) => s.id !== id));
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Services</h1>
          <p style={{ color: "var(--muted)", fontSize: "13px", marginTop: "4px" }}>
            {services.length} registered service{services.length !== 1 ? "s" : ""}
          </p>
        </div>
        <button className="btn btn-primary" onClick={() => router.push("/dashboard/new")}>
          + New Service
        </button>
      </div>

      {loading ? (
        <p style={{ color: "var(--muted)" }}>Loading…</p>
      ) : services.length === 0 ? (
        <div className="empty card">
          <h3>No services yet</h3>
          <p style={{ marginBottom: "20px" }}>
            Register your first backend service to get an API key.
          </p>
          <button className="btn btn-primary" onClick={() => router.push("/dashboard/new")}>
            + Register Service
          </button>
        </div>
      ) : (
        <div className="card" style={{ padding: 0 }}>
          <table className="table">
            <thead>
              <tr>
                <th>Name</th>
                <th>Language</th>
                <th>Status</th>
                <th>Last seen</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {services.map((svc) => {
                const st = keyStatus[svc.id] ?? calcStatus(null);
                return (
                  <tr key={svc.id} style={{ cursor: "pointer" }}>
                    <td onClick={() => router.push(`/dashboard/service?id=${svc.id}`)} style={{ fontWeight: 600 }}>
                      {svc.name}
                      {svc.description && (
                        <span style={{ fontWeight: 400, color: "var(--muted)", marginLeft: "8px", fontSize: "12px" }}>
                          {svc.description}
                        </span>
                      )}
                    </td>
                    <td>
                      {svc.language ? (
                        <span className={`badge ${LANG_BADGE[svc.language] ?? "badge-gray"}`}>{svc.language}</span>
                      ) : (
                        <span style={{ color: "var(--muted)" }}>—</span>
                      )}
                    </td>
                    <td>
                      <span className={`badge ${st.cls}`}>{st.dot} {st.label}</span>
                    </td>
                    <td style={{ color: "var(--muted)", fontSize: "12px" }}>
                      {getLastSeen(svc.id, keyStatus)}
                    </td>
                    <td style={{ textAlign: "right" }}>
                      <button className="btn btn-ghost" style={{ fontSize: "12px", marginRight: "6px" }}
                        onClick={() => router.push(`/dashboard/service?id=${svc.id}`)}>
                        Manage keys
                      </button>
                      <button className="btn btn-danger" style={{ fontSize: "12px" }}
                        onClick={() => deleteService(svc.id)}>
                        Delete
                      </button>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

function calcStatus(lastUsedAt: string | null) {
  if (!lastUsedAt) return { dot: "⚫", label: "Not connected", cls: "badge-gray" };
  const mins = (Date.now() - new Date(lastUsedAt).getTime()) / 60_000;
  if (mins < 5)  return { dot: "🟢", label: "Live", cls: "badge-green" };
  if (mins < 60) return { dot: "🟡", label: "Recent", cls: "badge-orange" };
  return { dot: "🔴", label: "Inactive", cls: "badge-red" };
}

function getLastSeen(serviceId: string, statusMap: Record<string, { dot: string; label: string; cls: string }>) {
  // We stored status but not the raw date — revisit if needed; show label for now
  const st = statusMap[serviceId];
  if (!st || st.label === "Not connected") return "Never";
  return st.label === "Live" ? "Just now" : st.label;
}
