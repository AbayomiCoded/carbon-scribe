import { Test, TestingModule } from '@nestjs/testing';
import { INestApplication } from '@nestjs/common';
import * as request from 'supertest';
import { AppModule } from '../app.module';
import * as helmet from 'helmet';

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

    expect(response.headers).toHaveProperty('permissions-policy');
    expect(response.headers['permissions-policy']).toContain('geolocation=');
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
});