import { X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
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

/** @param {{ key: string, value: string }[] | undefined} meta */
function metaValue(meta, key) {
  const v = meta?.find((m) => m.key === key)?.value
  if (v == null) {
    return ''
  }
  return String(v)
}

/**
 * @param {{ meta?: { key: string, value: string }[], recentErrors?: string[] } | null} node
 * @param {string} environment
 */
function buildDeploymentRows(node, environment) {
  if (!node) {
    return []
  }
  const sha = metaValue(node.meta, 'last_deploy.sha').trim()
  const ago = metaValue(node.meta, 'last_deploy.ago').trim()
  const hasLastDeploy = sha !== '' || ago !== ''

  const rows = []
  if (hasLastDeploy) {
    rows.push({
      id: sha !== '' ? sha.slice(0, 7) : 'deploy',
      env: environment,
      current: true,
      branch: '—',
      commit: sha !== '' ? sha : '—',
      date: ago !== '' ? ago : '—',
      duration: '—',
    })
  }

  const errs = Array.isArray(node.recentErrors) ? node.recentErrors : []
  errs.forEach((msg, i) => {
    const t = String(msg ?? '').trim()
    if (t === '') {
      return
    }
    rows.push({
      id: `recent-${i}-${t.slice(0, 12)}`,
      env: 'Recent activity',
      current: false,
      branch: '—',
      commit: t,
      date: '—',
      duration: '—',
    })
  })

  return rows
}

const statusLabel = {
  healthy: 'healthy',
  degraded: 'degraded',
  down: 'down',
  source: 'source',
  unknown: 'unknown',
}

/** @param {{ text: string, copied: boolean, onCopy: () => void }} props */
function CopyBlock({ label, text, copied, onCopy }) {
  const trimmed = text.trim()
  if (trimmed === '') {
    return null
  }
  return (
    <div className="mt-3 first:mt-0">
      <div className="text-[11px] font-medium text-gray-600">{label}</div>
      <pre className="mt-1 max-h-[28vh] overflow-auto whitespace-pre-wrap break-all rounded-md border border-gray-200 bg-gray-50 p-2 font-mono text-[11px] leading-relaxed text-gray-900">
        {trimmed}
      </pre>
      <button
        type="button"
        onClick={onCopy}
        className="mt-1.5 rounded-md border border-gray-300 bg-white px-2.5 py-1 text-xs font-medium text-gray-800 hover:bg-gray-50"
      >
        {copied ? 'Copied' : 'Copy'}
      </button>
    </div>
  )
}

/** @param {{ node: object | null, environment: string }} props */
export function DetailPanel({ node, environment }) {
  const { stackName, nodeId } = useParams()
  const navigate = useNavigate()
  const [tab, setTab] = useState('deployments')
  const [copyWhich, setCopyWhich] = useState(null)

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

  useEffect(() => {
    setCopyWhich(null)
  }, [node?.id, tab])

  const deployments = node ? buildDeploymentRows(node, environment) : []
  const logsCmd = node?.logs != null ? String(node.logs).trim() : ''
  const statusCmd = node?.statusCommand != null ? String(node.statusCommand).trim() : ''
  const hasLogsCommand = logsCmd !== ''
  const hasStatusCommand = statusCmd !== ''

  const region = node ? metaLookup(node.meta, 'region') : '—'
  const branch = node ? metaLookup(node.meta, 'branch') : '—'
  const st = node ? statusLabel[node.status] ?? node.status : '—'
  const project = node ? metaLookup(node.meta, 'project') : '—'
  const service = node ? metaLookup(node.meta, 'service') : '—'
  const errRate = node ? metaValue(node.meta, 'error_rate').trim() : ''
  const dailyCost = node ? metaValue(node.meta, 'daily_cost_usd').trim() : ''

  const copyText = async (which, text) => {
    try {
      await navigator.clipboard.writeText(text)
      setCopyWhich(which)
      window.setTimeout(() => setCopyWhich(null), 2000)
    } catch {
      setCopyWhich(null)
    }
  }

  const recentList = node && Array.isArray(node.recentErrors) ? node.recentErrors.filter((s) => String(s).trim() !== '') : []

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

      {node && (project !== '—' || service !== '—' || errRate !== '' || dailyCost !== '' || recentList.length > 0) && (
        <div className="space-y-1.5 border-b border-gray-200 px-3 py-2 text-[11px] text-gray-600">
          {(project !== '—' || service !== '—') && (
            <div className="flex flex-wrap gap-x-2 gap-y-0.5">
              {project !== '—' && (
                <span>
                  <span className="text-gray-400">project</span> {project}
                </span>
              )}
              {service !== '—' && (
                <span>
                  <span className="text-gray-400">service</span> {service}
                </span>
              )}
            </div>
          )}
          {(errRate !== '' || dailyCost !== '') && (
            <div className="flex flex-wrap gap-x-2">
              {errRate !== '' && (
                <span>
                  <span className="text-gray-400">error rate</span> {errRate}
                </span>
              )}
              {dailyCost !== '' && (
                <span>
                  <span className="text-gray-400">daily $</span> {dailyCost}
                </span>
              )}
            </div>
          )}
          {recentList.length > 0 && (
            <div>
              <div className="mb-0.5 text-[10px] uppercase tracking-wide text-gray-400">recent errors</div>
              <ul className="list-inside list-disc space-y-0.5 text-gray-800">
                {recentList.slice(0, 8).map((line, i) => (
                  <li key={`${i}-${String(line).slice(0, 24)}`} className="break-words">
                    {String(line)}
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      )}

      {!node && (
        <div className="border-b border-gray-200 px-3 py-3 text-sm text-gray-500">
          No node exists for this id in the stack graph.
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
                {deployments.length === 0 ? (
                  <div className="space-y-2 p-3 text-sm text-gray-600">
                    <p className="font-medium text-gray-800">No deployment data</p>
                    {(node.provider ?? '').toLowerCase() === 'custom' ? (
                      <p>
                        Custom nodes only expose health from the <span className="font-mono text-xs">status:</span> shell
                        exit code right now. Last deploy and recent errors are not filled for this provider yet.
                      </p>
                    ) : (
                      <p>
                        Deployable providers (Vercel, Render, …) still return a placeholder from{' '}
                        <span className="font-mono text-xs">perch status</span> in this build, so{' '}
                        <span className="font-mono text-xs">last_deploy</span> and errors do not appear here.
                      </p>
                    )}
                    <p className="text-xs text-gray-500">
                      In <span className="font-mono">examples/scenarios/full-stack</span>, switch the header environment
                      to <span className="font-semibold text-gray-700">dev</span> for demo{' '}
                      <span className="font-mono text-[11px]">custom</span> nodes. Production{' '}
                      <span className="font-mono text-[11px]">web</span> is Vercel YAML only (no shell hooks).
                    </p>
                  </div>
                ) : (
                  deployments.map((d, i) => <DeployRow key={`${node.id}-row-${i}`} deployment={d} />)
                )}
              </div>
            )}

            {tab === 'logs' && (
              <div className="p-3 text-sm text-gray-700">
                <p className="text-xs text-gray-500">
                  Commands from perch.yaml (same as the TUI). Run locally in your shell; nothing is executed from the
                  browser.
                </p>
                {hasStatusCommand && (
                  <CopyBlock
                    label="Health check (status:)"
                    text={statusCmd}
                    copied={copyWhich === 'status'}
                    onCopy={() => void copyText('status', statusCmd)}
                  />
                )}
                {hasLogsCommand && (
                  <CopyBlock
                    label="Tail logs (logs:)"
                    text={logsCmd}
                    copied={copyWhich === 'logs'}
                    onCopy={() => void copyText('logs', logsCmd)}
                  />
                )}
                {!hasStatusCommand && !hasLogsCommand && (
                  <div className="mt-2 space-y-2 text-sm text-gray-600">
                    <p>No <span className="font-mono text-xs">status:</span> or <span className="font-mono text-xs">logs:</span> shell lines for this node in perch.yaml.</p>
                    {(node.provider ?? '').toLowerCase() !== 'custom' && (
                      <p className="text-xs text-gray-500">
                        Platform nodes use provider APIs, not YAML shell hooks. Try environment{' '}
                        <span className="font-semibold text-gray-700">dev</span> in this example stack to see sample{' '}
                        <span className="font-mono text-[11px]">status:</span> / <span className="font-mono text-[11px]">logs:</span>{' '}
                        on <span className="font-mono text-[11px]">web</span>, <span className="font-mono text-[11px]">api</span>, etc.
                      </p>
                    )}
                  </div>
                )}
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
