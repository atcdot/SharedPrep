import type { ReactNode } from "react";
import { Navigate } from "react-router-dom";
import { useAuth } from "@/lib/auth";

export function RequireAuth({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();

  if (loading) {
    return <p className="text-muted-foreground pt-8 text-center">Loading...</p>;
  }
  if (!user) {
    return <Navigate to="/" replace />;
  }
  return children;
}
