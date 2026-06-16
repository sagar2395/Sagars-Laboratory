import { useState, useEffect, useCallback } from 'react'
import type { NotifyFn, PackEntry } from '../types'
import { api } from '../api/client'

interface Props {
  notify: NotifyFn
  refreshTick: number
  requestConfirm: (req: import('../components/ConfirmDialog').ConfirmRequest) => void
}

/** Format a star rating like "★★★★½ (42)" */
function formatRating(rating: number, count: number): string {
  if (count === 0) return '—'
  const full = Math.floor(rating)
  const half = rating - full >= 0.5 ? 1 : 0
  return '★'.repeat(full) + (half ? '½' : '') + `☆`.repeat(5 - full - half) + ` (${count})`
}

function TierBadge({ tier }: { tier: string }) {
  const cls = tier === 'enterprise' ? 'badge-enterprise'
    : tier === 'premium' ? 'badge-premium'
    : 'badge-community'
  return <span className={`badge ${cls}`}>{tier || 'community'}</span>
}

function VerifiedBadge({ verified }: { verified?: boolean }) {
  if (!verified) return null
  return <span className="badge badge-verified" title="Verified publisher">✓ Verified</span>
}

export function Marketplace({ notify, refreshTick, requestConfirm }: Props) {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<PackEntry[]>([])
  const [installed, setInstalled] = useState<PackEntry[]>([])
  const [loading, setLoading] = useState(false)
  const [activeTab, setActiveTab] = useState<'browse' | 'installed'>('browse')
  const [selectedPack, setSelectedPack] = useState<PackEntry | null>(null)
  const [busy, setBusy] = useState<string | null>(null) // pack name being acted on

  const loadInstalled = useCallback(() => {
    api.listInstalledPacks()
      .then(setInstalled)
      .catch(e => notify('error', 'Failed to load installed packs', e instanceof Error ? e.message : String(e)))
  }, [notify])

  const search = useCallback((term: string) => {
    setLoading(true)
    api.searchPacks(term)
      .then(setResults)
      .catch(e => notify('error', 'Pack search failed', e instanceof Error ? e.message : String(e)))
      .finally(() => setLoading(false))
  }, [notify])

  useEffect(() => {
    search(query)
    loadInstalled()
  }, [refreshTick]) // eslint-disable-line react-hooks/exhaustive-deps

  function handleSearch(e: React.FormEvent) {
    e.preventDefault()
    search(query)
  }

  function handleInstall(pack: PackEntry) {
    requestConfirm({
      title: `Install "${pack.displayName || pack.name}"?`,
      message: `This will download and install the pack. Only install packs from sources you trust.`,
      confirmLabel: 'Install',
      danger: false,
      onConfirm: async () => {
        setBusy(pack.name)
        try {
          await api.installPack(pack.name)
          notify('success', `Installing ${pack.name}`, 'Pack installation started.')
          loadInstalled()
        } catch (e) {
          notify('error', `Failed to install ${pack.name}`, e instanceof Error ? e.message : String(e))
        } finally {
          setBusy(null)
        }
      },
    })
  }

  function handleRemove(pack: PackEntry) {
    requestConfirm({
      title: `Remove "${pack.name}"?`,
      message: `This will uninstall the pack and its scenarios.`,
      confirmLabel: 'Remove',
      danger: true,
      onConfirm: async () => {
        setBusy(pack.name)
        try {
          await api.removePack(pack.name)
          notify('success', `Removed ${pack.name}`)
          loadInstalled()
          search(query)
        } catch (e) {
          notify('error', `Failed to remove ${pack.name}`, e instanceof Error ? e.message : String(e))
        } finally {
          setBusy(null)
        }
      },
    })
  }

  const displayList = activeTab === 'installed' ? installed : results

  return (
    <div className="view-container">
      {selectedPack ? (
        <PackDetail
          pack={selectedPack}
          onBack={() => setSelectedPack(null)}
          onInstall={handleInstall}
          onRemove={handleRemove}
          busy={busy === selectedPack.name}
        />
      ) : (
        <>
          <div className="view-header">
            <h2>Marketplace</h2>
            <p className="view-subtitle">
              Browse and install scenario packs from the community registry.
              Community packs are free and open-source; premium packs require entitlement.
            </p>
          </div>

          <div className="marketplace-tabs">
            <button
              className={`marketplace-tab${activeTab === 'browse' ? ' active' : ''}`}
              onClick={() => setActiveTab('browse')}
            >
              Browse ({results.length})
            </button>
            <button
              className={`marketplace-tab${activeTab === 'installed' ? ' active' : ''}`}
              onClick={() => setActiveTab('installed')}
            >
              Installed ({installed.length})
            </button>
          </div>

          {activeTab === 'browse' && (
            <form className="marketplace-search" onSubmit={handleSearch}>
              <input
                className="search-input"
                type="search"
                placeholder="Search packs…"
                value={query}
                onChange={e => setQuery(e.target.value)}
                aria-label="Search packs"
              />
              <button className="btn btn-sm" type="submit" disabled={loading}>
                {loading ? 'Searching…' : 'Search'}
              </button>
            </form>
          )}

          {loading && <div className="loading-spinner" role="status">Loading…</div>}

          {!loading && displayList.length === 0 && (
            <div className="empty-state">
              {activeTab === 'installed'
                ? 'No packs installed. Browse the marketplace to install one.'
                : query
                  ? `No packs matching "${query}".`
                  : 'No packs available — check your PACK_REGISTRY_INDEX configuration.'
              }
            </div>
          )}

          <div className="pack-grid">
            {displayList.map(pack => (
              <div
                key={pack.name}
                className="pack-card"
                onClick={() => setSelectedPack(pack)}
                role="button"
                tabIndex={0}
                onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') setSelectedPack(pack) }}
              >
                <div className="pack-card-header">
                  <div className="pack-card-name">{pack.displayName || pack.name}</div>
                  <div className="pack-card-badges">
                    <TierBadge tier={pack.tier} />
                    <VerifiedBadge verified={pack.verified} />
                    {pack.installed && <span className="badge badge-installed">Installed</span>}
                  </div>
                </div>
                <div className="pack-card-meta">
                  <span className="pack-publisher">{pack.publisher || '—'}</span>
                  {pack.version && <span className="pack-version">v{pack.version}</span>}
                </div>
                <p className="pack-description">{pack.description || 'No description.'}</p>
                {pack.rating !== undefined && pack.rating > 0 && (
                  <div className="pack-rating">{formatRating(pack.rating, pack.ratingCount ?? 0)}</div>
                )}
                {(pack.keywords ?? []).length > 0 && (
                  <div className="pack-keywords">
                    {(pack.keywords ?? []).slice(0, 4).map(k => (
                      <span key={k} className="pack-keyword">{k}</span>
                    ))}
                  </div>
                )}
                <div className="pack-card-actions" onClick={e => e.stopPropagation()}>
                  {pack.installed ? (
                    <button
                      className="btn btn-sm btn-danger"
                      onClick={() => handleRemove(pack)}
                      disabled={busy === pack.name}
                    >
                      {busy === pack.name ? 'Removing…' : 'Remove'}
                    </button>
                  ) : (
                    <button
                      className="btn btn-sm"
                      onClick={() => handleInstall(pack)}
                      disabled={busy === pack.name}
                    >
                      {busy === pack.name ? 'Installing…' : 'Install'}
                    </button>
                  )}
                </div>
              </div>
            ))}
          </div>
        </>
      )}
    </div>
  )
}

function PackDetail({
  pack, onBack, onInstall, onRemove, busy,
}: {
  pack: PackEntry
  onBack: () => void
  onInstall: (p: PackEntry) => void
  onRemove: (p: PackEntry) => void
  busy: boolean
}) {
  return (
    <div className="pack-detail">
      <button className="btn btn-sm" onClick={onBack}>&larr; Back</button>
      <div className="pack-detail-header">
        <h3>{pack.displayName || pack.name}</h3>
        <div className="pack-detail-badges">
          <TierBadge tier={pack.tier} />
          <VerifiedBadge verified={pack.verified} />
          {pack.installed && <span className="badge badge-installed">Installed</span>}
        </div>
      </div>

      <dl className="pack-detail-meta">
        <dt>Name</dt>      <dd>{pack.name}</dd>
        {pack.version &&  <><dt>Version</dt>   <dd>v{pack.version}</dd></>}
        {pack.publisher && <><dt>Publisher</dt> <dd>{pack.publisher}</dd></>}
        {pack.license &&  <><dt>License</dt>   <dd>{pack.license}</dd></>}
        {pack.ref &&      <><dt>Source</dt>    <dd className="pack-ref">{pack.ref}</dd></>}
        {pack.downloads !== undefined && pack.downloads > 0 && (
          <><dt>Downloads</dt><dd>{pack.downloads.toLocaleString()}</dd></>
        )}
        {pack.rating !== undefined && pack.rating > 0 && (
          <><dt>Rating</dt><dd>{formatRating(pack.rating, pack.ratingCount ?? 0)}</dd></>
        )}
      </dl>

      {pack.description && <p className="pack-detail-desc">{pack.description}</p>}

      {(pack.keywords ?? []).length > 0 && (
        <div className="pack-keywords">
          {(pack.keywords ?? []).map(k => <span key={k} className="pack-keyword">{k}</span>)}
        </div>
      )}

      {(pack.categories ?? []).length > 0 && (
        <div className="pack-categories">
          {(pack.categories ?? []).map(c => <span key={c} className="pack-category">{c}</span>)}
        </div>
      )}

      <div className="pack-detail-actions">
        {pack.installed ? (
          <button className="btn btn-danger" onClick={() => onRemove(pack)} disabled={busy}>
            {busy ? 'Removing…' : 'Remove Pack'}
          </button>
        ) : (
          <button className="btn" onClick={() => onInstall(pack)} disabled={busy}>
            {busy ? 'Installing…' : 'Install Pack'}
          </button>
        )}
      </div>
    </div>
  )
}
