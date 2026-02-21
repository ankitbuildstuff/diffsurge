"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { Mail, CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { LoadingPage } from "@/components/ui/loading-spinner";
import { createClient } from "@/lib/supabase/client";

export default function VerifyEmailPage() {
  const [loading, setLoading] = useState(true);
  const [verified, setVerified] = useState(false);
  const [email, setEmail] = useState<string | null>(null);
  const router = useRouter();
  const supabase = createClient();

  useEffect(() => {
    const checkAuth = async () => {
      const {
        data: { session },
      } = await supabase.auth.getSession();

      if (session?.user) {
        setEmail(session.user.email || null);

        // Check if email is already confirmed
        if (session.user.email_confirmed_at) {
          setVerified(true);
        }
      }

      setLoading(false);
    };

    checkAuth();
  }, [supabase]);

  const handleResend = async () => {
    if (!email) return;

    try {
      const { error } = await supabase.auth.resend({
        type: "signup",
        email,
      });

      if (error) throw error;
    } catch (error: any) {
      console.error("Failed to resend verification email:", error);
    }
  };

  if (loading) {
    return <LoadingPage />;
  }

  if (verified) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-zinc-50 px-4">
        <Card className="w-full max-w-md">
          <CardHeader>
            <div className="mb-4 flex justify-center">
              <div className="rounded-full bg-emerald-50 p-3">
                <CheckCircle2 className="h-6 w-6 text-emerald-600" />
              </div>
            </div>
            <CardTitle className="text-center">Email verified!</CardTitle>
            <CardDescription className="text-center">
              Your email has been successfully verified.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Button
              className="w-full"
              onClick={() => router.push("/dashboard")}
            >
              Go to Dashboard
            </Button>
          </CardContent>
        </Card>
      </div>
    );
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-zinc-50 px-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <div className="mb-4 flex justify-center">
            <div className="rounded-full bg-blue-50 p-3">
              <Mail className="h-6 w-6 text-blue-600" />
            </div>
          </div>
          <CardTitle className="text-center">Verify your email</CardTitle>
          <CardDescription className="text-center">
            We&apos;ve sent a verification link to{" "}
            {email && <strong>{email}</strong>}
          </CardDescription>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-center text-sm text-zinc-600">
            Click the link in the email to confirm your account. If you
            don&apos;t see it, check your spam folder.
          </p>
          <div className="flex flex-col gap-2">
            <Button
              onClick={handleResend}
              variant="secondary"
              className="w-full"
            >
              Resend verification email
            </Button>
            <Link href="/login" className="w-full">
              <Button variant="ghost" className="w-full">
                Back to login
              </Button>
            </Link>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
