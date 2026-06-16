// SPDX-License-Identifier: Apache-2.0
package api

import (
	"net/http"

	"go.flightdeck.dev/labctl/pkg/edition"
)

// handleEditionInfo handles GET /api/edition (task 076).
// Returns the active edition and its feature line so the UI can show
// which features are available without hardcoding edition logic.
func (s *Server) handleEditionInfo(w http.ResponseWriter, r *http.Request) {
	ed := edition.Current()
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"edition":     string(ed),
		"displayName": ed.DisplayName(),
		"summary":     ed.Summary(),
		"features":    ed.FeatureLine(),
	})
}
