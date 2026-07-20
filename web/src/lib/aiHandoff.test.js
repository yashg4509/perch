import { describe, expect, it, vi } from 'vitest'
import { buildErrorPrompt, isErroredNode, openInAIWithFallback } from './aiHandoff.js'

describe('isErroredNode', () => {
  it('treats degraded/down status as errored', () => {
    expect(isErroredNode({ status: 'degraded' })).toBe(true)
    expect(isErroredNode({ status: 'down' })).toBe(true)
    expect(isErroredNode({ status: 'healthy' })).toBe(false)
  })

  it('treats non-empty recentErrors as errored', () => {
    expect(isErroredNode({ status: 'healthy', recentErrors: ['db timeout'] })).toBe(true)
  })
})

describe('buildErrorPrompt', () => {
  it('includes node context and simulated log payload', () => {
    const prompt = buildErrorPrompt({
      node: {
        id: 'api',
        label: 'api',
        provider: 'custom',
        recentErrors: ['dial tcp: connection refused'],
      },
      environment: 'dev',
      logsResult: {
        stderrLines: ['health check failed'],
        stdoutLines: ['starting up...'],
      },
    })
    expect(prompt).toContain('Node: api')
    expect(prompt).toContain('Environment: dev')
    expect(prompt).toContain('dial tcp: connection refused')
    expect(prompt).toContain('health check failed')
    expect(prompt).toContain('starting up...')
    expect(prompt).toContain('most likely root cause')
  })
})

describe('openInAIWithFallback', () => {
  it('uses deep link when open succeeds', async () => {
    const openWindow = vi.fn(() => ({}))
    const copyText = vi.fn()
    const res = await openInAIWithFallback({
      tool: 'cursor',
      prompt: 'hello',
      openWindow,
      copyText,
    })
    expect(res.mode).toBe('deep_link')
    expect(openWindow).toHaveBeenCalledOnce()
    expect(copyText).not.toHaveBeenCalled()
  })

  it('copies prompt when deep link cannot open', async () => {
    const openWindow = vi.fn(() => null)
    const copyText = vi.fn(async () => {})
    const res = await openInAIWithFallback({
      tool: 'codex',
      prompt: 'hello',
      openWindow,
      copyText,
    })
    expect(res.mode).toBe('copied')
    expect(copyText).toHaveBeenCalledWith('hello')
  })
})
