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

export default function DashboardPage() {
  const router = useRouter();
  const [services, setServices] = useState<Service[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadServices();
  }, []);

  async function loadServices() {
    const { data } = await supabase
      .from("services")
      .select("*")
      .order("created_at", { ascending: false });
    setServices(data ?? []);
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
        <button
          className="btn btn-primary"
          onClick={() => router.push("/dashboard/new")}
        >
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
                <th>Created</th>
                <th></th>
              </tr>
            </thead>
            <tbody>
              {services.map((svc) => (
                <tr key={svc.id} style={{ cursor: "pointer" }}>
                  <td
                    onClick={() => router.push(`/dashboard/services/${svc.id}`)}
                    style={{ fontWeight: 600 }}
                  >
                    {svc.name}
                    {svc.description && (
                      <span style={{ fontWeight: 400, color: "var(--muted)", marginLeft: "8px", fontSize: "12px" }}>
                        {svc.description}
                      </span>
                    )}
                  </td>
                  <td>
                    {svc.language ? (
                      <span className={`badge ${LANG_BADGE[svc.language] ?? "badge-gray"}`}>
                        {svc.language}
                      </span>
                    ) : (
                      <span style={{ color: "var(--muted)" }}>—</span>
                    )}
                  </td>
                  <td style={{ color: "var(--muted)", fontSize: "12px" }}>
                    {new Date(svc.created_at).toLocaleDateString()}
                  </td>
                  <td style={{ textAlign: "right" }}>
                    <button
                      className="btn btn-ghost"
                      style={{ fontSize: "12px", marginRight: "6px" }}
                      onClick={() => router.push(`/dashboard/services/${svc.id}`)}
                    >
                      Manage keys
                    </button>
                    <button
                      className="btn btn-danger"
                      style={{ fontSize: "12px" }}
                      onClick={() => deleteService(svc.id)}
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
