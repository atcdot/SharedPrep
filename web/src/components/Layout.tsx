import type { ReactNode } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "@/lib/auth";

export function Layout({ children }: { children: ReactNode }) {
  const { user, logout } = useAuth();

  return (
    <div className="min-h-screen bg-background">
      <header className="border-b">
        <div className="mx-auto flex max-w-3xl items-center justify-between px-4 py-3">
          <Link to="/" className="text-xl font-bold tracking-tight">
            SharedPrep
          </Link>
          {user && (
            <div className="flex items-center gap-3">
              <Link
                to="/settings"
                className="text-sm text-muted-foreground hover:text-foreground"
              >
                Settings
              </Link>
              <span className="text-sm text-muted-foreground">
                {user.displayName ?? user.firstName}
              </span>
              {user.photoUrl && (
                <img
                  src={user.photoUrl}
                  alt={user.firstName}
                  className="h-7 w-7 rounded-full"
                />
              )}
              <button
                onClick={logout}
                className="text-xs text-muted-foreground hover:text-foreground"
              >
                Logout
              </button>
            </div>
          )}
        </div>
      </header>
      <main className="mx-auto max-w-3xl px-4 py-6">{children}</main>
    </div>
  );
}
