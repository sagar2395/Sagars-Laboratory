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

/** One provider entry from GET /api/platform — the full registry view
 *  (every provider per category, e.g. ingress → traefik AND nginx). */
export interface PlatformProviderEntry {
  name: string
  category: string
  installed: boolean
}

/** GET /api/platform response: category key → providers.
 *  Category keys may be nested ("monitoring/metrics"). */
export type PlatformProvidersMap = Record<string, PlatformProviderEntry[]>

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
  cluster: ClusterInfo | null
  platform: PlatformStatus
  apps: AppInfo[] | null
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
  current?: boolean
}

// ── Async actions / jobs ─────────────────────────────────────────────────────

/** 202 response body for every POST action endpoint. */
export interface ActionAccepted {
  jobId: string
  status: string
}

/** One entry from GET /api/jobs — recorded action history, newest first. */
export interface JobInfo {
  id: string
  action: string
  status: 'running' | 'succeeded' | 'failed'
  error?: string
  startedAt: string
  endedAt?: string
}

// ── WebSocket events ─────────────────────────────────────────────────────────

export type ActionEventType = 'action_start' | 'action_output' | 'action_end' | 'action_error'

export interface ActionEvent {
  id?: string
  type: ActionEventType
  action?: string
  command?: string
  output?: string
  stream?: 'stdout' | 'stderr'
  exitCode?: number
  error?: string
  timestamp?: string
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

export type NotifyFn = (level: NotifLevel, title: string, detail?: string) => void
