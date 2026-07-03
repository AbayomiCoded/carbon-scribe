import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../app.module';
import helmet from 'helmet';

describe('Security Headers (e2e)', () => {
  let app: INestApplication;

  beforeAll(async () => {
    const moduleFixture: TestingModule = await Test.createTestingModule({
      imports: [AppModule],
    }).compile();

    app = moduleFixture.createNestApplication();
    app.use(helmet());
    await app.init();
  });

  afterAll(async () => {
    await app.close();
  });

  it('should return security headers', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    // Verify security headers are present
    expect(response.headers).toHaveProperty('x-frame-options');
    expect(response.headers['x-frame-options']).toBe('DENY');

    expect(response.headers).toHaveProperty('x-content-type-options');
    expect(response.headers['x-content-type-options']).toBe('nosniff');

    expect(response.headers).toHaveProperty('x-xss-protection');
    expect(response.headers['x-xss-protection']).toBe('1; mode=block');

    expect(response.headers).toHaveProperty('strict-transport-security');
    expect(response.headers['strict-transport-security']).toContain('max-age=');

    expect(response.headers).toHaveProperty('referrer-policy');
    expect(response.headers['referrer-policy']).toBe('strict-origin-when-cross-origin');

    expect(response.headers).toHaveProperty('x-permitted-cross-domain-policies');
    expect(response.headers['x-permitted-cross-domain-policies']).toBe('none');

    // Permissions-Policy header check - using the manual header set in main.ts
    expect(response.headers).toHaveProperty('permissions-policy');
    expect(response.headers['permissions-policy']).toContain('geolocation=(self)');
  });

  it('should have CSP header when enabled', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    // CSP header should be present
    expect(response.headers).toHaveProperty('content-security-policy');
    const csp = response.headers['content-security-policy'];
    expect(csp).toContain("default-src 'self'");
    expect(csp).toContain("script-src");
  });

  it('should not expose X-Powered-By header', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).not.toHaveProperty('x-powered-by');
  });

  it('should have HSTS header in production-like environments', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('strict-transport-security');
    const hsts = response.headers['strict-transport-security'];
    expect(hsts).toContain('max-age=');
    expect(hsts).toContain('includeSubDomains');
  });

  it('should have Cross-Origin-Resource-Policy header', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('cross-origin-resource-policy');
    expect(response.headers['cross-origin-resource-policy']).toBe('same-origin');
  });

  it('should have Cross-Origin-Opener-Policy header', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('cross-origin-opener-policy');
    expect(response.headers['cross-origin-opener-policy']).toBe('same-origin');
  });

  it('should have X-DNS-Prefetch-Control header', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('x-dns-prefetch-control');
    expect(response.headers['x-dns-prefetch-control']).toBe('off');
  });

  it('should have Origin-Agent-Cluster header', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('origin-agent-cluster');
    expect(response.headers['origin-agent-cluster']).toBe('?1');
  });

  it('should have Cache-Control headers for API responses', async () => {
    const response = await request(app.getHttpServer())
      .get('/health')
      .expect(200);

    expect(response.headers).toHaveProperty('cache-control');
    expect(response.headers['cache-control']).toContain('no-store');
    expect(response.headers['cache-control']).toContain('no-cache');
    expect(response.headers['cache-control']).toContain('must-revalidate');
  });
});