package k8s

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ClusterInfo holds basic cluster information.
type ClusterInfo struct {
	Context    string `json:"context"`
	Server     string `json:"server"`
	K8sVersion string `json:"k8sVersion"`
	NodeCount  int    `json:"nodeCount"`
	Connected  bool   `json:"connected"`
}

// PodInfo holds information about a pod.
type PodInfo struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Restarts  string `json:"restarts"`
	Age       string `json:"age"`
}

// AppStatus holds the deployment status of an application.
type AppStatus struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Replicas  string    `json:"replicas"`
	Ready     string    `json:"ready"`
	Available string    `json:"available"`
	Pods      []PodInfo `json:"pods"`
	Deployed  bool      `json:"deployed"`
}

// GetClusterInfo returns current cluster information.
func GetClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	info := &ClusterInfo{}

	// Get current context
	ctxOut, err := kubectl(ctx, "config", "current-context")
	if err != nil {
		return info, nil // not connected
	}
	info.Context = ctxOut
	info.Connected = true

	// Get server URL
	serverOut, err := kubectl(ctx, "config", "view", "--minify", "-o", "jsonpath={.clusters[0].cluster.server}")
	if err == nil {
		info.Server = serverOut
	}

	// Get k8s version
	versionOut, err := kubectl(ctx, "version", "--short", "-o", "json")
	if err == nil {
		// Simple parse: look for serverVersion
		if idx := strings.Index(versionOut, "gitVersion"); idx >= 0 {
			sub := versionOut[idx:]
			if start := strings.Index(sub, "\"v"); start >= 0 {
				if end := strings.Index(sub[start+1:], "\""); end >= 0 {
					info.K8sVersion = sub[start+1 : start+1+end]
				}
			}
		}
	}

	// Get node count
	nodesOut, err := kubectl(ctx, "get", "nodes", "--no-headers")
	if err == nil && nodesOut != "" {
		info.NodeCount = len(strings.Split(strings.TrimSpace(nodesOut), "\n"))
	}

	return info, nil
}

// GetNamespacePods returns pods in a namespace.
func GetNamespacePods(ctx context.Context, namespace string) ([]PodInfo, error) {
	out, err := kubectl(ctx, "get", "pods", "-n", namespace, "--no-headers",
		"-o", "custom-columns=NAME:.metadata.name,STATUS:.status.phase,READY:.status.conditions[?(@.type=='Ready')].status,RESTARTS:.status.containerStatuses[0].restartCount,AGE:.metadata.creationTimestamp")
	if err != nil {
		return nil, err
	}

	var pods []PodInfo
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		pod := PodInfo{
			Name:      fields[0],
			Namespace: namespace,
			Status:    fields[1],
		}
		if len(fields) > 2 {
			pod.Ready = fields[2]
		}
		if len(fields) > 3 {
			pod.Restarts = fields[3]
		}
		if len(fields) > 4 {
			pod.Age = fields[4]
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

// GetAppStatus returns the deployment status of an app.
func GetAppStatus(ctx context.Context, appName, namespace string) (*AppStatus, error) {
	status := &AppStatus{
		Name:      appName,
		Namespace: namespace,
	}

	// Check if the namespace exists
	_, err := kubectl(ctx, "get", "namespace", namespace, "--no-headers")
	if err != nil {
		return status, nil // Not deployed
	}

	// Get deployment info
	deplOut, err := kubectl(ctx, "get", "deployment", "-n", namespace, "--no-headers",
		"-o", "custom-columns=NAME:.metadata.name,REPLICAS:.spec.replicas,READY:.status.readyReplicas,AVAILABLE:.status.availableReplicas")
	if err == nil && deplOut != "" {
		status.Deployed = true
		fields := strings.Fields(deplOut)
		if len(fields) >= 2 {
			status.Replicas = fields[1]
		}
		if len(fields) >= 3 {
			status.Ready = fields[2]
		}
		if len(fields) >= 4 {
			status.Available = fields[3]
		}
	}

	// Get pods
	pods, err := GetNamespacePods(ctx, namespace)
	if err == nil {
		status.Pods = pods
	}

	return status, nil
}

// NamespaceExists checks if a namespace exists.
func NamespaceExists(ctx context.Context, namespace string) bool {
	_, err := kubectl(ctx, "get", "namespace", namespace, "--no-headers")
	return err == nil
}

func kubectl(ctx context.Context, args ...string) (string, error) {
	path, err := exec.LookPath("kubectl")
	if err != nil {
		return "", fmt.Errorf("kubectl not found in PATH")
	}

	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
