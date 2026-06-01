#!/usr/bin/env bash
set -euo pipefail

# Detect host OS and architecture once; all download URLs use these.
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"   # linux | darwin
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64|amd64)   ARCH=amd64 ;;
  arm64|aarch64)  ARCH=arm64 ;;
  *)
    echo "ERROR: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERSION_FILE="${SCRIPT_DIR}/../versions.env"

if [ ! -f "$VERSION_FILE" ]; then
    echo "Error: versions.env not found at $VERSION_FILE" >&2
    exit 1
fi

# shellcheck source=/dev/null
source "$VERSION_FILE"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# ---------------------------------------------------------------------------
# Tool installers
# ---------------------------------------------------------------------------

install_kubectl() {
    echo -e "${YELLOW}Installing kubectl (v${KUBECTL_VERSION})...${NC}"

    if command -v kubectl &>/dev/null; then
        current_version=$(kubectl version --client 2>/dev/null \
            | grep "Client Version:" | awk '{print $NF}' | sed 's/v//')
        if [ "$current_version" = "${KUBECTL_VERSION}" ]; then
            echo -e "${GREEN}kubectl v${KUBECTL_VERSION} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}kubectl version mismatch (have v${current_version}, want v${KUBECTL_VERSION})${NC}"
    fi

    echo "Downloading kubectl v${KUBECTL_VERSION} for ${OS}/${ARCH}..."
    curl -fsSLo /tmp/kubectl \
        "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl"
    chmod +x /tmp/kubectl
    sudo mv /tmp/kubectl /usr/local/bin/kubectl

    installed=$(kubectl version --client 2>/dev/null \
        | grep "Client Version:" | awk '{print $NF}' | sed 's/v//')
    if [ "$installed" = "${KUBECTL_VERSION}" ]; then
        echo -e "${GREEN}kubectl v${KUBECTL_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: kubectl version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
}

install_docker() {
    echo -e "${YELLOW}Checking Docker (v${DOCKER_VERSION})...${NC}"

    if command -v docker &>/dev/null; then
        current_version=$(docker --version 2>/dev/null \
            | sed 's/.*version \([^,]*\).*/\1/')
        if [ "$current_version" = "${DOCKER_VERSION}" ]; then
            echo -e "${GREEN}Docker v${DOCKER_VERSION} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}Docker v${current_version} installed (want v${DOCKER_VERSION})${NC}"
    fi

    echo -e "${YELLOW}Docker requires manual installation:${NC}"
    if [ "$OS" = "darwin" ]; then
        echo -e "${YELLOW}  macOS options (any one will work):${NC}"
        echo -e "${YELLOW}    Docker Desktop: https://www.docker.com/products/docker-desktop${NC}"
        echo -e "${YELLOW}    Colima:         brew install colima && colima start${NC}"
        echo -e "${YELLOW}    OrbStack:       https://orbstack.dev${NC}"
    else
        echo -e "${YELLOW}  Linux: https://docs.docker.com/engine/install/${NC}"
        echo -e "${YELLOW}  Ensure your user is in the 'docker' group.${NC}"
    fi
    echo -e "${RED}Docker installation skipped — please install manually and re-run.${NC}"
}

install_k3d() {
    echo -e "${YELLOW}Installing k3d (v${K3D_VERSION})...${NC}"

    if command -v k3d &>/dev/null; then
        current_version=$(k3d version 2>/dev/null \
            | grep 'k3d version' \
            | sed 's/.*k3d version v\([^ -]*\).*/\1/')
        if [ "$current_version" = "${K3D_VERSION}" ]; then
            echo -e "${GREEN}k3d v${K3D_VERSION} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}k3d version mismatch (have v${current_version}, want v${K3D_VERSION})${NC}"
    fi

    echo "Downloading k3d v${K3D_VERSION} for ${OS}-${ARCH}..."
    curl -fsSLo /tmp/k3d \
        "https://github.com/k3d-io/k3d/releases/download/v${K3D_VERSION}/k3d-${OS}-${ARCH}"
    chmod +x /tmp/k3d
    sudo mv /tmp/k3d /usr/local/bin/k3d

    installed=$(k3d version 2>/dev/null \
        | grep 'k3d version' \
        | sed 's/.*k3d version v\([^ -]*\).*/\1/')
    if [ "$installed" = "${K3D_VERSION}" ]; then
        echo -e "${GREEN}k3d v${K3D_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: k3d version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
}

install_helm() {
    echo -e "${YELLOW}Installing Helm (v${HELM_VERSION})...${NC}"

    if command -v helm &>/dev/null; then
        current_version=$(helm version --short 2>/dev/null \
            | sed 's/v\([^+]*\).*/\1/')
        if [ "$current_version" = "${HELM_VERSION}" ]; then
            echo -e "${GREEN}Helm v${HELM_VERSION} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}Helm version mismatch (have v${current_version}, want v${HELM_VERSION})${NC}"
    fi

    echo "Downloading Helm v${HELM_VERSION} for ${OS}-${ARCH}..."
    cd /tmp
    curl -fsSLo helm.tar.gz \
        "https://get.helm.sh/helm-v${HELM_VERSION}-${OS}-${ARCH}.tar.gz"
    tar -xzf helm.tar.gz
    sudo mv "${OS}-${ARCH}/helm" /usr/local/bin/helm
    rm -rf helm.tar.gz "${OS}-${ARCH}"
    cd - >/dev/null

    installed=$(helm version --short 2>/dev/null | sed 's/v\([^+]*\).*/\1/')
    if [ "$installed" = "${HELM_VERSION}" ]; then
        echo -e "${GREEN}Helm v${HELM_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: Helm version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
}

install_az_cli() {
    echo -e "${YELLOW}Installing Azure CLI (v${AZ_CLI_VERSION})...${NC}"

    if command -v az &>/dev/null; then
        current_version=$(az --version 2>/dev/null | head -1 | awk '{print $NF}')
        if [ "$current_version" = "${AZ_CLI_VERSION}" ]; then
            echo -e "${GREEN}Azure CLI v${AZ_CLI_VERSION} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}Azure CLI v${current_version} installed (want v${AZ_CLI_VERSION})${NC}"
    fi

    if [ "$OS" = "darwin" ]; then
        if command -v brew &>/dev/null; then
            echo "Installing Azure CLI via Homebrew..."
            brew install azure-cli
        else
            echo -e "${YELLOW}Homebrew not found. Install Azure CLI manually:${NC}"
            echo -e "${YELLOW}  https://learn.microsoft.com/en-us/cli/azure/install-azure-cli-macos${NC}"
            echo -e "${RED}Azure CLI installation skipped.${NC}"
            return 0
        fi
    else
        # Debian/Ubuntu only; for other distros see the docs above.
        curl -sL https://aka.ms/InstallAzureCLIDeb | sudo bash
    fi

    installed=$(az --version 2>/dev/null | head -1 | awk '{print $NF}')
    echo -e "${GREEN}Azure CLI v${installed} installed${NC}"
}

install_aws_cli() {
    echo -e "${YELLOW}Installing AWS CLI (v${AWS_CLI_VERSION:-2})...${NC}"

    if command -v aws &>/dev/null; then
        current_version=$(aws --version 2>/dev/null | awk '{print $1}' | cut -d/ -f2)
        echo -e "${GREEN}AWS CLI v${current_version} already installed${NC}"
        return 0
    fi

    echo "Downloading AWS CLI v2 for ${OS}/${ARCH}..."
    cd /tmp

    if [ "$OS" = "darwin" ]; then
        curl -fsSLo AWSCLIV2.pkg "https://awscli.amazonaws.com/AWSCLIV2.pkg"
        sudo installer -pkg AWSCLIV2.pkg -target /
        rm -f AWSCLIV2.pkg
    else
        # Map arm64 → aarch64 for the Linux archive name
        aws_arch="x86_64"
        [ "$ARCH" = "arm64" ] && aws_arch="aarch64"
        curl -fsSLo awscliv2.zip \
            "https://awscli.amazonaws.com/awscli-exe-linux-${aws_arch}.zip"
        unzip -qo awscliv2.zip
        sudo ./aws/install --update
        rm -rf awscliv2.zip aws
    fi

    cd - >/dev/null
    installed=$(aws --version 2>/dev/null | awk '{print $1}' | cut -d/ -f2)
    echo -e "${GREEN}AWS CLI v${installed} installed${NC}"
}

install_terraform() {
    local tf_version="${TERRAFORM_VERSION:-1.7.0}"
    echo -e "${YELLOW}Installing Terraform (v${tf_version})...${NC}"

    if command -v terraform &>/dev/null; then
        current_version=$(terraform version -json 2>/dev/null \
            | grep '"terraform_version"' \
            | sed 's/.*"terraform_version":[[:space:]]*"\([^"]*\)".*/\1/')
        if [ "$current_version" = "$tf_version" ]; then
            echo -e "${GREEN}Terraform v${tf_version} already installed${NC}"
            return 0
        fi
        echo -e "${YELLOW}Terraform v${current_version} installed (want v${tf_version})${NC}"
    fi

    echo "Downloading Terraform v${tf_version} for ${OS}/${ARCH}..."
    cd /tmp
    curl -fsSLo "terraform_${tf_version}.zip" \
        "https://releases.hashicorp.com/terraform/${tf_version}/terraform_${tf_version}_${OS}_${ARCH}.zip"
    unzip -qo "terraform_${tf_version}.zip"
    sudo mv terraform /usr/local/bin/terraform
    rm -f "terraform_${tf_version}.zip"
    cd - >/dev/null

    installed=$(terraform version -json 2>/dev/null \
        | grep '"terraform_version"' \
        | sed 's/.*"terraform_version":[[:space:]]*"\([^"]*\)".*/\1/')
    if [ "$installed" = "$tf_version" ]; then
        echo -e "${GREEN}Terraform v${tf_version} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: Terraform version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Profile groups
# ---------------------------------------------------------------------------

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
    kubectl cluster-info 2>/dev/null \
        || echo -e "${YELLOW}Cluster not yet running — run 'make runtime-up' first${NC}"
    echo -e "${GREEN}========== K3D Profile Complete ==========${NC}\n"
}

install_aks_profile() {
    echo -e "${GREEN}========== Installing AKS Profile ==========${NC}"
    install_common
    install_helm
    install_terraform
    install_az_cli
    echo -e "${GREEN}========== AKS Profile Complete ==========${NC}\n"
}

install_eks_profile() {
    echo -e "${GREEN}========== Installing EKS Profile ==========${NC}"
    install_common
    install_helm
    install_terraform
    install_aws_cli
    echo -e "${GREEN}========== EKS Profile Complete ==========${NC}\n"
}

# ---------------------------------------------------------------------------
# Entry point
# ---------------------------------------------------------------------------

show_help() {
    cat <<EOF
Usage: setup-tools.sh [PROFILE]

Detected: OS=${OS}, ARCH=${ARCH}

Profiles:
  k3d     kubectl + docker + k3d + helm      (local cluster)
  aks     kubectl + helm + terraform + az    (Azure AKS)
  eks     kubectl + helm + terraform + aws   (AWS EKS)
  common  kubectl only
  all     all of the above

Examples:
  ./setup-tools.sh k3d
  ./setup-tools.sh aks
EOF
}

main() {
    local profile="${1:-k3d}"

    printf "${GREEN}╔════════════════════════════════════════════╗\n"
    printf "║  Setup Tools — Profile: %-6s (%s/%s)\n" "$profile" "$OS" "$ARCH"
    printf "╚════════════════════════════════════════════╝\n\n${NC}"

    case "$profile" in
        k3d)    install_k3d_profile ;;
        aks)    install_aks_profile ;;
        eks)    install_eks_profile ;;
        common) install_common ;;
        all)
            install_k3d_profile
            install_aks_profile
            install_eks_profile
            ;;
        help|--help|-h)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Error: unknown profile '$profile'${NC}" >&2
            show_help
            exit 1
            ;;
    esac

    echo -e "${GREEN}Setup complete!${NC}"
}

main "$@"
