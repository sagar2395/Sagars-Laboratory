# Future Improvements

Potential enhancements for Sagars-Laboratory, roughly ordered by impact.

## 1. Multi-tenancy / RBAC
Add support for multiple users with role-based access to the dashboard. Allow read-only viewers, operators who can deploy apps, and admins who can modify platform components.

## 2. GitOps-driven Configuration
Store all dashboard and platform state in Git. Reconcile cluster state automatically via ArgoCD so that manual UI actions are reflected as commits and drift is detected.

## 3. Persistent Action History
Record every command execution (build, deploy, install, scenario activate) in a lightweight store (SQLite or BoltDB). Provide a searchable history view in the UI with timing, exit codes, and output logs.

## 4. Service Mesh Integration
Add Istio or Linkerd as a platform component. Include traffic visualization, mTLS status, and canary deployment controls in the dashboard.

## 5. Cost Tracking
For cloud runtimes (AKS, EKS), fetch and display estimated costs. Track resource usage over time and show cost-per-namespace or cost-per-app breakdowns.

## 6. Backup & Restore
Implement automated cluster state backup using Velero or a custom solution. Provide point-in-time restore capabilities and scheduled backup policies configurable from the UI.

## 7. Multi-cluster Simultaneous View
Allow viewing and comparing multiple clusters side-by-side. Useful for comparing k3d local development state against a staging AKS/EKS cluster.

## 8. Integration Tests
Build end-to-end tests that spin up a k3d cluster, deploy the full platform stack, deploy applications, activate scenarios, and validate the entire flow programmatically.

## 9. Helm Chart for labctl
Package the labctl API server and UI as a Helm chart that can be deployed into any Kubernetes cluster, rather than running as a local binary.

## 10. External Notifications
Send notifications to Slack, Discord, or email when long-running operations complete or fail. Configurable per-action notification rules.

## 11. Log Aggregation in UI
Integrate Loki or a similar log aggregation system and provide a log viewer directly in the dashboard, allowing real-time log tailing per application or namespace.

## 12. Plugin Architecture
Allow third-party platform components and scenarios to be added as plugins. Define a standard interface for discovery, installation, and lifecycle management.

## 13. Mobile-responsive Dashboard
Optimize the web UI for tablet and phone form factors. Prioritize status views and quick actions for mobile use cases.
