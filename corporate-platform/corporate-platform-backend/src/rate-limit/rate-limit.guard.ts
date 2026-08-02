import {
  CanActivate,
  ExecutionContext,
  Injectable,
  Logger,
  HttpException,
  HttpStatus,
} from '@nestjs/common';
import { Reflector } from '@nestjs/core';
import { Request, Response } from 'express';
import { RateLimitService } from './rate-limit.service';
import { RATE_LIMIT_KEY } from './rate-limit.decorator';
import { RateLimitDecoratorOptions } from './rate-limit.types';
import { SecurityService } from '../security/security.service';
import { SecurityEvents } from '../security/constants/security-events.constants';

@Injectable()
export class RateLimitGuard implements CanActivate {
  private readonly logger = new Logger(RateLimitGuard.name);

  constructor(
    private readonly reflector: Reflector,
    private readonly rateLimitService: RateLimitService,
    private readonly securityService: SecurityService,
  ) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const request = context.switchToHttp().getRequest<Request>();
    const response = context.switchToHttp().getResponse<Response>();

    // Get rate limit configuration from decorator
    const config = this.reflector.get<RateLimitDecoratorOptions>(
      RATE_LIMIT_KEY,
      context.getHandler(),
    );

    // If no config, allow the request
    if (!config) {
      return true;
    }

    // Build rate limit key
    const key = this.buildKey(request, config);

    // Check if IP is whitelisted
    const isWhitelisted = await this.isIpWhitelisted(request);
    if (isWhitelisted) {
      return true;
    }

    // Get default config
    const defaultConfig = {
      windowMs: config.windowMs || 60000,
      maxRequests: config.max || 10,
      keyPrefix: config.keyPrefix || 'default',
      skipOnError: config.skipOnError || false,
      enableGraduatedCooldown: config.enableGraduatedCooldown || false,
      message: config.message || 'Too many requests, please try again later.',
      statusCode: config.statusCode || HttpStatus.TOO_MANY_REQUESTS,
    };

    try {
      // Check rate limit
      const result = await this.rateLimitService.checkRateLimit(
        key,
        defaultConfig,
      );

      // Set rate limit headers
      this.setRateLimitHeaders(response, result);

      // If not allowed, handle violation
      if (!result.allowed) {
        await this.handleViolation(request, key, config, result);
        throw new HttpException(
          {
            statusCode: defaultConfig.statusCode,
            message: defaultConfig.message,
            retryAfter: result.retryAfter,
          },
          defaultConfig.statusCode,
        );
      }

      // Track successful request
      await this.trackRequest(request, config);

      return true;
    } catch (error) {
      // If it's an HttpException, rethrow it
      if (error instanceof HttpException) {
        throw error;
      }

      // Log error and allow or block based on configuration
      const err = error as Error;
      this.logger.error(`Rate limit guard error: ${err.message}`);

      if (defaultConfig.skipOnError) {
        return true;
      }

      throw error;
    }
  }

  /**
   * Build rate limit key from request and config
   */
  private buildKey(
    request: Request,
    config: RateLimitDecoratorOptions,
  ): string {
    const parts: string[] = [];

    // Add user ID if authenticated
    const user = (request as any).user;
    if (user?.sub) {
      parts.push(`user:${user.sub}`);
    }

    // Add company ID if available
    if (user?.companyId) {
      parts.push(`company:${user.companyId}`);
    }

    // Add IP address
    const ip = this.rateLimitService.getClientIp(
      request.headers,
      request.connection,
    );
    parts.push(`ip:${ip}`);

    // Add endpoint-specific parts
    if (config.keyPrefix) {
      parts.push(`endpoint:${config.keyPrefix}`);
    }

    // Add auction ID for bidding endpoints
    if (request.params?.id) {
      parts.push(`auction:${request.params.id}`);
    }

    return parts.join(':');
  }

  /**
   * Check if IP is whitelisted
   */
  private async isIpWhitelisted(request: Request): Promise<boolean> {
    try {
      const user = (request as any).user;
      const companyId = user?.companyId || null;
      const ip = this.rateLimitService.getClientIp(
        request.headers,
        request.connection,
      );

      return await this.securityService.isIpAllowed(companyId, ip);
    } catch {
      return false;
    }
  }

  /**
   * Set rate limit headers on response
   */
  private setRateLimitHeaders(response: Response, result: any): void {
    response.setHeader('X-RateLimit-Limit', result.max);
    response.setHeader(
      'X-RateLimit-Remaining',
      Math.max(0, result.max - result.current),
    );
    response.setHeader('X-RateLimit-Reset', Math.ceil(result.resetTime / 1000));

    if (result.retryAfter) {
      response.setHeader('Retry-After', result.retryAfter);
    }
  }

  /**
   * Handle rate limit violation
   */
  private async handleViolation(
    request: Request,
    key: string,
    config: RateLimitDecoratorOptions,
    result: any,
  ): Promise<void> {
    const user = (request as any).user;
    const ip = this.rateLimitService.getClientIp(
      request.headers,
      request.connection,
    );

    // Log violation
    await this.rateLimitService.logViolation({
      endpoint: request.path || request.url,
      userId: user?.sub,
      companyId: user?.companyId,
      ip,
      key,
      current: result.current,
      max: result.max,
      resetTime: result.resetTime,
      timestamp: new Date(),
    });

    // Track blocked request
    await this.trackBlockedRequest(request, config);

    // Log security event
    if (user?.sub) {
      await this.securityService.logEvent({
        eventType: SecurityEvents.RateLimitExceeded,
        userId: user.sub,
        companyId: user.companyId,
        ipAddress: ip,
        userAgent: request.headers['user-agent'] as string,
        resource: request.path || request.url,
        method: request.method,
        status: 'blocked',
        statusCode: 429,
        details: {
          rateLimit: {
            key,
            current: result.current,
            max: result.max,
            resetTime: new Date(result.resetTime).toISOString(),
          },
        },
      });
    }
  }

  /**
   * Track successful request
   */
  private async trackRequest(
    request: Request,
    config: RateLimitDecoratorOptions,
  ): Promise<void> {
    const client = this.rateLimitService['redisService'].getClient();
    const endpoint = config.keyPrefix || 'default';
    const today = new Date().toISOString().slice(0, 10);

    try {
      await client.incr(`rate-limit:requests:${endpoint}:${today}`);
      await client.expire(`rate-limit:requests:${endpoint}:${today}`, 86400);
    } catch {
      // Ignore tracking errors
    }
  }

  /**
   * Track blocked request
   */
  private async trackBlockedRequest(
    request: Request,
    config: RateLimitDecoratorOptions,
  ): Promise<void> {
    const client = this.rateLimitService['redisService'].getClient();
    const endpoint = config.keyPrefix || 'default';
    const today = new Date().toISOString().slice(0, 10);

    try {
      await client.incr(`rate-limit:blocked:${endpoint}:${today}`);
      await client.expire(`rate-limit:blocked:${endpoint}:${today}`, 86400);
    } catch {
      // Ignore tracking errors
    }
  }
}
