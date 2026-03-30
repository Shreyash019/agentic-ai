import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);
  const servicePort: number | undefined = Number(process.env.AUTH_SERVICE_PORT);
  if (!servicePort) {
    process.exit(1);
  }
  await app.listen(servicePort);
}
void bootstrap();
