import { readApiErrorMessage } from './errors';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export interface ScheduleEntry {
  id: number;
  office_hour_id: number;
  course_id: number;
  course_code: string;
  course_name: string;
  ta_username: string;
  ta_id: number;
  day_of_week: number;
  start_time: string;
  end_time: string;
  location: string;
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('token') ?? localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

export async function listStudentSchedule(): Promise<ScheduleEntry[]> {
  const res = await fetch(`${BASE_URL}/api/student/schedule`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch schedule'));
  }
  const body = (await res.json()) as { schedule?: ScheduleEntry[] };
  return body.schedule ?? [];
}

export async function addToStudentSchedule(officeHourID: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/student/schedule`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ office_hour_id: officeHourID }),
  });
  if (!res.ok && res.status !== 200) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to add to schedule'));
  }
}

export async function removeFromStudentSchedule(id: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/student/schedule/${id}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok && res.status !== 204) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to remove from schedule'));
  }
}