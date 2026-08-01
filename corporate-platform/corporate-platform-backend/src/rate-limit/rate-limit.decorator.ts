import { SetMetadata } from '@nestjs/common';
import { RateLimitDecoratorOptions } from './rate-limit.types';

export const RATE_LIMIT_KEY = 'rate_limit';

/**
 * Decorator to apply rate limiting to an endpoint
 *
 * @param options - Rate limit configuration options
 * @example
 * @RateLimit({ max: 5, windowMs: 60000, keyPrefix: 'place-bid' })
 * async placeBid() { ... }
 */
export function RateLimit(options: RateLimitDecoratorOptions) {
  return SetMetadata(RATE_LIMIT_KEY, options);
}

/**
 * Predefined rate limits for common use cases
 */
export const RateLimits = {
  /**
   * Bidding: 5 bids per minute per user per auction
   */
  BIDDING: {
    max: 5,
    windowMs: 60000,
    keyPrefix: 'bidding',
    enableGraduatedCooldown: true,
    message: 'Too many bid attempts. Please wait before trying again.',
  },

  /**
   * Global auction bidding: 20 bids per minute per auction
   */
  GLOBAL_AUCTION_BIDDING: {
    max: 20,
    windowMs: 60000,
    keyPrefix: 'global-auction-bidding',
    enableGraduatedCooldown: true,
    message: 'Too many bids on this auction. Please wait before trying again.',
  },

  /**
   * Retirement: 3 retirements per minute per user
   */
  RETIREMENT: {
    max: 3,
    windowMs: 60000,
    keyPrefix: 'retirement',
    enableGraduatedCooldown: true,
    message: 'Too many retirement requests. Please wait before trying again.',
  },

  /**
   * Company retirement: 10 retirements per minute per company
   */
  COMPANY_RETIREMENT: {
    max: 10,
    windowMs: 60000,
    keyPrefix: 'company-retirement',
    enableGraduatedCooldown: true,
    message: 'Too many company retirements. Please wait before trying again.',
  },

  /**
   * Similar credits: 30 requests per minute per IP
   */
  SIMILAR_CREDITS: {
    max: 30,
    windowMs: 60000,
    keyPrefix: 'similar-credits',
    message:
      'Too many similar credits requests. Please wait before trying again.',
  },

  /**
   * Search: 20 requests per minute per IP
   */
  SEARCH: {
    max: 20,
    windowMs: 60000,
    keyPrefix: 'search',
    message: 'Too many search requests. Please wait before trying again.',
  },
};
