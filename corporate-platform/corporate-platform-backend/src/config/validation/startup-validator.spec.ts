import { Test, TestingModule } from '@nestjs/testing';
import { StartupValidator } from './startup-validator';
import { ConfigService } from '../config.service';
import { ServiceValidator } from './service-validator';

describe('StartupValidator', () => {
  let startupValidator: StartupValidator;
  let configService: ConfigService;
  let serviceValidator: ServiceValidator;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        StartupValidator,
        {
          provide: ConfigService,
          useValue: {
            getAppConfig: jest.fn().mockReturnValue({
              nodeEnv: 'production',
              port: 3000,
            }),
            getDatabaseConfig: jest.fn().mockReturnValue({
              url: 'postgresql://user:pass@localhost:5432/db',
            }),
            getRedisConfig: jest.fn().mockReturnValue({
              host: 'localhost',
              port: 6379,
            }),
            getKafkaConfig: jest.fn().mockReturnValue({
              brokers: ['localhost:9092'],
            }),
            getStellarConfig: jest.fn().mockReturnValue({
              network: 'mainnet',
              horizonUrl: 'https://horizon.stellar.org',
              sorobanRpcUrl: 'https://soroban.stellar.org',
            }),
            getAuthConfig: jest.fn().mockReturnValue({
              jwtSecret: 'secure-jwt-secret-with-32-chars-min!',
            }),
            getServicesConfig: jest.fn().mockReturnValue({
              ipfsGateway: 'https://gateway.pinata.cloud/ipfs/',
            }),
          },
        },
        {
          provide: ServiceValidator,
          useValue: {
            validateAllServices: jest.fn().mockResolvedValue([
              { service: 'Database', connected: true, latencyMs: 10 },
              { service: 'Redis', connected: true, latencyMs: 5 },
              { service: 'Kafka', connected: true, latencyMs: 20 },
              { service: 'Stellar', connected: true, latencyMs: 50 },
              { service: 'IPFS', connected: true, latencyMs: 30 },
            ]),
          },
        },
      ],
    }).compile();

    startupValidator = module.get<StartupValidator>(StartupValidator);
    configService = module.get<ConfigService>(ConfigService);
    serviceValidator = module.get<ServiceValidator>(ServiceValidator);
  });

  it('should be defined', () => {
    expect(startupValidator).toBeDefined();
  });

  describe('validateStartup', () => {
    it('should return valid result when all checks pass', async () => {
      const result = await startupValidator.validateStartup();
      expect(result.valid).toBe(true);
      expect(result.errors).toHaveLength(0);
      expect(result.connectivity).toHaveLength(5);
    });

    it('should return errors when validation fails', async () => {
      jest.spyOn(configService, 'getDatabaseConfig').mockReturnValue({
        url: '',
        poolSize: 10,
      });

      const result = await startupValidator.validateStartup();
      expect(result.valid).toBe(false);
      expect(result.errors.length).toBeGreaterThan(0);
    });

    it('should include connectivity results', async () => {
      const result = await startupValidator.validateStartup();
      expect(result.connectivity).toBeDefined();
      expect(result.connectivity.length).toBeGreaterThan(0);
    });
  });
});