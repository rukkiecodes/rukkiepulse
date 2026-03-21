"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";

export default function NewServicePage() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [language, setLanguage] = useState("node");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    const { data: { user } } = await supabase.auth.getUser();
    if (!user) { router.replace("/login"); return; }

    const { data, error } = await supabase
      .from("services")
      .insert({ name, description: description || null, language, owner_id: user.id })
      .select()
      .single();

    if (error) {
      setError(error.message);
      setLoading(false);
    } else {
      router.push(`/dashboard/services/${data.id}`);
    }
  }

  return (
    <div className="page" style={{ maxWidth: "540px" }}>
      <div className="page-header">
        <h1 className="page-title">Register Service</h1>
      </div>

      <div className="card">
        <form onSubmit={handleCreate}>
          <div className="form-field">
            <label htmlFor="name">Service Name *</label>
            <input
              id="name"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="auth-service"
              required
              autoFocus
            />
          </div>

          <div className="form-field">
            <label htmlFor="desc">Description</label>
            <input
              id="desc"
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder="Handles user authentication"
            />
          </div>

          <div className="form-field">
            <label htmlFor="lang">Language / Runtime</label>
            <select
              id="lang"
              value={language}
              onChange={(e) => setLanguage(e.target.value)}
            >
              <option value="node">Node.js</option>
              <option value="python">Python</option>
              <option value="go">Go</option>
              <option value="other">Other</option>
            </select>
          </div>

          {error && <p className="error-msg">{error}</p>}

          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={loading}>
              {loading ? "Creating…" : "Create & get API key →"}
            </button>
            <button
              type="button"
              className="btn btn-ghost"
              onClick={() => router.back()}
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
