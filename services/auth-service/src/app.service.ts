import { Injectable } from '@nestjs/common';

@Injectable()
export class AppService {
  serviceHealth(): { status: string } {
    return { status: 'Auth Service OK' };
  }
}
