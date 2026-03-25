import { createContext, useContext, useState, useEffect } from "react";

import type { UserRole } from "../api/auth";

interface User {
  id: number;
  username: string;
  email: string;
  role: UserRole;
}

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (token: string, user: User) => void;
  logout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(null);

  useEffect(() => {
    const storedToken = sessionStorage.getItem("token") ?? localStorage.getItem("token");
    const storedUser = sessionStorage.getItem("user") ?? localStorage.getItem("user");

    if (storedToken && storedUser) {
      // Normalize to tab-scoped auth so different tabs can use different accounts.
      sessionStorage.setItem("token", storedToken);
      sessionStorage.setItem("user", storedUser);
      localStorage.removeItem("token");
      localStorage.removeItem("user");
      setToken(storedToken);
      setUser(JSON.parse(storedUser));
    }
  }, []);

  function handleLogin(token: string, user: User) {
    sessionStorage.setItem("token", token);
    sessionStorage.setItem("user", JSON.stringify(user));
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setToken(token);
    setUser(user);
  }

  function handleLogout() {
    sessionStorage.removeItem("token");
    sessionStorage.removeItem("user");
    localStorage.removeItem("token");
    localStorage.removeItem("user");
    setToken(null);
    setUser(null);
  }

  return (
    <AuthContext.Provider
      value={{ user, token, login: handleLogin, logout: handleLogout }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
