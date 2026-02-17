export type UserRole = "student" | "ta";

export interface AuthResponse {
  token: string;
  user: {
    id: number;
    username: string;
    email: string;
    role: UserRole;
  };
}

const API_BASE = "/api";

export async function login(email: string, password: string): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/login`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ email, password }),
  });

  if (!res.ok) {
    throw new Error("Login failed");
  }

  return res.json();
}

export async function register(
  username: string,
  email: string,
  password: string,
  role: "student" | "ta"
): Promise<AuthResponse> {
  const res = await fetch(`${API_BASE}/register`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, email, password, role }),
  });

  if (!res.ok) {
    throw new Error("Registration failed");
  }

  return res.json();
}
