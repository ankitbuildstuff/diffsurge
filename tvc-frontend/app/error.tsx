"use client";

import { useEffect } from "react";
import { AlertTriangle } from "lucide-react";
import { Button } from "@/components/ui/button";

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Route error:", error);
  }, [error]);

  return (
    <div className="flex min-h-[60vh] flex-col items-center justify-center p-8">
      <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-red-50">
        <AlertTriangle size={28} className="text-red-500" />
      </div>
      <h2 className="text-lg font-semibold text-zinc-900">
        Something went wrong
      </h2>
      <p className="mt-1 max-w-sm text-center text-sm text-zinc-500">
        An unexpected error occurred. Please try again.
      </p>
      <Button variant="secondary" size="sm" className="mt-6" onClick={reset}>
        Try again
      </Button>
    </div>
  );
}
