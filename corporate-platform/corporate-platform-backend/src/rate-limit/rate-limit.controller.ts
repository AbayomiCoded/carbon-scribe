import { Controller, Get, Param, UseGuards } from '@nestjs/common';
import { RateLimitService } from './rate-limit.service';
import { JwtAuthGuard } from '../auth/guards/jwt-auth.guard';
import { CurrentUser } from '../auth/decorators/current-user.decorator';
import { JwtPayload } from '../auth/interfaces/jwt-payload.interface';

@Controller('api/v1/rate-limit')
@UseGuards(JwtAuthGuard)
export class RateLimitController {
  constructor(private readonly rateLimitService: RateLimitService) {}

  /**
   * Get rate limit metrics for an endpoint
   */
  @Get('metrics/:endpoint')
  async getMetrics(
    @CurrentUser() user: JwtPayload,
    @Param('endpoint') endpoint: string,
  ) {
    // Only allow admins to view metrics
    if (user.role !== 'admin') {
      return { error: 'Admin access required' };
    }

    return this.rateLimitService.getMetrics(endpoint);
  }

  /**
   * Reset rate limit for a key
   */
  @Get('reset/:keyPrefix/:key')
  async resetRateLimit(
    @CurrentUser() user: JwtPayload,
    @Param('keyPrefix') keyPrefix: string,
    @Param('key') key: string,
  ) {
    // Only allow admins to reset
    if (user.role !== 'admin') {
      return { error: 'Admin access required' };
    }

    await this.rateLimitService.resetRateLimit(key, keyPrefix);
    return { success: true };
  }
}