// SPDX-License-Identifier: Apache-2.0

package scenario

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"go.flightdeck.dev/labctl/pkg/checks"
	"go.flightdeck.dev/labctl/pkg/entitlement"
	"go.flightdeck.dev/labctl/pkg/extension"
)

// denyAll is an injected entitlement test-double: it refuses every pack, the way
// a premium policy would reject an unentitled install.
type denyAll struct{}

func (denyAll) Authorize(_ context.Context, req entitlement.Request) error {
	return entitlement.Denied(req.Name, "test policy denies all")
}

func TestInstall_RoutesThroughEntitlement(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "pack-scenario"), packScenarioYAML)

	engine := NewEngine(root, "k3d.local", "k3d")
	// Default (allow-all) lets it install.
	if _, err := engine.InstallPackFromDir(fixture, "allowed-pack", false); err != nil {
		t.Fatalf("default entitlement should allow install: %v", err)
	}

	// Inject a denying entitlement → install is refused cleanly and nothing lands.
	engine.Entitlement = denyAll{}
	_, err := engine.InstallPackFromDir(fixture, "denied-pack", false)
	var de *entitlement.DeniedError
	if !errors.As(err, &de) {
		t.Fatalf("expected a *entitlement.DeniedError, got %v", err)
	}
	if _, err := NewEngine(root, "k3d.local", "k3d").Get("pack-scenario"); err == nil {
		// the allowed-pack install above provides pack-scenario; ensure the denied
		// one did not create a second catalog entry / partial install.
	}
	if len(engine.Packs()) != 1 {
		t.Fatalf("denied install must not land; packs = %d", len(engine.Packs()))
	}
}

// denyCheckHooks aborts at PreCheck, proving the run path routes through the
// extension.Hooks seam (a premium policy could gate checks this way).
type denyCheckHooks struct {
	extension.NoopHooks
	called bool
}

func (h *denyCheckHooks) PreCheck(_ context.Context, ev extension.Event) error {
	h.called = true
	return errors.New("blocked check " + ev.Check)
}

func TestVerify_RoutesThroughHooks(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "pack-scenario"), packScenarioYAML)
	engine := NewEngine(root, "k3d.local", "k3d")
	if _, err := engine.InstallPackFromDir(fixture, "hooks-pack", false); err != nil {
		t.Fatal(err)
	}

	hooks := &denyCheckHooks{}
	engine.Hooks = hooks
	_, err := engine.Verify(context.Background(), "pack-scenario", checks.NewRunner())
	if err == nil || !strings.Contains(err.Error(), "pre-check hook") {
		t.Fatalf("expected pre-check hook to abort Verify, got %v", err)
	}
	if !hooks.called {
		t.Error("PreCheck hook was not invoked on the run path")
	}
}

func TestInstallVia_UsesResolverAndEntitlement(t *testing.T) {
	root := t.TempDir()
	fixture := t.TempDir()
	writeScenarioDir(t, filepath.Join(fixture, "pack-scenario"), packScenarioYAML)

	engine := NewEngine(root, "k3d.local", "k3d")
	// LocalResolver copies the fixture; the engine installs it via the seam.
	p, err := engine.InstallVia(context.Background(), extension.LocalResolver{}, fixture, "via-pack", false)
	if err != nil {
		t.Fatalf("InstallVia: %v", err)
	}
	if p.Name != "via-pack" || len(p.Scenarios) != 1 {
		t.Fatalf("pack = %+v", p)
	}
	if _, err := engine.Get("pack-scenario"); err != nil {
		t.Fatalf("resolver-installed scenario not discoverable: %v", err)
	}

	// A denying entitlement still blocks the resolver path.
	engine.Entitlement = denyAll{}
	if _, err := engine.InstallVia(context.Background(), extension.LocalResolver{}, fixture, "blocked", false); err == nil ||
		!strings.Contains(err.Error(), "not entitled") {
		t.Fatalf("expected entitlement denial through InstallVia, got %v", err)
	}
}
