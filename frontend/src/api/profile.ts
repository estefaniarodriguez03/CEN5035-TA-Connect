import { readApiErrorMessage } from "./errors";

export interface Course {
  id: number;
  code: string;
  name: string;
}

export interface UpdateProfileRequest {
  username?: string;
  courses?: Course[];
}

export interface UpdateProfileResponse {
  id: number;
  username: string;
  email: string;
  role: string;
  courses?: Course[];
}

const API_BASE = "/api";

export async function updateUserProfile(
  userId: number,
  data: UpdateProfileRequest
): Promise<UpdateProfileResponse> {
  const token = sessionStorage.getItem("token") ?? localStorage.getItem("token");

  const res = await fetch(`${API_BASE}/users/${userId}/profile`, {
    method: "PUT",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
    },
    body: JSON.stringify(data),
  });

  if (!res.ok) {
    const responseData = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(responseData, "Failed to update profile"));
  }

  return res.json();
}
