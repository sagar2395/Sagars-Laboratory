export interface ClusterInfo {
  provider: string;
  version: string;
  nodes: number;
  ready: boolean;
  context: string;
}

export interface ComponentStatus {
  provider: string;
  active: boolean;
}

export interface PlatformStatusResp {
  ingress: ComponentStatus;
  metrics: ComponentStatus;
  logging: ComponentStatus;
  tracing: ComponentStatus;
}

export interface AppStatusResp {
  name: string;
  buildStrategy: string;
  deployStrategy: string;
  deployed: boolean;
  replicas?: string;
  ready?: string;
}

export interface StatusResponse {
  domainSuffix: string;
  cluster: ClusterInfo;
  platform: PlatformStatusResp;
  apps: AppStatusResp[];
}

export interface ScenarioStatus {
  name: string;
  active: boolean;
  description: string;
  category: string;
}

export interface ActionEvent {
  ID: string;
  Type: string;
  Action: string;
  ExitCode?: number;
  Error?: string;
  Timestamp: string;
  Line?: string; // If streaming output
}

const API_BASE = '/api';

export const apiClient = {
  async getStatus(): Promise<StatusResponse> {
    const res = await fetch(`${API_BASE}/status`);
    if (!res.ok) throw new Error('Failed to fetch status');
    return res.json();
  },

  async getScenarios(): Promise<ScenarioStatus[]> {
    const res = await fetch(`${API_BASE}/scenarios`);
    if (!res.ok) throw new Error('Failed to fetch scenarios');
    return res.json();
  },

  async scenarioUp(name: string): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/scenarios/${name}/up`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to activate scenario: ${name}`);
    return res.json();
  },

  async scenarioDown(name: string): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/scenarios/${name}/down`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to deactivate scenario: ${name}`);
    return res.json();
  },

  async appDeploy(name: string): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/apps/${name}/deploy`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to deploy app: ${name}`);
    return res.json();
  },

  async appDestroy(name: string): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/apps/${name}/destroy`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to destroy app: ${name}`);
    return res.json();
  },
  
  async platformUp(): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/platform/up`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to install platform`);
    return res.json();
  },
  
  async componentUp(category: string, name: string): Promise<{ status: string }> {
    const res = await fetch(`${API_BASE}/platform/${category}/${name}/up`, { method: 'POST' });
    if (!res.ok) throw new Error(`Failed to start component`);
    return res.json();
  }
};
