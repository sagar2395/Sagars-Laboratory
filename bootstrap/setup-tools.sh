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

INSTALL_DIR="/usr/local/bin"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Create /usr/local/bin if absent (default on Apple Silicon Macs).
ensure_install_dir() {
    if [ ! -d "$INSTALL_DIR" ]; then
        echo "Creating $INSTALL_DIR..."
        sudo mkdir -p "$INSTALL_DIR"
    fi
}

# Block until Docker daemon responds or timeout.
_wait_for_docker() {
    local retries=0
    echo "Waiting for Docker daemon..."
    until docker info &>/dev/null; do
        retries=$((retries + 1))
        if [ "$retries" -ge 60 ]; then
            echo -e "${RED}ERROR: Docker daemon not reachable after 60 s.${NC}" >&2
            echo "Check logs with: colima status   (macOS) or   sudo journalctl -u docker   (Linux)" >&2
            exit 1
        fi
        sleep 1
    done
}

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

    ensure_install_dir
    echo "Downloading kubectl v${KUBECTL_VERSION} for ${OS}/${ARCH}..."
    curl -fsSLo /tmp/kubectl \
        "https://dl.k8s.io/release/v${KUBECTL_VERSION}/bin/${OS}/${ARCH}/kubectl"
    chmod +x /tmp/kubectl
    sudo mv /tmp/kubectl "${INSTALL_DIR}/kubectl"

    installed=$(kubectl version --client 2>/dev/null \
        | grep "Client Version:" | awk '{print $NF}' | sed 's/v//')
    if [ "$installed" = "${KUBECTL_VERSION}" ]; then
        echo -e "${GREEN}kubectl v${KUBECTL_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: kubectl version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
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

    ensure_install_dir
    echo "Downloading k3d v${K3D_VERSION} for ${OS}-${ARCH}..."
    curl -fsSLo /tmp/k3d \
        "https://github.com/k3d-io/k3d/releases/download/v${K3D_VERSION}/k3d-${OS}-${ARCH}"
    chmod +x /tmp/k3d
    sudo mv /tmp/k3d "${INSTALL_DIR}/k3d"

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

    ensure_install_dir
    echo "Downloading Helm v${HELM_VERSION} for ${OS}-${ARCH}..."
    local helm_tmp
    helm_tmp="$(mktemp -d)"
    curl -fsSLo "${helm_tmp}/helm.tar.gz" \
        "https://get.helm.sh/helm-v${HELM_VERSION}-${OS}-${ARCH}.tar.gz"
    tar -xzf "${helm_tmp}/helm.tar.gz" -C "${helm_tmp}"
    sudo mv "${helm_tmp}/${OS}-${ARCH}/helm" "${INSTALL_DIR}/helm"
    rm -rf "${helm_tmp}"

    installed=$(helm version --short 2>/dev/null | sed 's/v\([^+]*\).*/\1/')
    if [ "$installed" = "${HELM_VERSION}" ]; then
        echo -e "${GREEN}Helm v${HELM_VERSION} installed and verified${NC}"
    else
        echo -e "${RED}ERROR: Helm version mismatch after install (got v${installed})${NC}" >&2
        exit 1
    fi
}

# ---------------------------------------------------------------------------
# Container runtime — Colima (macOS) or Docker Engine (Linux)
# ---------------------------------------------------------------------------

# macOS: ensure Homebrew is present.
_install_homebrew() {
    if command -v brew &>/dev/null; then
        echo -e "${GREEN}Homebrew already installed${NC}"
        return 0
    fi
    echo "Installing Homebrew (this may take a few minutes)..."
    NONINTERACTIVE=1 /bin/bash -c \
        "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    # Activate brew in the current shell — path differs between Apple Silicon and Intel.
    if [ -x /opt/homebrew/bin/brew ]; then
        eval "$(/opt/homebrew/bin/brew shellenv)"
    elif [ -x /usr/local/bin/brew ]; then
        eval "$(/usr/local/bin/brew shellenv)"
    fi
    echo -e "${GREEN}Homebrew installed${NC}"
}

# macOS: install Colima + Docker CLI, then start the VM.
install_colima() {
    echo -e "${YELLOW}Setting up Colima (macOS container runtime)...${NC}"

    if ! command -v colima &>/dev/null; then
        _install_homebrew
        echo "Installing colima and docker CLI via Homebrew..."
        brew install colima docker
    else
        echo -e "${GREEN}Colima already installed${NC}"
    fi

    if colima status &>/dev/null; then
        echo -e "${GREEN}Colima already running${NC}"
        return 0
    fi

    echo "Starting Colima (first run may download a VM image — ~1 min)..."
    colima start
    _wait_for_docker
    echo -e "${GREEN}Colima started — Docker daemon ready${NC}"
}

# Linux: detect distro ID from /etc/os-release.
_linux_distro() {
    if [ -f /etc/os-release ]; then
        # shellcheck source=/dev/null
        . /etc/os-release
        echo "${ID:-unknown}"
    else
        echo "unknown"
    fi
}

# Linux Debian/Ubuntu family — Docker's official apt repo.
_docker_apt() {
    local distro="$1"
    echo "Installing Docker Engine via apt (${distro})..."
    sudo apt-get update -qq
    sudo apt-get install -y -qq ca-certificates curl gnupg lsb-release
    sudo install -m 0755 -d /etc/apt/keyrings
    curl -fsSL "https://download.docker.com/linux/${distro}/gpg" \
        | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    sudo chmod a+r /etc/apt/keyrings/docker.gpg
    local codename
    codename="$(. /etc/os-release && echo "${VERSION_CODENAME:-}")"
    [ -z "$codename" ] && codename="$(lsb_release -cs 2>/dev/null || echo "")"
    printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/%s %s stable\n' \
        "$(dpkg --print-architecture)" "$distro" "$codename" \
        | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
    sudo apt-get update -qq
    sudo apt-get install -y -qq \
        docker-ce docker-ce-cli containerd.io \
        docker-buildx-plugin docker-compose-plugin
}

# Linux RHEL/CentOS/Fedora/Rocky/Alma family — Docker's official rpm repo.
_docker_rpm() {
    local repo_distro="$1"   # centos | fedora | rhel
    echo "Installing Docker Engine via dnf/yum (${repo_distro})..."
    local pm="yum"
    command -v dnf &>/dev/null && pm="dnf"
    sudo "$pm" install -y "${pm}-plugins-core" 2>/dev/null \
        || sudo "$pm" install -y yum-utils 2>/dev/null || true
    sudo "$pm" config-manager \
        --add-repo "https://download.docker.com/linux/${repo_distro}/docker-ce.repo"
    sudo "$pm" install -y \
        docker-ce docker-ce-cli containerd.io \
        docker-buildx-plugin docker-compose-plugin
}

# Linux: install Docker Engine natively using the host package manager.
install_docker_linux() {
    echo -e "${YELLOW}Installing Docker Engine (Linux)...${NC}"

    if command -v docker &>/dev/null && docker info &>/dev/null; then
        echo -e "${GREEN}Docker Engine already installed and running${NC}"
        return 0
    fi

    local distro
    distro="$(_linux_distro)"

    case "$distro" in
        ubuntu|debian|linuxmint|pop|elementary|kali|raspbian)
            _docker_apt "$distro"
            ;;
        fedora)
            _docker_rpm "fedora"
            ;;
        rhel|centos|rocky|almalinux)
            _docker_rpm "centos"
            ;;
        amzn)
            # Amazon Linux 2 / 2023 ships Docker in its own repo.
            sudo yum install -y docker
            ;;
        arch|manjaro|endeavouros)
            sudo pacman -Sy --noconfirm docker
            ;;
        alpine)
            sudo apk add --no-cache docker
            ;;
        *)
            echo -e "${RED}ERROR: Unsupported Linux distribution '${distro}'.${NC}" >&2
            echo "Install Docker manually: https://docs.docker.com/engine/install/" >&2
            exit 1
            ;;
    esac

    # Enable and start the Docker service.
    if command -v systemctl &>/dev/null && systemctl list-units --type=service &>/dev/null; then
        sudo systemctl enable docker
        sudo systemctl start docker
    elif command -v service &>/dev/null; then
        sudo service docker start
    fi

    # Add current user to the docker group for non-root access.
    local current_user="${USER:-$(id -un)}"
    if ! id -nG "$current_user" 2>/dev/null | grep -qw docker; then
        sudo usermod -aG docker "$current_user"
        echo -e "${YELLOW}Added '$current_user' to the docker group.${NC}"
        echo -e "${YELLOW}NOTE: Run 'newgrp docker' or re-login for the change to take effect.${NC}"
    fi

    _wait_for_docker
    echo -e "${GREEN}Docker Engine installed and running${NC}"
}

# Public entry point — routes to the right runtime for the current OS.
install_docker() {
    if [ "$OS" = "darwin" ]; then
        install_colima
    else
        install_docker_linux
    fi
}

# ---------------------------------------------------------------------------
# Cloud CLI tools
# ---------------------------------------------------------------------------

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
        _install_homebrew
        brew install azure-cli
    else
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

    local aws_tmp
    aws_tmp="$(mktemp -d)"

    if [ "$OS" = "darwin" ]; then
        echo "Downloading AWS CLI v2 for macOS..."
        curl -fsSLo "${aws_tmp}/AWSCLIV2.pkg" "https://awscli.amazonaws.com/AWSCLIV2.pkg"
        sudo installer -pkg "${aws_tmp}/AWSCLIV2.pkg" -target /
    else
        local aws_arch="x86_64"
        [ "$ARCH" = "arm64" ] && aws_arch="aarch64"
        echo "Downloading AWS CLI v2 for Linux/${aws_arch}..."
        curl -fsSLo "${aws_tmp}/awscliv2.zip" \
            "https://awscli.amazonaws.com/awscli-exe-linux-${aws_arch}.zip"
        unzip -qo "${aws_tmp}/awscliv2.zip" -d "${aws_tmp}"
        sudo "${aws_tmp}/aws/install" --update
    fi

    rm -rf "${aws_tmp}"
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

    ensure_install_dir
    echo "Downloading Terraform v${tf_version} for ${OS}/${ARCH}..."
    local tf_tmp
    tf_tmp="$(mktemp -d)"
    curl -fsSLo "${tf_tmp}/terraform.zip" \
        "https://releases.hashicorp.com/terraform/${tf_version}/terraform_${tf_version}_${OS}_${ARCH}.zip"
    unzip -qo "${tf_tmp}/terraform.zip" -d "${tf_tmp}"
    sudo mv "${tf_tmp}/terraform" "${INSTALL_DIR}/terraform"
    rm -rf "${tf_tmp}"

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
    install_docker   # Colima on macOS; Docker Engine on Linux
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
  k3d     kubectl + colima/docker + k3d + helm   (local cluster)
  aks     kubectl + helm + terraform + az         (Azure AKS)
  eks     kubectl + helm + terraform + aws        (AWS EKS)
  common  kubectl only
  all     all of the above

Container runtime installed per OS:
  macOS   → Colima (lightweight Docker VM via Homebrew)
  Linux   → Docker Engine (native; distro auto-detected)

Examples:
  ./setup-tools.sh k3d
  ./setup-tools.sh aks
  make setup-tools PROFILE=k3d
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
