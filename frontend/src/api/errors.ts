/** Parses standard API error JSON: { code, message, details }. Falls back to legacy { error }. */
export function readApiErrorMessage(data: unknown, fallback: string): string {
  if (!data || typeof data !== "object") {
    return fallback;
  }
  const o = data as Record<string, unknown>;
  if (typeof o.message === "string" && o.message) {
    return o.message;
  }
  if (typeof o.error === "string" && o.error) {
    return o.error;
  }
  return fallback;
}

export type ApiErrorInfo = { code: string; message: string; details: unknown };

/** Full error body for branching (e.g. code === "queue_empty"). */
export async function readApiErrorFromResponse(
  res: Response,
  fallback: string
): Promise<ApiErrorInfo> {
  const raw = await res.json().catch(() => ({}));
  const data = raw && typeof raw === "object" ? (raw as Record<string, unknown>) : {};
  const code = typeof data.code === "string" && data.code ? data.code : "unknown";
  return {
    code,
    message: readApiErrorMessage(raw, fallback),
    details: "details" in data ? data.details : null,
  };
}
