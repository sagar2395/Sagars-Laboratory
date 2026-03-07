#!/usr/bin/env bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Shutting down k3d cluster...${NC}"

# Get list of running k3d clusters
CLUSTERS=$(k3d cluster list 2>/dev/null | awk '{print $1}')

if [ -z "$CLUSTERS" ]; then
   echo -e "${YELLOW}No k3d clusters found.${NC}"
   exit 0
fi

# Delete each cluster
for cluster in $CLUSTERS; do
   echo "Deleting cluster: $cluster"
   k3d cluster delete "$cluster" || {
      echo -e "${RED}Failed to delete cluster: $cluster${NC}"
      exit 1
   }
done

echo -e "${YELLOW}Validating cluster shutdown...${NC}"

# Verify all clusters are deleted
REMAINING=$(k3d cluster list --no-header 2>/dev/null | awk '{print $1}' | wc -l)

if [ "$REMAINING" -eq 0 ]; then
   echo -e "${GREEN}✓ All k3d clusters have been successfully shut down.${NC}"
   exit 0
else
   echo -e "${RED}✗ Validation failed: $REMAINING cluster(s) still exist.${NC}"
   exit 1
fi