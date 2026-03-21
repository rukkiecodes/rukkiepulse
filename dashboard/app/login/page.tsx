"use client";
import { useState } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  async function handleLogin(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    const { error } = await supabase.auth.signInWithPassword({ email, password });
    if (error) {
      setError(error.message);
      setLoading(false);
    } else {
      router.replace("/dashboard");
    }
  }

  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", padding: "24px" }}>
      <div style={{ width: "100%", maxWidth: "400px" }}>
        {/* Logo */}
        <div style={{ textAlign: "center", marginBottom: "36px" }}>
          <h1 style={{ fontSize: "28px", fontWeight: 700, letterSpacing: "-0.5px" }}>
            <span style={{ color: "var(--green)" }}>rukkie</span>
            <span style={{ color: "var(--orange)" }}>pulse</span>
          </h1>
          <p style={{ color: "var(--muted)", marginTop: "6px", fontSize: "13px" }}>
            Observability Dashboard
          </p>
        </div>

        <div className="card">
          <h2 style={{ fontSize: "16px", fontWeight: 600, marginBottom: "20px" }}>Sign in</h2>

          <form onSubmit={handleLogin}>
            <div className="form-field">
              <label htmlFor="email">Email</label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
                autoFocus
              />
            </div>

            <div className="form-field">
              <label htmlFor="password">Password</label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>

            {error && <p className="error-msg">{error}</p>}

            <button
              type="submit"
              className="btn btn-primary"
              disabled={loading}
              style={{ width: "100%", justifyContent: "center", marginTop: "20px" }}
            >
              {loading ? "Signing in…" : "Sign in →"}
            </button>
          </form>
        </div>

        <p style={{ textAlign: "center", color: "var(--muted)", marginTop: "20px", fontSize: "12px" }}>
          <a href="https://rukkiepulse.netlify.app" style={{ color: "var(--muted)" }}>
            ← Back to docs
          </a>
        </p>
      </div>
    </div>
  );
}
