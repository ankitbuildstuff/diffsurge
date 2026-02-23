"use client";

import { useEffect } from "react";

export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error("Global error:", error);
  }, [error]);

  return (
    <html lang="en">
      <body className="bg-white text-zinc-950 antialiased">
        <div className="flex min-h-screen flex-col items-center justify-center p-8">
          <div className="mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-red-50">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="28"
              height="28"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
              className="text-red-500"
            >
              <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3" />
              <path d="M12 9v4" />
              <path d="M12 17h.01" />
            </svg>
          </div>
          <h2 className="text-lg font-semibold">Something went wrong</h2>
          <p className="mt-1 max-w-sm text-center text-sm text-zinc-500">
            A critical error occurred. Please refresh the page.
          </p>
          <button
            onClick={reset}
            className="mt-6 rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800"
          >
            Try again
          </button>
        </div>
      </body>
    </html>
  );
}
