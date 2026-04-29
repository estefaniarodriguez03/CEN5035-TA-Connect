import { readApiErrorMessage } from './errors';

const BASE_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

export interface StudentCourse {
  id: number;
  code: string;
  name: string;
  color?: string;
}

function getAuthHeaders(): HeadersInit {
  const token = sessionStorage.getItem('token') ?? localStorage.getItem('token');
  return {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };
}

/**
 * Get all courses the authenticated student is enrolled in
 */
export async function getStudentCourses(): Promise<StudentCourse[]> {
  const res = await fetch(`${BASE_URL}/api/student/courses`, {
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to fetch student courses'));
  }
  const body = (await res.json()) as { courses?: StudentCourse[] };
  return body.courses ?? [];
}

/**
 * Add a course to the student's enrolled courses
 */
export async function addStudentCourse(courseId: number): Promise<StudentCourse> {
  const res = await fetch(`${BASE_URL}/api/student/courses`, {
    method: 'POST',
    headers: getAuthHeaders(),
    body: JSON.stringify({ course_id: courseId }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to add student course'));
  }
  const body = (await res.json()) as { course: StudentCourse };
  return body.course;
}

/**
 * Remove a course from the student's enrolled courses
 */
export async function removeStudentCourse(courseId: number): Promise<void> {
  const res = await fetch(`${BASE_URL}/api/student/courses/${courseId}`, {
    method: 'DELETE',
    headers: getAuthHeaders(),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to remove student course'));
  }
}

/**
 * Update the student's enrolled courses (batch operation)
 */
export async function updateStudentCourses(courseIds: number[]): Promise<StudentCourse[]> {
  const res = await fetch(`${BASE_URL}/api/student/courses/batch`, {
    method: 'PUT',
    headers: getAuthHeaders(),
    body: JSON.stringify({ course_ids: courseIds }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(readApiErrorMessage(err, 'Failed to update student courses'));
  }
  const body = (await res.json()) as { courses: StudentCourse[] };
  return body.courses;
}
