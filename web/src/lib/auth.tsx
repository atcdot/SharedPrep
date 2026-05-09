import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from "react";

export interface User {
  id: string;
  firstName: string;
  lastName: string;
  username: string;
  photoUrl: string;
  displayName: string | null;
  showTelegram: boolean;
}

interface AuthState {
  user: User | null;
  loading: boolean;
  login: (idToken: string) => Promise<void>;
  logout: () => Promise<void>;
  updateProfile: (data: { display_name?: string; show_telegram?: boolean }) => Promise<void>;
}

const AuthContext = createContext<AuthState>({
  user: null,
  loading: true,
  login: async () => {},
  logout: async () => {},
  updateProfile: async () => {},
});

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/auth/me", { credentials: "include" })
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data) {
          setUser({
            id: data.id,
            firstName: data.firstName,
            lastName: data.lastName,
            username: data.username,
            photoUrl: data.photoUrl,
            displayName: data.displayName ?? null,
            showTelegram: data.showTelegram ?? true,
          });
        }
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (idToken: string) => {
    const res = await fetch("/auth/telegram", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ id_token: idToken }),
      credentials: "include",
    });
    if (!res.ok) throw new Error("Login failed");
    const json = await res.json();
    setUser({
      id: json.id,
      firstName: json.firstName,
      lastName: json.lastName,
      username: json.username,
      photoUrl: json.photoUrl,
      displayName: json.displayName ?? null,
      showTelegram: json.showTelegram ?? true,
    });
  }, []);

  const logout = useCallback(async () => {
    await fetch("/auth/logout", { method: "POST", credentials: "include" });
    setUser(null);
  }, []);

  const updateProfile = useCallback(async (data: { display_name?: string; show_telegram?: boolean }) => {
    const res = await fetch("/auth/profile", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
      credentials: "include",
    });
    if (!res.ok) throw new Error("Failed to update profile");
    setUser((prev) => {
      if (!prev) return prev;
      return {
        ...prev,
        ...(data.display_name !== undefined && { displayName: data.display_name }),
        ...(data.show_telegram !== undefined && { showTelegram: data.show_telegram }),
      };
    });
  }, []);

  return (
    <AuthContext value={{ user, loading, login, logout, updateProfile }}>
      {children}
    </AuthContext>
  );
}

export function useAuth() {
  return useContext(AuthContext);
}
