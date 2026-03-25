export interface QueueEntry {
  id: number;
  queue_id: number;
  student_id: number;
  position: number;
  joined_at: string;
  username: string;
}

export interface QueueData {
  id: number;
  course_id: number;
  ta_id: number;
  status: string;
  created_at: string;
  entries: QueueEntry[];
}

export interface JoinQueueResponse {
  id: number;
  queue_id: number;
  position: number;
  joined_at: string;
}

export interface QueueEvent {
  type: "STUDENT_JOINED" | "STUDENT_LEFT" | "QUEUE_UPDATED" | "STUDENT_SERVED";
  queue_id: number;
  payload?: any;
}

export interface NextQueueResponse {
  queue_id: number;
  status: "in_session";
  student: QueueEntry;
}

export interface CreateQueueResponse {
  id: number;
  course_id: number;
  ta_id: number;
  status: "open";
  created_at: string;
}

export type QueueStatus = "open" | "paused" | "closed";

export interface UpdateQueueStatusResponse {
  id: number;
  status: QueueStatus;
}

type ActiveQueueMap = Record<string, number>;

const API_BASE = "/api";
const ACTIVE_QUEUE_STORAGE_KEY = "activeQueuesByCourse";
const ACTIVE_OFFICE_HOUR_QUEUE_STORAGE_KEY = "activeQueuesByOfficeHour";

function getAuthToken(): string {
  const token = sessionStorage.getItem("token") ?? localStorage.getItem("token");
  return token || "";
}

async function parseErrorMessage(res: Response, fallback: string): Promise<string> {
  try {
    const data = await res.json();
    if (data?.error && typeof data.error === "string") {
      return data.error;
    }
  } catch {
    // ignore JSON parse errors and use fallback
  }
  return fallback;
}

export async function joinQueue(queueID: number): Promise<JoinQueueResponse> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/join`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to join queue");
    throw new Error(message);
  }

  return res.json();
}

export async function leaveQueue(queueID: number): Promise<void> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/leave`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to leave queue");
    throw new Error(message);
  }
}

export async function getQueue(queueID: number): Promise<QueueData> {
  const res = await fetch(`${API_BASE}/queues/${queueID}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to get queue data");
    throw new Error(message);
  }

  return res.json();
}

export async function getQueueOrNull(queueID: number): Promise<QueueData | null> {
  const res = await fetch(`${API_BASE}/queues/${queueID}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (res.status === 404) {
    return null;
  }

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to get queue data");
    throw new Error(message);
  }

  return res.json();
}

export async function getActiveQueueByCourse(courseID: number): Promise<QueueData | null> {
  const res = await fetch(`${API_BASE}/queues/active?course_id=${courseID}`, {
    method: "GET",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (res.status === 404) {
    return null;
  }

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to get active queue data");
    throw new Error(message);
  }

  return res.json();
}

export async function nextQueueStudent(queueID: number): Promise<NextQueueResponse> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/next`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to advance queue");
    throw new Error(message);
  }

  return res.json();
}

export async function createQueue(courseID: number): Promise<CreateQueueResponse> {
  const res = await fetch(`${API_BASE}/queues`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
    body: JSON.stringify({ course_id: courseID }),
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to create queue");
    throw new Error(message);
  }

  return res.json();
}

export async function updateQueueStatus(queueID: number, status: QueueStatus): Promise<UpdateQueueStatusResponse> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/status`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
    body: JSON.stringify({ status }),
  });

  if (!res.ok) {
    const message = await parseErrorMessage(res, "Failed to update queue status");
    throw new Error(message);
  }

  return res.json();
}

function readActiveQueueMap(): ActiveQueueMap {
  const raw = localStorage.getItem(ACTIVE_QUEUE_STORAGE_KEY);
  if (!raw) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw) as ActiveQueueMap;
    return parsed ?? {};
  } catch {
    return {};
  }
}

function writeActiveQueueMap(map: ActiveQueueMap): void {
  localStorage.setItem(ACTIVE_QUEUE_STORAGE_KEY, JSON.stringify(map));
}

function readActiveOfficeHourQueueMap(): ActiveQueueMap {
  const raw = localStorage.getItem(ACTIVE_OFFICE_HOUR_QUEUE_STORAGE_KEY);
  if (!raw) {
    return {};
  }
  try {
    const parsed = JSON.parse(raw) as ActiveQueueMap;
    return parsed ?? {};
  } catch {
    return {};
  }
}

function writeActiveOfficeHourQueueMap(map: ActiveQueueMap): void {
  localStorage.setItem(ACTIVE_OFFICE_HOUR_QUEUE_STORAGE_KEY, JSON.stringify(map));
}

function getOfficeHourKey(courseCode: string, timeRange: string): string {
  return `${courseCode}|${timeRange}`;
}

export function getActiveQueueForCourse(courseCode: string): number | null {
  const map = readActiveQueueMap();
  const queueID = map[courseCode];
  return typeof queueID === "number" ? queueID : null;
}

export function setActiveQueueForCourse(courseCode: string, queueID: number): void {
  const map = readActiveQueueMap();
  map[courseCode] = queueID;
  writeActiveQueueMap(map);
}

export function clearActiveQueueForCourse(courseCode: string): void {
  const map = readActiveQueueMap();
  delete map[courseCode];
  writeActiveQueueMap(map);
}

export function getActiveQueueForOfficeHour(courseCode: string, timeRange: string): number | null {
  const key = getOfficeHourKey(courseCode, timeRange);
  const map = readActiveOfficeHourQueueMap();
  const queueID = map[key];
  return typeof queueID === "number" ? queueID : null;
}

export function setActiveQueueForOfficeHour(courseCode: string, timeRange: string, queueID: number): void {
  const key = getOfficeHourKey(courseCode, timeRange);
  const map = readActiveOfficeHourQueueMap();
  map[key] = queueID;
  writeActiveOfficeHourQueueMap(map);
}

export function clearActiveQueueForOfficeHour(courseCode: string, timeRange: string): void {
  const key = getOfficeHourKey(courseCode, timeRange);
  const map = readActiveOfficeHourQueueMap();
  delete map[key];
  writeActiveOfficeHourQueueMap(map);
}

export function subscribeToQueueEvents(
  queueID: number,
  onEvent: (event: QueueEvent) => void,
  onError: (error: Error) => void
): () => void {
  const eventSource = new EventSource(`${API_BASE}/queues/${queueID}/events`);

  eventSource.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      onEvent(data);
    } catch (error) {
      onError(new Error("Failed to parse event"));
    }
  };

  eventSource.onerror = () => {
    onError(new Error("SSE connection error"));
    eventSource.close();
  };

  // Return unsubscribe function
  return () => {
    eventSource.close();
  };
}
