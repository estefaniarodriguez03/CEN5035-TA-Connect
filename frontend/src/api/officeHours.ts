import { readApiErrorMessage } from './errors';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export interface OfficeHour {
  id: number;
  ta_id: number;
  ta_username?: string;
  course_id: number;
  day_of_week: number;
  start_time: string;
  end_time: string;
  location: string;
}

export interface CreateOfficeHourPayload {
  course_id: number;
  day_of_week: number;
  start_time: string;
  end_time: string;
  location: string;
}

// Sunday=0 to match the backend's 0-6 encoding
export const DAY_NAMES = ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'];

// Alias so TADashboard and MyOfficeHoursPage can import listOfficeHoursByTA
export const listOfficeHoursByTA = getOfficeHoursByTA;

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('token') ?? localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function getOfficeHoursByTA(taId: number): Promise<OfficeHour[]> {
  const res = await fetch(`${BASE_URL}/api/office-hours/ta/${taId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch office hours'));
  }
  return res.json();
}

export async function listOfficeHoursByCourse(courseId: number): Promise<OfficeHour[]> {
  const res = await fetch(`${BASE_URL}/api/office-hours/course/${courseId}`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch office hours for course'));
  }
  return res.json();
}

export async function createOfficeHour(payload: CreateOfficeHourPayload): Promise<OfficeHour> {
  const res = await fetch(`${BASE_URL}/api/office-hours`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to create office hour'));
  }
  return res.json();
}

export async function updateOfficeHour(id: number, payload: CreateOfficeHourPayload): Promise<OfficeHour> {
  const res = await fetch(`${BASE_URL}/api/office-hours/${id}`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify(payload),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to update office hour'));
  }
  return res.json();
}

export async function deleteOfficeHour(id: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/office-hours/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok && res.status !== 204) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to delete office hour'));
  }
}