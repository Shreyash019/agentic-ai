import { AsyncLocalStorage } from 'async_hooks';
import { Injectable } from '@nestjs/common';
import { ILogContext } from './log-context.interface';

/**
 * Module-level ALS instance shared by the middleware and this service.
 * Not exported from the module on purpose — consumers should use the
 * service methods rather than touching the store directly.
 */
export const logContextStorage = new AsyncLocalStorage<ILogContext>();

/**
 * Singleton service that lets any guard, pipe, or handler enrich the
 * canonical log context for the current request.
 *
 * Usage in a handler:
 *   constructor(private readonly logCtx: LogContextService) {}
 *
 *   handleSomething() {
 *     this.logCtx.enrich({ userId: user.id, action: 'login' });
 *   }
 */
@Injectable()
export class LogContextService {
  /**
   * Merge additional fields into the current request's log context.
   * Safe to call from any async context within the request lifecycle.
   * Does nothing if called outside a request (e.g. during bootstrap).
   */
  enrich(fields: Partial<ILogContext>): void {
    const ctx = logContextStorage.getStore();
    if (ctx) {
      Object.assign(ctx, fields);
    }
  }

  /**
   * Returns the log context for the current request, or undefined when
   * called outside a request context (e.g. scheduled tasks).
   */
  get(): ILogContext | undefined {
    return logContextStorage.getStore();
  }
}
