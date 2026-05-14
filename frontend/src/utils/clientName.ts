export function formatClientName(userAgent: string | null | undefined): string {
  const value = (userAgent || '').trim()
  if (!value) return ''
  const lower = value.toLowerCase()

  if (lower.includes('openclaw')) return 'OpenClaw'
  if (lower.includes('claude-cli') || lower.includes('claude-code') || lower.includes('claude code')) return 'Claude Code'
  if (lower.includes('codex_vscode')) return 'Codex VS Code'
  if (lower.includes('codex_cli_rs') || lower.includes('codex-cli')) return 'Codex CLI'
  if (lower.includes('codex_app') || lower.includes('codex desktop') || lower.includes('codex_chatgpt_desktop')) return 'Codex Desktop'
  if (lower.includes('cursor')) return 'Cursor'
  if (lower.includes('trae')) return 'Trae'
  if (lower.includes('kiro')) return 'Kiro'
  if (lower.includes('windsurf')) return 'Windsurf'
  if (lower.includes('cline')) return 'Cline'
  if (lower.includes('roo-code') || lower.includes('roo code')) return 'Roo Code'
  if (lower.includes('opencode')) return 'OpenCode'
  if (lower.includes('openai-node') || lower.includes('openai-python') || lower.includes('openai-java') || lower.includes('openai-go') || lower.includes('openai/')) return 'OpenAI SDK'
  if (lower.includes('anthropic-sdk') || lower.includes('anthropic-python') || lower.includes('anthropic-typescript')) return 'Anthropic SDK'
  if (lower.includes('curl/')) return 'curl'
  if (lower.includes('postmanruntime')) return 'Postman'
  if (lower.includes('insomnia')) return 'Insomnia'
  if (lower.includes('python-requests')) return 'Python requests'
  if (lower.includes('go-http-client')) return 'Go HTTP client'
  if (lower.includes('okhttp')) return 'OkHttp'
  if (lower.includes('axios') || lower.includes('undici') || lower.includes('node-fetch')) return 'Node.js client'
  if (lower.includes('edg/')) return 'Edge'
  if (lower.includes('chrome/')) return 'Chrome'
  if (lower.includes('firefox/')) return 'Firefox'
  if (lower.includes('safari/') && !lower.includes('chrome/')) return 'Safari'

  const product = value.match(/^([A-Za-z][A-Za-z0-9._ -]{1,40})(?:\/|\s|$)/)?.[1]?.trim()
  return product || value
}
