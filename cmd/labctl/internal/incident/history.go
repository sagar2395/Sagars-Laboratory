package incident

// MTTR tracking (task 048): every completed incident run is appended to an
// append-only JSONL file in .labctl/history/. Time-to-detect is proxied by
// the first `incident status` call; time-to-resolve is when the detection
// check first passes (or the escape hatch runs). Task 052 will fold this
// into the unified results store.

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Record is one completed incident run.
type Record struct {
	Fault          string    `json:"fault"`
	Category       string    `json:"category"`
	Severity       string    `json:"severity"`
	Silent         bool      `json:"silent"`
	InjectedAt     time.Time `json:"injectedAt"`
	FirstCheckedAt time.Time `json:"firstCheckedAt,omitempty"`
	ResolvedAt     time.Time `json:"resolvedAt"`
	// DetectSeconds: injection → first status check (how long until someone looked).
	DetectSeconds int64 `json:"detectSeconds,omitempty"`
	// ResolveSeconds: injection → resolution (the MTTR).
	ResolveSeconds int64 `json:"resolveSeconds"`
	HintsUsed      int   `json:"hintsUsed"`
	// ResolvedBy: "manual" (detection check passed) or "auto" (resolve.sh
	// escape hatch — scores as a non-completion in challenge mode).
	ResolvedBy string `json:"resolvedBy"`
}

func (e *Engine) historyFile() string {
	return filepath.Join(e.ProjectRoot, ".labctl", "history", "incidents.jsonl")
}

func (e *Engine) appendHistory(rec Record) error {
	if err := os.MkdirAll(filepath.Dir(e.historyFile()), 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(e.historyFile(), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// History returns all recorded runs, oldest first. Corrupt lines are
// skipped — history must never block operations.
func (e *Engine) History() ([]Record, error) {
	f, err := os.Open(e.historyFile())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var out []Record
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var rec Record
		if err := json.Unmarshal(scanner.Bytes(), &rec); err != nil {
			continue
		}
		out = append(out, rec)
	}
	return out, scanner.Err()
}

// finishRun builds and appends the run record when an incident ends.
func (e *Engine) finishRun(active *Active, f *Fault, resolvedBy string) {
	now := time.Now().UTC()
	rec := Record{
		Fault:          f.Name,
		Category:       f.Category,
		Severity:       f.Severity,
		Silent:         active.Silent,
		InjectedAt:     active.InjectedAt,
		FirstCheckedAt: active.FirstCheckedAt,
		ResolvedAt:     now,
		ResolveSeconds: int64(now.Sub(active.InjectedAt).Seconds()),
		HintsUsed:      active.HintsRevealed,
		ResolvedBy:     resolvedBy,
	}
	if !active.FirstCheckedAt.IsZero() {
		rec.DetectSeconds = int64(active.FirstCheckedAt.Sub(active.InjectedAt).Seconds())
	}
	// Best effort: history failures must not fail the resolution itself.
	_ = e.appendHistory(rec)
}
