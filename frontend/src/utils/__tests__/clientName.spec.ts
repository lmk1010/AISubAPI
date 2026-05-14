import { describe, expect, it } from 'vitest'

import { formatClientName } from '../clientName'

describe('formatClientName', () => {
  it('recognizes common API clients from user agents', () => {
    expect(formatClientName('OpenClaw/1.0')).toBe('OpenClaw')
    expect(formatClientName('codex_cli_rs/0.98.0')).toBe('Codex CLI')
    expect(formatClientName('codex_vscode/1.0.0')).toBe('Codex VS Code')
    expect(formatClientName('claude-cli/2.1.92 (external, cli)')).toBe('Claude Code')
    expect(formatClientName('curl/8.7.1')).toBe('curl')
    expect(formatClientName('openai-node/4.0.0')).toBe('OpenAI SDK')
  })

  it('falls back to the first product token', () => {
    expect(formatClientName('CustomClient/2.3.4 something')).toBe('CustomClient')
    expect(formatClientName('')).toBe('')
  })
})
