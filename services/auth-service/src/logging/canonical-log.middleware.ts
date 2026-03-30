import { randomUUID } from 'crypto';
import { Injectable, NestMiddleware } from '@nestjs/common';
import { Request, Response, NextFunction } from 'express';
import { ILogContext } from './log-context.interface';
import { logContextStorage } from './log-context.service';

/**
 * Canonical log middleware — emits ONE structured JSON line per request.
 *
 * Placement in the request lifecycle:
 *   [this middleware] → guards → pipes → handler → exception filter → res.finish
 *
 * The middleware opens an AsyncLocalStorage context so every downstream
 * guard, pipe, or handler can call LogContextService.enrich() to add
 * business fields.  All accumulated fields are flushed in a single log line
 * when the response is finished.
 *
 * Header contract (set by the gateway today, nginx in the future):
 *   X-Request-ID — unique per HTTP request; honoured if present, generated
 *                  here as a fallback so the service is self-contained.
 *   X-Trace-ID   — distributed trace token propagated across services.
 */
@Injectable()
export class CanonicalLogMiddleware implements NestMiddleware {
  use(req: Request, res: Response, next: NextFunction): void {
    // --- 1. Resolve IDs ---
    // Prefer the values injected by the gateway.  If the service is called
    // directly (e.g. during local development / integration tests) we
    // generate a local ID so every request still gets a log line.
    const requestId =
      (req.headers['x-request-id'] as string | undefined) ?? randomUUID();
    const traceId =
      (req.headers['x-trace-id'] as string | undefined) ?? requestId;

    // Echo X-Request-ID back so callers can correlate without reading logs.
    res.setHeader('X-Request-ID', requestId);

    // --- 2. Seed the log context ---
    const ctx: ILogContext = {
      requestId,
      traceId,
      method: req.method,
      path: req.path,
      startTime: Date.now(),
    };

    // --- 3. Run the rest of the request inside the ALS context ---
    // Everything called from next() — guards, pipes, handlers, filters —
    // shares this store via logContextStorage.getStore().
    logContextStorage.run(ctx, () => {
      res.on('finish', () => {
        const store = logContextStorage.getStore();
        if (!store) return;

        const logLine = {
          ...store,
          statusCode: res.statusCode,
          durationMs: Date.now() - store.startTime,
          // startTime is an internal marker; exclude it from the emitted line.
          startTime: undefined,
        };

        // Replace with your logger (Pino / Winston) when ready.
        // console.log is used here to keep the module dependency-free.
        console.log(JSON.stringify(logLine));
      });

      next();
    });
  }
}
