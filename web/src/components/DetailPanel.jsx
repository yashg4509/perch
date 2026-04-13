import { X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { mockDeployments, mockLogs } from '../data/mock.js'
import { DeployRow } from './DeployRow.jsx'

function providerTitle(provider) {
  const map = {
    vercel: 'Vercel',
    supabase: 'Supabase',
    github: 'GitHub',
    render: 'Render',
    custom: 'Custom',
  }
  return map[provider] ?? provider
}

function metaLookup(meta, key) {
  return meta?.find((m) => m.key === key)?.value ?? '—'
}

const statusLabel = {
  healthy: 'healthy',
  degraded: 'degraded',
  down: 'down',
  source: 'source',
  unknown: 'unknown',
}

/** @param {{ node: object | null, environment: string }} props */
export function DetailPanel({ node, environment }) {
  const { stackName, nodeId } = useParams()
  const navigate = useNavigate()
  const [tab, setTab] = useState('deployments')

  const close = () => {
    navigate(`/stack/${stackName}`)
  }

  useEffect(() => {
    const onKey = (e) => {
      if (e.key === 'Escape') navigate(`/stack/${stackName}`)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [navigate, stackName])

  const deployments = node ? mockDeployments[node.id] ?? [] : []
  const logs = node ? mockLogs[node.id] ?? [] : []

  const region = node ? metaLookup(node.meta, 'region') : '—'
  const branch = node ? metaLookup(node.meta, 'branch') : '—'
  const st = node ? statusLabel[node.status] ?? node.status : '—'

  return (
    <aside className="flex h-full w-[300px] shrink-0 flex-col border-l border-gray-200 bg-white">
      <div className="flex items-start justify-between gap-2 border-b border-gray-200 px-3 py-3">
        <div className="min-w-0">
          <div className="text-[13px] font-medium leading-tight text-black">
            {node ? node.label : 'Not found'}
          </div>
          <div className="mt-0.5 text-[11px] text-gray-400">
            {node ? (
              <>
                {providerTitle(node.provider)} · {environment}
              </>
            ) : (
              <span className="font-mono text-gray-500">{nodeId}</span>
            )}
          </div>
        </div>
        <button
          type="button"
          onClick={close}
          className="rounded-md p-1 text-gray-500 hover:bg-gray-100 hover:text-gray-900"
          aria-label="Close panel"
        >
          <X className="h-4 w-4" strokeWidth={2} />
        </button>
      </div>

      {node && (
        <div className="grid grid-cols-3 gap-2 border-b border-gray-200 px-3 py-2">
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">region</div>
            <div className="truncate text-xs font-medium text-gray-900">{region}</div>
          </div>
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">status</div>
            <div className="truncate text-xs font-medium text-gray-900">{st}</div>
          </div>
          <div>
            <div className="text-[10px] uppercase tracking-wide text-gray-400">branch</div>
            <div className="truncate text-xs font-medium text-gray-900">{branch}</div>
          </div>
        </div>
      )}

      {!node && (
        <div className="border-b border-gray-200 px-3 py-3 text-sm text-gray-500">
          No node exists for this id in the mock graph.
        </div>
      )}

      {node && (
        <>
          <div className="flex border-b border-gray-200 text-xs text-gray-500">
            {[
              { id: 'deployments', label: 'deployments' },
              { id: 'logs', label: 'logs' },
              { id: 'env', label: 'env vars' },
            ].map((t) => (
              <button
                key={t.id}
                type="button"
                onClick={() => setTab(t.id)}
                className={`flex-1 border-r border-gray-200 px-2 py-2 capitalize last:border-r-0 hover:bg-gray-50 hover:text-gray-900 ${
                  tab === t.id ? 'bg-gray-50 text-gray-900' : ''
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>

          <div className="min-h-0 flex-1 overflow-y-auto">
            {tab === 'deployments' && (
              <div>
                {deployments.map((d) => (
                  <DeployRow key={d.id} deployment={d} />
                ))}
              </div>
            )}

            {tab === 'logs' && (
              <div className="space-y-1 p-3 font-mono text-xs leading-relaxed">
                {logs.map((line, i) => {
                  const levelClass =
                    line.level === 'warn'
                      ? 'text-amber-500'
                      : line.level === 'error'
                        ? 'text-red-500'
                        : 'text-gray-500'
                  return (
                    <div key={`${line.ts}-${i}`} className={levelClass}>
                      [{line.ts}] [{line.level}] {line.msg}
                    </div>
                  )
                })}
              </div>
            )}

            {tab === 'env' && (
              <p className="p-3 text-sm italic text-gray-500">
                env vars are redacted — open provider dashboard to reveal
              </p>
            )}
          </div>
        </>
      )}
    </aside>
  )
}
