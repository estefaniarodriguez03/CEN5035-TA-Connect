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
