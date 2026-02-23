import Link from "next/link";

export default function NotFound() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-zinc-50 p-8">
      <div className="mb-4 text-6xl font-bold text-zinc-200">404</div>
      <h2 className="text-lg font-semibold text-zinc-900">Page not found</h2>
      <p className="mt-1 max-w-sm text-center text-sm text-zinc-500">
        The page you&apos;re looking for doesn&apos;t exist or has been moved.
      </p>
      <div className="mt-6 flex gap-3">
        <Link
          href="/"
          className="rounded-lg bg-zinc-900 px-4 py-2 text-sm font-medium text-white hover:bg-zinc-800"
        >
          Go home
        </Link>
        <Link
          href="/dashboard"
          className="rounded-lg border border-zinc-200 bg-white px-4 py-2 text-sm font-medium text-zinc-700 hover:bg-zinc-50"
        >
          Dashboard
        </Link>
      </div>
    </div>
  );
}
