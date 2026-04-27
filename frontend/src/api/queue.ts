import { readApiErrorMessage, readApiErrorFromResponse } from "./errors";

export interface QueueEntry {
  id: number;
  queue_id: number;
  student_id: number;
  position: number;
  joined_at: string;
  username: string;
  /** Present when server sends ETA; seconds until this position is seen. */
  estimated_wait_seconds?: number;
}

export interface QueueData {
  id: number;
  course_id: number;
  ta_id: number;
  status: string;
  created_at: string;
  /** True when there are zero entries (explicit from API; may be derived from entries.length). */
  is_empty?: boolean;
  /** Max estimated wait for last in line (seconds). */
  estimated_wait_time_seconds?: number;
  average_session_duration_seconds?: number;
  ta_average_session_duration_seconds?: number;
  entries: QueueEntry[];
}

export interface JoinQueueResponse {
  id: number;
  queue_id: number;
  position: number;
  joined_at: string;
}

export interface QueueStateChangePayload {
  previous_status: QueueStatus;
  status: QueueStatus;
}

export interface StudentUpNextPayload {
  student_id: number;
  position: number;
}

export interface AnnouncementSentPayload {
  id: number;
  message: string;
  ta_id: number;
  created_at: string;
}

export interface QueueEvent {
  type:
    | "STUDENT_JOINED"
    | "STUDENT_LEFT"
    | "QUEUE_UPDATED"
    | "STUDENT_SERVED"
    | "STUDENT_UP_NEXT"
    | "ANNOUNCEMENT_SENT"
    | "QUEUE_STATE_CHANGED";
  queue_id: number;
  payload?:
    | QueueStateChangePayload
    | StudentUpNextPayload
    | AnnouncementSentPayload
    | Record<string, unknown>;
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
    return readApiErrorMessage(data, fallback);
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

export async function postQueueAnnouncement(queueID: number, message: string): Promise<void> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/announcement`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
    body: JSON.stringify({ message }),
  });

  if (!res.ok) {
    const errMsg = await parseErrorMessage(res, "Failed to send announcement");
    throw new Error(errMsg);
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

export type NextQueueError = Error & { code: string; details: unknown };

export async function nextQueueStudent(queueID: number): Promise<NextQueueResponse> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/next`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getAuthToken()}`,
    },
  });

  if (!res.ok) {
    const e = await readApiErrorFromResponse(res, "Failed to advance queue");
    const err = new Error(e.message) as NextQueueError;
    err.code = e.code;
    err.details = e.details;
    throw err;
  }

  return res.json();
}

/** User-facing line + caption for the “browse” wait card (not yet in queue). */
export function browseWaitDisplay(data: QueueData): { line: string; sub: string } {
  const n = data.entries.length;
  const isEmpty = data.is_empty ?? n === 0;
  if (data.status === "closed") {
    return { line: "—", sub: "This queue is closed" };
  }
  if (data.status === "paused") {
    return { line: "—", sub: "The queue is paused" };
  }
  if (isEmpty || n === 0) {
    return { line: "No students waiting", sub: "The queue is open — be the first to join" };
  }
  const est = data.estimated_wait_time_seconds;
  if (typeof est === "number" && est > 0) {
    const min = Math.max(1, Math.ceil(est / 60));
    return { line: `~${min} min`, sub: `Estimated time if you joined the end of the line (${n} in queue)` };
  }
  const legacyMin = n * 4;
  return {
    line: legacyMin === 0 ? "0 min" : `~${legacyMin} min`,
    sub: `Based on ${n} student(s) in queue`,
  };
}

/** Wait label for the in-queue card (uses per-entry seconds when available). */
export function myWaitMinutesFromEntry(entry: QueueEntry | undefined): number {
  if (!entry) {
    return 0;
  }
  const s = entry.estimated_wait_seconds;
  if (typeof s === "number" && s >= 0) {
    return Math.ceil(s / 60);
  }
  return Math.max(0, (entry.position - 1) * 4);
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
  return updateQueueState(queueID, status);
}

export async function updateQueueState(queueID: number, status: QueueStatus): Promise<UpdateQueueStatusResponse> {
  const res = await fetch(`${API_BASE}/queues/${queueID}/state`, {
    method: "PATCH",
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

const SSE_EVENT_NAMES: QueueEvent["type"][] = [
  "STUDENT_JOINED",
  "STUDENT_LEFT",
  "STUDENT_SERVED",
  "STUDENT_UP_NEXT",
  "ANNOUNCEMENT_SENT",
  "QUEUE_UPDATED",
  "QUEUE_STATE_CHANGED",
];

export function subscribeToQueueEvents(
  queueID: number,
  onEvent: (event: QueueEvent) => void,
  onError: (error: Error) => void
): () => void {
  const eventSource = new EventSource(`${API_BASE}/queues/${queueID}/events`);

  const handler = (msg: MessageEvent) => {
    try {
      const data: QueueEvent = JSON.parse(msg.data);
      onEvent(data);
    } catch {
      onError(new Error("Failed to parse event"));
    }
  };

  SSE_EVENT_NAMES.forEach((name) => {
    eventSource.addEventListener(name, handler as EventListener);
  });

  eventSource.onerror = () => {
    onError(new Error("SSE connection error"));
  };

  return () => {
    SSE_EVENT_NAMES.forEach((name) => {
      eventSource.removeEventListener(name, handler as EventListener);
    });
    eventSource.close();
  };
}
