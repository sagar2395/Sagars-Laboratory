// ── Cluster & Status ─────────────────────────────────────────────────────────

export interface ClusterInfo {
  context: string
  server: string
  k8sVersion: string
  nodeCount: number
  connected: boolean
}

export interface PlatformComponent {
  provider: string
  active: boolean
}

export interface PlatformStatus {
  ingress?: PlatformComponent
  metrics?: PlatformComponent
  logging?: PlatformComponent
  tracing?: PlatformComponent
  gitops?: PlatformComponent
  chaos?: PlatformComponent
  policy?: PlatformComponent
  secrets?: PlatformComponent
  [key: string]: PlatformComponent | undefined
}

export interface AppInfo {
  name: string
  buildStrategy: string
  deployStrategy: string
  deployed: boolean
  replicas?: string
  ready?: string
  namespace?: string
}

export interface StatusResponse {
  cluster: ClusterInfo
  platform: PlatformStatus
  apps: AppInfo[]
  domainSuffix: string
}

// ── Scenarios ────────────────────────────────────────────────────────────────

export interface ScenarioPrerequisites {
  platform?: string[]
  apps?: string[]
}

export interface ExploreURL {
  label: string
  url: string
}

export interface ExploreCommand {
  label: string
  command: string
}

export interface Explore {
  urls?: ExploreURL[]
  commands?: ExploreCommand[]
  tips?: string[]
}

export interface ScenarioComponent {
  name: string
  type: string
  chart?: string
  namespace?: string
  path?: string
  script?: string
  version?: string
}

export interface Scenario {
  name: string
  displayName: string
  description: string
  category: string
  active: boolean
  runtimes?: string[]
  prerequisites?: ScenarioPrerequisites
  components?: ScenarioComponent[]
  explore?: Explore
}

// ── Dashboards & Runtimes ────────────────────────────────────────────────────

export interface DashboardURL {
  name: string
  label: string
  url: string
  available: boolean
  category: string
}

export interface Runtime {
  name?: string
  profile?: string
  active?: boolean
}

// ── WebSocket events ─────────────────────────────────────────────────────────

export type ActionEventType = 'action_start' | 'action_output' | 'action_end' | 'action_error'

export interface ActionEvent {
  type: ActionEventType
  action?: string
  command?: string
  output?: string
  stream?: 'stdout' | 'stderr'
  exitCode?: number
  error?: string
}

export interface WSMessage {
  type: 'status' | 'action'
  data: unknown
}

// ── Log entries ───────────────────────────────────────────────────────────────

export type LogLevel = 'cmd' | 'output' | 'stderr' | 'success' | 'error'

export interface LogEntry {
  id: number
  ts: string
  level: LogLevel
  text: string
}

// ── Notifications ─────────────────────────────────────────────────────────────

export type NotifLevel = 'info' | 'success' | 'error'

export interface Notification {
  id: number
  level: NotifLevel
  title: string
  detail?: string
}
