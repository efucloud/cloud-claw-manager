export type ProviderBreakdown = {
  provider?: string;
  requests?: number;
};

export type InstanceTelemetry = {
  namespace?: string;
  name?: string;
  displayName?: string;
  purpose?: string;
  ownerUsername?: string;
  ownerEmail?: string;
  ownerId?: string;
  accessible?: boolean;
  readyPods?: number;
  provider?: string;
  providerCandidates?: string[];
  providerPrimary?: string;
  providerModelIds?: string[];
  providerBaseUrl?: string;
  providerApiType?: string;
  endpoint?: string;
  gatewayToken?: string;
  memoryBytes?: number;
  networkRxBytesPerSecond?: number;
  networkTxBytesPerSecond?: number;
  inputTokens24h?: number;
  outputTokens24h?: number;
  costUSD24h?: number;
  runtimeInsights?: RuntimeInsights;
  updatedAt?: string;
};

export type RuntimeInsights = {
  collectedAt?: string;
  sourcePod?: string;
  sourceContainer?: string;
  statusState?: string;
  gatewayState?: string;
  securityState?: string;
  runtimeVersion?: string;
  openclawConfigPath?: string;
  openclawAgentsCount?: number;
  openclawProvidersCount?: number;
  openclawPrimaryModel?: string;
  openclawWorkspace?: string;
  controlUiDisableDeviceAuth?: boolean;
  controlUiAllowedOrigins?: string[];
  cronJobsCount?: number;
  sessionIndexFilesCount?: number;
  lastError?: string;
};

export type DashboardOverview = {
  totalInstances?: number;
  accessibleInstances?: number;
  readyPods?: number;
  memoryBytes?: number;
  networkRxBytesPerSecond?: number;
  networkTxBytesPerSecond?: number;
  inputTokens24h?: number;
  outputTokens24h?: number;
  costUSD24h?: number;
};

export type DashboardResponse = {
  overview?: DashboardOverview;
  providers?: ProviderBreakdown[];
  instances?: InstanceTelemetry[];
};

export type OpenClawTemplate = {
  name: string;
  namespace: string;
  description?: string;
  schema?: Record<string, any>;
  defaults?: Record<string, any>;
  createdAt?: string;
  updatedAt?: string;
};
