"use client";

import { useState } from "react";
import { createClient } from "@/lib/supabase/client";
import { useRouter } from "next/navigation";
import { siteConfig } from "@/lib/constants";
import Link from "next/link";

export default function SignupPage() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [orgName, setOrgName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [success, setSuccess] = useState(false);
  const router = useRouter();

  async function handleSignup(e: React.FormEvent) {
    e.preventDefault();
    setLoading(true);
    setError("");

    if (password.length < 8) {
      setError("Password must be at least 8 characters");
      setLoading(false);
      return;
    }

    const supabase = createClient();
    const { error } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: { org_name: orgName },
        emailRedirectTo: `${window.location.origin}/auth/callback`,
      },
    });

    if (error) {
      setError(error.message);
      setLoading(false);
      return;
    }

    setSuccess(true);
  }

  async function handleOAuth(provider: "github" | "google") {
    const supabase = createClient();
    await supabase.auth.signInWithOAuth({
      provider,
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });
  }

  if (success) {
    return (
      <div
        className="flex min-h-screen items-center justify-center px-4"
        style={{ background: "var(--bg-primary)" }}
      >
        <div
          className="bg-research-grid fixed inset-0 pointer-events-none"
          style={{ opacity: 0.4 }}
        />
        <div className="relative w-full max-w-sm text-center">
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: "50%",
              border: "1px solid var(--border-light)",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              margin: "0 auto 16px",
              color: "var(--accent-teal)",
            }}
          >
            <svg
              width="24"
              height="24"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={1.5}
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h1
            className="font-editorial"
            style={{ fontSize: 24, color: "var(--text-primary)" }}
          >
            Check your email
          </h1>
          <p
            style={{
              marginTop: 12,
              fontSize: 14,
              color: "var(--text-muted)",
              lineHeight: 1.6,
            }}
          >
            We&apos;ve sent a verification link to{" "}
            <strong style={{ color: "var(--text-secondary)" }}>{email}</strong>.
            Click the link to activate your account.
          </p>
          <Link
            href="/login"
            style={{
              display: "inline-block",
              marginTop: 24,
              fontSize: 13,
              color: "var(--text-muted)",
              textDecoration: "underline",
              textUnderlineOffset: 3,
            }}
          >
            Back to login
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div
      className="flex min-h-screen items-center justify-center px-4"
      style={{ background: "var(--bg-primary)" }}
    >
      <div
        className="bg-research-grid fixed inset-0 pointer-events-none"
        style={{ opacity: 0.4 }}
      />

      <div className="relative w-full max-w-sm">
        {/* Logo */}
        <div style={{ marginBottom: 40, textAlign: "center" }}>
          <Link
            href="/"
            style={{
              display: "inline-flex",
              alignItems: "center",
              gap: 10,
              textDecoration: "none",
              marginBottom: 32,
            }}
          >
            <svg width="28" height="28" viewBox="0 0 28 28" fill="none">
              <rect width="28" height="28" rx="6" fill="#1A1714" />
              <path d="M7 10l7-4 7 4-7 4-7-4z" fill="#A1A1AA" />
              <path d="M7 14l7 4 7-4" stroke="#fff" strokeWidth="1.5" />
              <path d="M7 18l7 4 7-4" stroke="#71717A" strokeWidth="1.5" />
            </svg>
            <span
              className="font-editorial"
              style={{ fontSize: 20, color: "var(--text-primary)" }}
            >
              {siteConfig.name}
            </span>
          </Link>

          <h1
            className="font-editorial"
            style={{
              fontSize: 28,
              lineHeight: 1.2,
              color: "var(--text-primary)",
              marginTop: 8,
            }}
          >
            Create your account
          </h1>
          <p
            style={{
              marginTop: 8,
              fontSize: 14,
              color: "var(--text-muted)",
            }}
          >
            Start catching API breaking changes today
          </p>
        </div>

        {/* OAuth */}
        <div style={{ marginBottom: 24 }}>
          <button
            onClick={() => handleOAuth("github")}
            className="card-flat"
            style={{
              width: "100%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
              gap: 8,
              padding: "10px 16px",
              fontSize: 13,
              fontWeight: 500,
              cursor: "pointer",
              color: "var(--text-secondary)",
              transition: "border-color 0.2s",
            }}
          >
            <svg
              style={{ width: 16, height: 16 }}
              viewBox="0 0 24 24"
              fill="currentColor"
            >
              <path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z" />
            </svg>
            Continue with GitHub
          </button>
        </div>

        {/* Divider */}
        <div
          style={{
            position: "relative",
            marginBottom: 24,
            textAlign: "center",
          }}
        >
          <div
            style={{
              position: "absolute",
              inset: 0,
              display: "flex",
              alignItems: "center",
            }}
          >
            <div
              style={{
                width: "100%",
                borderTop: "1px solid var(--border-subtle)",
              }}
            />
          </div>
          <span
            className="micro-label"
            style={{
              position: "relative",
              padding: "0 12px",
              background: "var(--bg-primary)",
              fontSize: 10,
            }}
          >
            or
          </span>
        </div>

        {/* Form */}
        <form onSubmit={handleSignup}>
          {error && (
            <div
              style={{
                padding: 12,
                borderRadius: 8,
                marginBottom: 16,
                fontSize: 13,
                background: "rgba(199, 116, 74, 0.08)",
                color: "var(--accent-orange)",
                border: "1px solid rgba(199, 116, 74, 0.15)",
              }}
            >
              {error}
            </div>
          )}

          <div style={{ marginBottom: 16 }}>
            <label
              htmlFor="org"
              className="micro-label"
              style={{ display: "block", marginBottom: 8, fontSize: 11 }}
            >
              Organization name
            </label>
            <input
              id="org"
              type="text"
              value={orgName}
              onChange={(e) => setOrgName(e.target.value)}
              required
              placeholder="Acme Inc."
              style={{
                width: "100%",
                padding: "10px 14px",
                fontSize: 14,
                borderRadius: 8,
                border: "1px solid var(--border-light)",
                background: "var(--bg-primary)",
                color: "var(--text-primary)",
                outline: "none",
                transition: "border-color 0.2s",
              }}
              onFocus={(e) =>
                (e.currentTarget.style.borderColor = "var(--accent-purple)")
              }
              onBlur={(e) =>
                (e.currentTarget.style.borderColor = "var(--border-light)")
              }
            />
          </div>

          <div style={{ marginBottom: 16 }}>
            <label
              htmlFor="email"
              className="micro-label"
              style={{ display: "block", marginBottom: 8, fontSize: 11 }}
            >
              Work email
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              placeholder="you@company.com"
              style={{
                width: "100%",
                padding: "10px 14px",
                fontSize: 14,
                borderRadius: 8,
                border: "1px solid var(--border-light)",
                background: "var(--bg-primary)",
                color: "var(--text-primary)",
                outline: "none",
                transition: "border-color 0.2s",
              }}
              onFocus={(e) =>
                (e.currentTarget.style.borderColor = "var(--accent-purple)")
              }
              onBlur={(e) =>
                (e.currentTarget.style.borderColor = "var(--border-light)")
              }
            />
          </div>

          <div style={{ marginBottom: 24 }}>
            <label
              htmlFor="password"
              className="micro-label"
              style={{ display: "block", marginBottom: 8, fontSize: 11 }}
            >
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              minLength={8}
              placeholder="Min. 8 characters"
              style={{
                width: "100%",
                padding: "10px 14px",
                fontSize: 14,
                borderRadius: 8,
                border: "1px solid var(--border-light)",
                background: "var(--bg-primary)",
                color: "var(--text-primary)",
                outline: "none",
                transition: "border-color 0.2s",
              }}
              onFocus={(e) =>
                (e.currentTarget.style.borderColor = "var(--accent-purple)")
              }
              onBlur={(e) =>
                (e.currentTarget.style.borderColor = "var(--border-light)")
              }
            />
          </div>

          <button
            type="submit"
            disabled={loading}
            className="btn-research"
            style={{
              width: "100%",
              height: 42,
              fontSize: 14,
              opacity: loading ? 0.6 : 1,
            }}
          >
            {loading ? "Creating account..." : "Create account"}
          </button>
        </form>

        <p
          style={{
            marginTop: 28,
            textAlign: "center",
            fontSize: 13,
            color: "var(--text-muted)",
          }}
        >
          Already have an account?{" "}
          <Link
            href="/login"
            style={{
              fontWeight: 500,
              color: "var(--text-primary)",
              textDecoration: "underline",
              textUnderlineOffset: 3,
            }}
          >
            Sign in
          </Link>
        </p>
      </div>
    </div>
  );
}
