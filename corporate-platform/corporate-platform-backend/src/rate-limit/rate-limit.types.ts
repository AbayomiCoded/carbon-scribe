export interface RateLimitConfig {
  windowMs: number;
  maxRequests: number;
  keyPrefix: string;
  message?: string;
  statusCode?: number;
  skipOnError?: boolean;
  enableGraduatedCooldown?: boolean;
}

export interface RateLimitResult {
  allowed: boolean;
  current: number;
  max: number;
  resetTime: number;
  retryAfter?: number;
  windowMs: number;
}

export interface RateLimitViolation {
  endpoint: string;
  userId?: string;
  companyId?: string;
  ip: string;
  key: string;
  current: number;
  max: number;
  resetTime: number;
  timestamp: Date;
}

export interface RateLimitMetrics {
  totalRequests: number;
  allowedRequests: number;
  blockedRequests: number;
  violations: number;
  byEndpoint: Record<
    string,
    {
      requests: number;
      blocked: number;
    }
  >;
}

export interface RateLimitDecoratorOptions {
  windowMs?: number;
  max?: number;
  keyPrefix?: string;
  skipOnError?: boolean;
  enableGraduatedCooldown?: boolean;
  message?: string;
  statusCode?: number;
}
