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
    echo -e "${YELLOW}Installing kubectl (${KUBECTL_VERSION})...${NC}"
    
    if command -v kubectl &> /dev/null; then
        echo -e "${GREEN}kubectl is already installed${NC}"
        kubectl version --client
        return 0
    fi
    
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
    kubectl version --client
    echo -e "${GREEN}kubectl installed successfully${NC}"
}

# k3d-specific functions
install_docker() {
    echo -e "${YELLOW}Installing Docker...${NC}"
    
    if command -v docker &> /dev/null; then
        echo -e "${GREEN}Docker is already installed${NC}"
        docker --version
        return 0
    fi
    
    echo -e "${YELLOW}Please install Docker manually or uncomment the automated installation${NC}"
    # Uncomment the line below to enable automated Docker installation
    # curl -fsSL https://get.docker.com | sh
    
    echo -e "${RED}Docker installation skipped. Please install manually.${NC}"
}

install_k3d() {
    echo -e "${YELLOW}Installing k3d (${K3D_VERSION})...${NC}"
    
    if command -v k3d &> /dev/null; then
        echo -e "${GREEN}k3d is already installed${NC}"
        k3d version
        return 0
    fi
    
    curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
    echo -e "${GREEN}k3d installed successfully${NC}"
}

install_helm() {
    echo -e "${YELLOW}Installing Helm (${HELM_VERSION})...${NC}"
    
    if command -v helm &> /dev/null; then
        echo -e "${GREEN}Helm is already installed${NC}"
        helm version
        return 0
    fi
    
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
    echo -e "${GREEN}Helm installed successfully${NC}"
}

# AKS-specific functions
install_az_cli() {
    echo -e "${YELLOW}Installing Azure CLI (${AZ_CLI_VERSION})...${NC}"
    
    if command -v az &> /dev/null; then
        echo -e "${GREEN}Azure CLI is already installed${NC}"
        az --version
        return 0
    fi
    
    curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    echo -e "${GREEN}Azure CLI installed successfully${NC}"
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
    kubectl cluster-info || echo -e "${YELLOW}Cluster info not available - run 'make cluster-up' first${NC}"
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

