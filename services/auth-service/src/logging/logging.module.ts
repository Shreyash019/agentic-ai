import { Module, MiddlewareConsumer, NestModule } from '@nestjs/common';
import { CanonicalLogMiddleware } from './canonical-log.middleware';
import { LogContextService } from './log-context.service';

/**
 * Apply the canonical log middleware to every route via '*path'.
 * Export LogContextService so any feature module can inject it to enrich
 * the current request's log context.
 */
@Module({
  providers: [LogContextService],
  exports: [LogContextService],
})
export class LoggingModule implements NestModule {
  configure(consumer: MiddlewareConsumer): void {
    consumer.apply(CanonicalLogMiddleware).forRoutes('*path');
  }
}
