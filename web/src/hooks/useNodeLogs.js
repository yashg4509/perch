import { useCallback, useEffect, useState } from 'react'

/**
 * On-demand fetch for /api/logs (provider nodes). Does not poll; call fetchLogs explicitly.
 * @param {string | undefined} nodeName
 * @param {string} env
 */
export function useNodeLogs(nodeName, env) {
  const [logs, setLogs] = useState(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState(null)

  useEffect(() => {
    setLogs(null)
    setError(null)
  }, [nodeName, env])

  const fetchLogs = useCallback(async () => {
    if (!nodeName) {
      return
    }
    setLoading(true)
    setError(null)
    try {
      const res = await fetch(
        `/api/logs?node=${encodeURIComponent(nodeName)}&env=${encodeURIComponent(env)}`,
        { headers: { Accept: 'application/json' } },
      )
      const text = await res.text()
      let data = null
      try {
        data = text !== '' ? JSON.parse(text) : null
      } catch {
        data = null
      }
      if (!res.ok) {
        const msg = data?.error ?? res.statusText ?? 'request failed'
        throw new Error(typeof msg === 'string' ? msg : JSON.stringify(msg))
      }
      setLogs(data)
    } catch (e) {
      setError(e instanceof Error ? e.message : String(e))
    } finally {
      setLoading(false)
    }
  }, [nodeName, env])

  return { logs, loading, error, fetchLogs }
}
