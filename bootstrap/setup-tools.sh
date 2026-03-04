#!/bin/bash

set -e

# Load versions from versions.env
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION_FILE="${SCRIPT_DIR}/../versions.env"

if [ ! -f "$VERSION_FILE" ]; then
    echo "Error: versions.env file not found at $VERSION_FILE"
    exit 1
fi

source "$VERSION_FILE"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Common functions for all profiles
install_kubectl() {
    echo -e "${YELLOW}Installing kubectl (v${KUBECTL_VERSION})...${NC}"
    
    if command -v kubectl &> /dev/null; then
        current_version=$(kubectl version --client 2>/dev/null | grep "Client Version:" | awk '{print $NF}' | sed 's/v//')
        if [ "$current_version" = "${KUBECTL_VERSION}" ]; then
            echo -e "${GREEN}kubectl v${KUBECTL_VERSION} is already installed${NC}"
            return 0
        else
            echo -e "${YELLOW}kubectl is installed but version mismatch (current: v${current_version}, wanted: v${KUBECTL_VERSION})${NC}"
        fi
    fi
    
    echo "Downloading kubectl v${KUBECTL_VERSION}..."
    curl -LO "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
    
    # Verify installation
    installed_version=$(kubectl version --client 2>/dev/null | grep "Client Version:" | awk '{print $NF}' | sed 's/v//')
    if [ "$installed_version" = "${KUBECTL_VERSION}" ]; then
        echo -e "${GREEN}kubectl v${KUBECTL_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: kubectl version mismatch after install (got v${installed_version}, wanted v${KUBECTL_VERSION})${NC}"
        exit 1
    fi
}

# k3d-specific functions
install_docker() {
    echo -e "${YELLOW}Installing Docker (v${DOCKER_VERSION})...${NC}"
    
    if command -v docker &> /dev/null; then
        current_version=$(docker --version 2>/dev/null | grep -oP 'version \K[^,]*')
        if [ "$current_version" = "${DOCKER_VERSION}" ]; then
            echo -e "${GREEN}Docker v${DOCKER_VERSION} is already installed${NC}"
            return 0
        else
            echo -e "${YELLOW}Docker is installed but version mismatch (current: v${current_version}, wanted: v${DOCKER_VERSION})${NC}"
        fi
    fi
    
    echo -e "${YELLOW}WARNING: Docker installation requires manual setup for version pinning.${NC}"
    echo -e "${YELLOW}Please install Docker v${DOCKER_VERSION} manually from: https://docs.docker.com/engine/install/${NC}"
    echo -e "${YELLOW}Or uncomment automated installation below (will install latest)${NC}"
    # curl -fsSL https://get.docker.com | sh
    
    echo -e "${RED}Docker installation skipped.${NC}"
}

install_k3d() {
    echo -e "${YELLOW}Installing k3d (v${K3D_VERSION})...${NC}"
    
    if command -v k3d &> /dev/null; then
        current_version=$(k3d version 2>/dev/null | grep -oP 'k3d version v\K[^-]*')
        if [ "$current_version" = "${K3D_VERSION}" ]; then
            echo -e "${GREEN}k3d v${K3D_VERSION} is already installed${NC}"
            return 0
        else
            echo -e "${YELLOW}k3d is installed but version mismatch (current: v${current_version}, wanted: v${K3D_VERSION})${NC}"
        fi
    fi
    
    echo "Downloading k3d v${K3D_VERSION}..."
    curl -Lo /tmp/k3d "https://github.com/k3d-io/k3d/releases/download/v${K3D_VERSION}/k3d-linux-amd64"
    chmod +x /tmp/k3d
    sudo mv /tmp/k3d /usr/local/bin/
    
    # Verify installation
    installed_version=$(k3d version 2>/dev/null | grep -oP 'k3d version v\K[^-]*')
    if [ "$installed_version" = "${K3D_VERSION}" ]; then
        echo -e "${GREEN}k3d v${K3D_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: k3d version mismatch after install (got v${installed_version}, wanted v${K3D_VERSION})${NC}"
        exit 1
    fi
}

install_helm() {
    echo -e "${YELLOW}Installing Helm (v${HELM_VERSION})...${NC}"
    
    if command -v helm &> /dev/null; then
        current_version=$(helm version --short 2>/dev/null | grep -oP 'v\K[^+]*')
        if [ "$current_version" = "${HELM_VERSION}" ]; then
            echo -e "${GREEN}Helm v${HELM_VERSION} is already installed${NC}"
            return 0
        else
            echo -e "${YELLOW}Helm is installed but version mismatch (current: v${current_version}, wanted: v${HELM_VERSION})${NC}"
        fi
    fi
    
    echo "Downloading Helm v${HELM_VERSION}..."
    cd /tmp
    curl -Lo helm.tar.gz "https://get.helm.sh/helm-v${HELM_VERSION}-linux-amd64.tar.gz"
    tar -xzf helm.tar.gz
    sudo mv linux-amd64/helm /usr/local/bin/
    rm -rf helm.tar.gz linux-amd64
    cd - > /dev/null
    
    # Verify installation
    installed_version=$(helm version --short 2>/dev/null | grep -oP 'v\K[^+]*')
    if [ "$installed_version" = "${HELM_VERSION}" ]; then
        echo -e "${GREEN}Helm v${HELM_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: Helm version mismatch after install (got v${installed_version}, wanted v${HELM_VERSION})${NC}"
        exit 1
    fi
}

# AKS-specific functions
install_az_cli() {
    echo -e "${YELLOW}Installing Azure CLI (v${AZ_CLI_VERSION})...${NC}"
    
    if command -v az &> /dev/null; then
        current_version=$(az --version 2>/dev/null | head -1 | awk '{print $NF}')
        if [ "$current_version" = "${AZ_CLI_VERSION}" ]; then
            echo -e "${GREEN}Azure CLI v${AZ_CLI_VERSION} is already installed${NC}"
            return 0
        else
            echo -e "${YELLOW}Azure CLI is installed but version mismatch (current: v${current_version}, wanted: v${AZ_CLI_VERSION})${NC}"
        fi
    fi
    
    echo "Installing Azure CLI v${AZ_CLI_VERSION}..."
    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    
    # Verify installation
    installed_version=$(az --version 2>/dev/null | head -1 | awk '{print $NF}')
    if [ "$installed_version" = "${AZ_CLI_VERSION}" ]; then
        echo -e "${GREEN}Azure CLI v${AZ_CLI_VERSION} installed and verified${NC}"
    else
        echo -e "${YELLOW}WARNING: Azure CLI version mismatch (got v${installed_version}, wanted v${AZ_CLI_VERSION})${NC}"
        echo -e "${YELLOW}This may be due to the installer only providing the latest stable version.${NC}"
    fi
}

# Profile-specific installation functions

install_common() {
    echo -e "${GREEN}========== Installing Common Tools ==========${NC}"
    install_kubectl
    echo -e "${GREEN}========== Common Tools Complete ==========${NC}\n"
}

install_k3d_profile() {
    echo -e "${GREEN}========== Installing K3D Profile ==========${NC}"
    install_common
    install_docker
    install_k3d
    install_helm
    echo -e "${GREEN}Validating cluster details...${NC}"
    kubectl cluster-info || echo -e "${YELLOW}Cluster info not available - run 'make runtime-up' first${NC}"
    echo -e "${GREEN}========== K3D Profile Complete ==========${NC}\n"
}

install_aks_profile() {
    echo -e "${GREEN}========== Installing AKS Profile ==========${NC}"
    install_common
    install_az_cli
    echo -e "${GREEN}========== AKS Profile Complete ==========${NC}\n"
}

# Help function
show_help() {
    cat << EOF
Usage: setup-tools.sh [PROFILE]

Available profiles:
  k3d       - Install tools for local k3d cluster setup (kubectl, docker, k3d)
  aks       - Install tools for Azure AKS cluster setup (kubectl, azure-cli)
  common    - Install common tools only (kubectl)
  all       - Install all tools for both profiles

Examples:
  ./setup-tools.sh k3d
  ./setup-tools.sh aks
  ./setup-tools.sh common
  ./setup-tools.sh all

If no profile is specified, defaults to 'k3d'.
EOF
}

# Main function
main() {
    local profile="${1:-k3d}"
    
    echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║     Setup Tools - Profile: ${profile}${NC}${GREEN}             ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}\n"
    
    case "$profile" in
        k3d)
            install_k3d_profile
            ;;
        aks)
            install_aks_profile
            ;;
        common)
            install_common
            ;;
        all)
            install_k3d_profile
            install_aks_profile
            ;;
        help|--help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown profile '$profile'${NC}"
            show_help
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Setup complete!${NC}"
}

# Run main function with provided arguments
main "$@"

