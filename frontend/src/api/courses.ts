import { readApiErrorMessage } from './errors';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export interface TACourse {
  id: number;
  code: string;
  name: string;
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('token') ?? localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function listMyTACourses(): Promise<TACourse[]> {
  const res = await fetch(`${BASE_URL}/api/ta/courses`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch TA courses'));
  }
  const body = (await res.json()) as { courses?: TACourse[] };
  return body.courses ?? [];
}

export async function addMyTACourse(code: string, name?: string): Promise<TACourse> {
  const res = await fetch(`${BASE_URL}/api/ta/courses`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ code, name: name ?? '' }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to add TA course'));
  }
  const body = (await res.json()) as { course: TACourse };
  return body.course;
}

export async function listAllCourses(): Promise<TACourse[]> {
  const res = await fetch(`${BASE_URL}/api/courses`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch courses'));
  }
  const body = (await res.json()) as { courses?: TACourse[] };
  return body.courses ?? [];
}