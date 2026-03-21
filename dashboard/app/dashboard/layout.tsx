"use client";
import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { supabase } from "@/lib/supabase";

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const router = useRouter();
  const [email, setEmail] = useState("");
  const [ready, setReady] = useState(false);

  useEffect(() => {
    supabase.auth.getSession().then(({ data }) => {
      if (!data.session) {
        router.replace("/login");
      } else {
        setEmail(data.session.user.email ?? "");
        setReady(true);
      }
    });
  }, [router]);

  async function handleLogout() {
    await supabase.auth.signOut();
    router.replace("/login");
  }

  if (!ready) return null;

  return (
    <>
      <nav className="nav">
        <a href="/dashboard" className="nav-logo">
          rukkie<span>pulse</span>
        </a>
        <div className="nav-right">
          <span style={{ color: "var(--muted)", fontSize: "12px" }}>{email}</span>
          <button className="btn btn-ghost" onClick={handleLogout} style={{ fontSize: "12px" }}>
            Sign out
          </button>
        </div>
      </nav>
      {children}
    </>
  );
}
