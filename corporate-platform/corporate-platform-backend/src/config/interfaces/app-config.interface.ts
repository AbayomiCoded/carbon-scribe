export interface AppConfig {
  nodeEnv: string;
  port: number;
  apiPrefix: string;
  serviceName: string;
}

export interface RedisConfig {
  url: string;
  timeout: number;
  maxRetries: number;
  retryDelay: number;
  connectionTimeout: number;
  idleTimeout: number;
}

export interface KafkaConfig {
  brokers: string[];
  clientId: string;
  groupId: string;
  producerTimeout: number;
  consumerTimeout: number;
  maxRetries: number;
  retryDelay: number;
  connectionTimeout: number;
}

export interface StellarConfig {
  network: 'public' | 'testnet';
  sorobanRpcUrl: string;
  horizonUrl: string;
  simulateTimeout: number;
  sendTimeout: number;
  getTransactionTimeout: number;
  getEventsTimeout: number;
  getLatestLedgerTimeout: number;
}

export interface TimeoutConfig {
  defaultTimeout: number;
  shutdownGracePeriod: number;
  healthCheckTimeout: number;
}

export interface CircuitBreakerConfig {
  enabled: boolean;
  failureThreshold: number;
  successThreshold: number;
  timeout: number;
  resetTimeout: number;
}