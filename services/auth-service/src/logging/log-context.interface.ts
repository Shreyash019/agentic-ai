/**
 * Fields accumulated throughout the lifetime of a single request.
 * The canonical log middleware emits exactly one JSON line containing
 * all of these fields when the response is finished.
 */
export interface ILogContext {
  // --- Identity ---
  /** Unique ID for this HTTP request. Sourced from X-Request-ID header. */
  requestId: string;
  /**
   * Distributed trace token. Equals requestId today.
   * When nginx becomes the LB it will inject X-Trace-ID itself; no code
   * change is needed here — the middleware just forwards whatever arrives.
   */
  traceId: string;

  // --- Request ---
  method: string;
  path: string;
  /** Epoch ms when the middleware received the request. */
  startTime: number;

  // --- Response (filled on res.finish) ---
  statusCode?: number;
  /** Total time from middleware entry to response flush, in milliseconds. */
  durationMs?: number;

  // --- Business fields (enriched by guards / handlers) ---
  userId?: string;
  [key: string]: unknown; // allow ad-hoc fields
}
