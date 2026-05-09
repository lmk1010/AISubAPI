<template>
  <div class="docs-shell">
    <MarketingHeader />

    <div class="docs-layout">
      <!-- Left sidebar nav -->
      <aside class="docs-sidebar">
        <nav class="docs-nav">
          <a
            v-for="sec in sections"
            :key="sec.id"
            :href="'#' + sec.id"
            class="docs-nav-item"
            :class="{ active: activeSection === sec.id }"
            @click.prevent="scrollTo(sec.id)"
          >
            <span class="docs-nav-dot" />
            {{ sec.label }}
          </a>
        </nav>
      </aside>

      <!-- Right content -->
      <main class="docs-main">
        <header class="page-hero">
          <span class="page-badge">{{ t('home.nav.docs') }}</span>
          <h1 class="page-title">API 文档</h1>
          <p class="page-desc">快速接入指南，帮助你在几分钟内开始使用 {{ siteName }} API</p>
        </header>

        <!-- Quick Start -->
        <section id="quick-start" class="doc-section">
          <h2 class="section-title">快速开始</h2>
          <div class="doc-card">
            <p class="doc-text">{{ siteName }} 完全兼容 OpenAI API 格式。你只需将请求地址替换为我们的 Base URL，并使用你的 API Key 即可。</p>
            <div class="steps">
              <div class="step">
                <span class="step-num">1</span>
                <div class="step-content">
                  <h4>获取 API Key</h4>
                  <p>登录后在 <router-link to="/keys" class="doc-link">密钥管理</router-link> 页面创建你的 API Key。</p>
                </div>
              </div>
              <div class="step">
                <span class="step-num">2</span>
                <div class="step-content">
                  <h4>配置 Base URL</h4>
                  <p>将 API 请求地址设置为：</p>
                  <div class="code-block">
                    <code>{{ apiBaseUrl || currentOrigin }}</code>
                    <button class="copy-btn" @click="copyText(apiBaseUrl || currentOrigin)" :title="copiedKey === 'base' ? '已复制' : '复制'">
                      <svg v-if="copiedKey !== 'base'" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-4 h-4"><rect x="9" y="9" width="13" height="13" rx="2" ry="2" /><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1" /></svg>
                      <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-4 h-4"><polyline points="20 6 9 17 4 12" /></svg>
                    </button>
                  </div>
                </div>
              </div>
              <div class="step">
                <span class="step-num">3</span>
                <div class="step-content">
                  <h4>发送请求</h4>
                  <p>使用标准 OpenAI SDK 或 HTTP 请求即可调用。</p>
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- Authentication -->
        <section id="auth" class="doc-section">
          <h2 class="section-title">认证方式</h2>
          <div class="doc-card">
            <p class="doc-text">所有 API 请求必须在 Header 中携带你的 API Key：</p>
            <div class="code-block">
              <code>Authorization: Bearer sk-your-api-key</code>
            </div>
            <div class="note">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="note-icon"><circle cx="12" cy="12" r="10" /><line x1="12" y1="16" x2="12" y2="12" /><line x1="12" y1="8" x2="12.01" y2="8" /></svg>
              <p>请妥善保管你的 API Key，不要在客户端代码或公开仓库中暴露。</p>
            </div>
          </div>
        </section>

        <!-- Code Examples -->
        <section id="examples" class="doc-section">
          <h2 class="section-title">代码示例</h2>
          <div class="tabs">
            <button v-for="tab in codeTabs" :key="tab.id" class="tab-btn" :class="{ active: activeTab === tab.id }" @click="activeTab = tab.id">{{ tab.label }}</button>
          </div>
          <div class="doc-card code-card">
            <pre class="code-pre"><code>{{ codeExamples[activeTab] }}</code></pre>
          </div>
        </section>

        <!-- Supported Models -->
        <section id="models" class="doc-section">
          <h2 class="section-title">支持的模型</h2>
          <div class="doc-card">
            <p class="doc-text">我们支持多个平台的最新旗舰与经济型模型，包括但不限于：</p>
            <div class="model-grid">
              <div class="model-group">
                <h4 class="model-platform">OpenAI</h4>
                <ul class="model-list">
                  <li>gpt-5.5 / gpt-5.5-pro</li>
                  <li>gpt-5.2 / gpt-5.2-pro</li>
                  <li>gpt-5 / gpt-5-mini</li>
                  <li>gpt-5.1-codex-max</li>
                  <li>o4-mini / o3 / o3-pro</li>
                  <li>gpt-4o / gpt-4o-mini</li>
                  <li>dall-e-3 / gpt-image-1</li>
                </ul>
              </div>
              <div class="model-group">
                <h4 class="model-platform">Anthropic</h4>
                <ul class="model-list">
                  <li>claude-opus-4.7</li>
                  <li>claude-opus-4.6</li>
                  <li>claude-sonnet-4.5</li>
                  <li>claude-sonnet-4</li>
                  <li>claude-haiku-4.5</li>
                  <li>claude-sonnet-4-20250514</li>
                </ul>
              </div>
              <div class="model-group">
                <h4 class="model-platform">Google</h4>
                <ul class="model-list">
                  <li>gemini-3-pro / gemini-3.1-pro</li>
                  <li>gemini-2.5-pro</li>
                  <li>gemini-2.5-flash</li>
                  <li>gemini-2.0-flash</li>
                </ul>
              </div>
              <div class="model-group">
                <h4 class="model-platform">xAI / 其他</h4>
                <ul class="model-list">
                  <li>grok-4</li>
                  <li>deepseek-v3 / deepseek-r1</li>
                  <li>kimi-k2</li>
                </ul>
              </div>
            </div>
            <p class="doc-hint">完整模型列表与实时价格请查看 <router-link to="/pricing" class="doc-link">定价页面</router-link>。</p>
          </div>
        </section>

        <!-- Endpoints -->
        <section id="endpoints" class="doc-section">
          <h2 class="section-title">API 端点</h2>
          <div class="doc-card">
            <div class="endpoint-table">
              <div class="endpoint-row endpoint-header">
                <span class="endpoint-method">Method</span>
                <span class="endpoint-path">Path</span>
                <span class="endpoint-desc">Description</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge post">POST</span></span>
                <span class="endpoint-path"><code>/v1/chat/completions</code></span>
                <span class="endpoint-desc">对话补全（支持流式）</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge post">POST</span></span>
                <span class="endpoint-path"><code>/v1/responses</code></span>
                <span class="endpoint-desc">Responses API（OpenAI 新格式）</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge post">POST</span></span>
                <span class="endpoint-path"><code>/v1/images/generations</code></span>
                <span class="endpoint-desc">图片生成</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge post">POST</span></span>
                <span class="endpoint-path"><code>/v1/embeddings</code></span>
                <span class="endpoint-desc">文本向量化</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge post">POST</span></span>
                <span class="endpoint-path"><code>/v1/audio/transcriptions</code></span>
                <span class="endpoint-desc">语音转文字</span>
              </div>
              <div class="endpoint-row">
                <span class="endpoint-method"><span class="method-badge get">GET</span></span>
                <span class="endpoint-path"><code>/v1/models</code></span>
                <span class="endpoint-desc">可用模型列表</span>
              </div>
            </div>
          </div>
        </section>

        <!-- Codex / Claude Code -->
        <section id="codex" class="doc-section">
          <h2 class="section-title">Codex / Claude Code 接入</h2>
          <div class="doc-card">
            <p class="doc-text">本站 API 可直接用于 OpenAI Codex CLI 和 Claude Code 等编程助手工具。推荐使用 <a href="https://github.com/lmk1010/EasyAIConfig" target="_blank" rel="noopener noreferrer" class="doc-link">EasyAIConfig</a> 一键配置。</p>

            <h4 class="sub-heading">Codex CLI 手动配置</h4>
            <p class="doc-text-sm">编辑 <code>~/.codex/config.toml</code>：</p>
            <div class="code-block code-block--multi">
<pre><code>[provider]
name = "openai"
base_url = "{{ apiBaseUrl || currentOrigin }}/v1"

[auth]
api_key = "sk-your-api-key"</code></pre>
            </div>

            <h4 class="sub-heading">Claude Code 配置</h4>
            <p class="doc-text-sm">编辑 <code>~/.claude/settings.json</code>，将 API 端点指向本站：</p>
            <div class="code-block code-block--multi">
<pre><code>{
  "api_base": "{{ apiBaseUrl || currentOrigin }}",
  "api_key": "sk-your-api-key"
}</code></pre>
            </div>

            <div class="note note--tip">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="note-icon"><path d="M12 2L2 7l10 5 10-5-10-5z" /><path d="M2 17l10 5 10-5" /><path d="M2 12l10 5 10-5" /></svg>
              <p>也可使用 <a href="https://github.com/lmk1010/EasyAIConfig" target="_blank" rel="noopener noreferrer" class="doc-link">EasyAIConfig</a> 桌面端一键完成上述配置，支持 Codex 官方登录 + 中转 API Key 双路径。</p>
            </div>
          </div>
        </section>

        <!-- FAQ -->
        <section id="faq" class="doc-section">
          <h2 class="section-title">常见问题</h2>
          <div class="faq-list">
            <details class="faq-item">
              <summary class="faq-question">支持流式输出（Streaming）吗？</summary>
              <p class="faq-answer">支持。在请求体中设置 <code>"stream": true</code> 即可获得 SSE 流式响应。</p>
            </details>
            <details class="faq-item">
              <summary class="faq-question">请求频率有限制吗？</summary>
              <p class="faq-answer">默认每分钟 60 次请求，突发上限 10 次。如需更高配额，请联系管理员。</p>
            </details>
            <details class="faq-item">
              <summary class="faq-question">如何查看用量和余额？</summary>
              <p class="faq-answer">登录后在 <router-link to="/dashboard" class="doc-link">控制面板</router-link> 可查看实时用量统计和余额信息。</p>
            </details>
            <details class="faq-item">
              <summary class="faq-question">兼容哪些 SDK 和工具？</summary>
              <p class="faq-answer">兼容所有支持自定义 Base URL 的 OpenAI SDK（Python / Node.js / Go），以及 Codex CLI、Claude Code、Cursor、Continue 等 AI 编程工具。</p>
            </details>
            <details class="faq-item">
              <summary class="faq-question">支持 Responses API 吗？</summary>
              <p class="faq-answer">支持。OpenAI 新版 Responses API（<code>/v1/responses</code>）已兼容，可在 Codex 等工具中直接使用。</p>
            </details>
          </div>
        </section>
      </main>
    </div>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import MarketingHeader from '@/components/marketing/MarketingHeader.vue'
import MarketingFooter from '@/components/marketing/MarketingFooter.vue'

const { t } = useI18n()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const apiBaseUrl = computed(() => appStore.cachedPublicSettings?.api_base_url || '')
const currentOrigin = ref(window.location.origin)

const sections = [
  { id: 'quick-start', label: '快速开始' },
  { id: 'auth', label: '认证方式' },
  { id: 'examples', label: '代码示例' },
  { id: 'models', label: '支持的模型' },
  { id: 'endpoints', label: 'API 端点' },
  { id: 'codex', label: 'Codex / Claude Code' },
  { id: 'faq', label: '常见问题' },
]

const activeSection = ref('quick-start')

function scrollTo(id: string) {
  const el = document.getElementById(id)
  if (el) {
    el.scrollIntoView({ behavior: 'smooth', block: 'start' })
    activeSection.value = id
  }
}

let scrollObserver: IntersectionObserver | null = null

onMounted(() => {
  appStore.fetchPublicSettings()
  scrollObserver = new IntersectionObserver(
    (entries) => {
      for (const entry of entries) {
        if (entry.isIntersecting) {
          activeSection.value = entry.target.id
        }
      }
    },
    { rootMargin: '-80px 0px -60% 0px', threshold: 0.1 }
  )
  sections.forEach((sec) => {
    const el = document.getElementById(sec.id)
    if (el) scrollObserver!.observe(el)
  })
})

onUnmounted(() => {
  scrollObserver?.disconnect()
})

const copiedKey = ref('')
function copyText(text: string, key = 'base') {
  navigator.clipboard.writeText(text)
  copiedKey.value = key
  setTimeout(() => { copiedKey.value = '' }, 2000)
}

const activeTab = ref<'python' | 'curl' | 'node'>('python')
const codeTabs: { id: 'python' | 'curl' | 'node'; label: string }[] = [
  { id: 'python', label: 'Python' },
  { id: 'curl', label: 'cURL' },
  { id: 'node', label: 'Node.js' },
]

const codeExamples = computed(() => {
  const base = apiBaseUrl.value || currentOrigin.value
  return {
    python: `from openai import OpenAI

client = OpenAI(
    api_key="sk-your-api-key",
    base_url="${base}/v1"
)

response = client.chat.completions.create(
    model="gpt-5",
    messages=[
        {"role": "user", "content": "Hello!"}
    ]
)
print(response.choices[0].message.content)`,
    curl: `curl ${base}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-api-key" \\
  -d '{
    "model": "gpt-5",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'`,
    node: `import OpenAI from 'openai';

const client = new OpenAI({
  apiKey: 'sk-your-api-key',
  baseURL: '${base}/v1',
});

const response = await client.chat.completions.create({
  model: 'gpt-5',
  messages: [
    { role: 'user', content: 'Hello!' }
  ],
});
console.log(response.choices[0].message.content);`,
  }
})
</script>

<style scoped>
.docs-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f1f3fb url('/light-bg.png') no-repeat center center / cover;
  color: #0f172a;
}

/* Left-right layout */
.docs-layout {
  flex: 1;
  display: flex;
  max-width: 1280px;
  margin: 0 auto;
  width: 100%;
  padding: 0 32px;
  position: relative;
  z-index: 1;
}

.docs-sidebar {
  display: none;
  width: 220px;
  flex-shrink: 0;
  padding: 40px 0;
  position: sticky;
  top: 88px;
  height: fit-content;
  max-height: calc(100vh - 120px);
  overflow-y: auto;
}

@media (min-width: 1024px) {
  .docs-sidebar { display: block; }
}

.docs-nav {
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-left: 2px solid rgba(226, 232, 240, 0.6);
  padding-left: 0;
}

.docs-nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 16px;
  margin-left: -2px;
  border-left: 2px solid transparent;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  text-decoration: none;
  transition: all 0.15s ease;
}
.docs-nav-item:hover {
  color: #0f172a;
}
.docs-nav-item.active {
  color: #2563eb;
  border-left-color: #2563eb;
  font-weight: 600;
}

.docs-nav-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #cbd5e1;
  flex-shrink: 0;
  transition: background 0.15s ease;
}
.docs-nav-item.active .docs-nav-dot {
  background: #2563eb;
}

.docs-main {
  flex: 1;
  min-width: 0;
  padding: 40px 0 80px 48px;
}

@media (max-width: 1023px) {
  .docs-main { padding: 40px 0 80px; }
}

.page-hero {
  margin-bottom: 48px;
}

.page-badge {
  display: inline-flex;
  align-items: center;
  padding: 6px 16px;
  border-radius: 999px;
  background: rgba(219, 234, 254, 0.7);
  color: #1d4ed8;
  font-size: 13px;
  font-weight: 500;
  border: 1px solid rgba(191, 219, 254, 0.6);
  margin-bottom: 16px;
}

.page-title {
  font-size: 36px;
  font-weight: 800;
  letter-spacing: -0.03em;
  margin-bottom: 10px;
}

.page-desc {
  font-size: 16px;
  color: #475569;
  max-width: 520px;
}

.doc-section {
  margin-bottom: 48px;
  scroll-margin-top: 100px;
}

.section-title {
  font-size: 22px;
  font-weight: 700;
  margin-bottom: 16px;
  letter-spacing: -0.01em;
  padding-bottom: 10px;
  border-bottom: 1px solid rgba(226, 232, 240, 0.5);
}

.sub-heading {
  font-size: 15px;
  font-weight: 600;
  margin: 20px 0 6px;
}

.doc-card {
  background: rgba(255, 255, 255, 0.78);
  backdrop-filter: blur(18px) saturate(180%);
  -webkit-backdrop-filter: blur(18px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.85);
  border-radius: 16px;
  padding: 24px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.04);
}

.doc-text {
  font-size: 15px;
  color: #334155;
  line-height: 1.7;
  margin-bottom: 18px;
}

.doc-text-sm {
  font-size: 14px;
  color: #475569;
  line-height: 1.6;
  margin-bottom: 6px;
}

.doc-text-sm code {
  background: rgba(241, 245, 249, 0.9);
  padding: 1px 5px;
  border-radius: 4px;
  font-size: 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.doc-hint {
  font-size: 13px;
  color: #64748b;
  margin-top: 16px;
}

.doc-link {
  color: #2563eb;
  text-decoration: none;
  font-weight: 500;
}
.doc-link:hover {
  text-decoration: underline;
}

/* Steps */
.steps { display: flex; flex-direction: column; gap: 20px; }
.step { display: flex; gap: 16px; align-items: flex-start; }
.step-num {
  flex-shrink: 0; width: 32px; height: 32px; border-radius: 50%;
  background: #2563eb; color: #fff; font-size: 14px; font-weight: 700;
  display: flex; align-items: center; justify-content: center;
}
.step-content h4 { font-size: 15px; font-weight: 600; margin-bottom: 4px; }
.step-content p { font-size: 14px; color: #475569; line-height: 1.6; }

/* Code blocks */
.code-block {
  display: flex; align-items: center; gap: 8px;
  background: #0f172a; border-radius: 10px;
  padding: 12px 16px; margin-top: 10px; overflow-x: auto;
}
.code-block--multi {
  display: block; padding: 0;
}
.code-block--multi pre {
  margin: 0; padding: 14px 16px;
}
.code-block code, .code-block--multi code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px; color: #e2e8f0; white-space: pre;
}
.copy-btn {
  flex-shrink: 0; display: flex; align-items: center; justify-content: center;
  width: 32px; height: 32px; border-radius: 8px;
  color: #94a3b8; background: transparent; transition: all 0.15s ease;
}
.copy-btn:hover { background: rgba(248, 250, 252, 0.1); color: #f8fafc; }

/* Tabs */
.tabs {
  display: flex; gap: 4px; margin-bottom: 12px;
  background: rgba(255, 255, 255, 0.5); border-radius: 10px;
  padding: 4px; border: 1px solid rgba(226, 232, 240, 0.6); width: fit-content;
}
.tab-btn {
  padding: 7px 16px; border-radius: 8px; font-size: 13px; font-weight: 500;
  color: #64748b; background: transparent; transition: all 0.15s ease;
}
.tab-btn:hover { color: #0f172a; }
.tab-btn.active {
  background: #fff; color: #0f172a; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}
.code-card { padding: 0; overflow: hidden; }
.code-pre {
  padding: 20px 24px; margin: 0; overflow-x: auto;
  background: #0f172a; border-radius: 16px;
}
.code-pre code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px; line-height: 1.7; color: #e2e8f0; white-space: pre;
}

/* Models */
.model-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 20px; }
.model-platform { font-size: 14px; font-weight: 600; margin-bottom: 8px; color: #0f172a; }
.model-list { list-style: none; padding: 0; margin: 0; }
.model-list li {
  font-size: 13px; color: #475569; padding: 3px 0;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

/* Endpoints */
.endpoint-table { display: flex; flex-direction: column; }
.endpoint-row {
  display: grid; grid-template-columns: 80px 1fr 1fr; gap: 12px;
  padding: 12px 0; border-bottom: 1px solid rgba(226, 232, 240, 0.6); align-items: center;
}
.endpoint-row:last-child { border-bottom: none; }
.endpoint-header { font-size: 12px; font-weight: 600; color: #64748b; text-transform: uppercase; letter-spacing: 0.04em; }
.endpoint-path code {
  font-size: 13px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; color: #0f172a;
}
.endpoint-desc { font-size: 13px; color: #475569; }
.method-badge { display: inline-block; padding: 2px 8px; border-radius: 6px; font-size: 11px; font-weight: 700; letter-spacing: 0.02em; }
.method-badge.post { background: rgba(219, 234, 254, 0.7); color: #1d4ed8; }
.method-badge.get { background: rgba(220, 252, 231, 0.7); color: #047857; }

/* Note */
.note {
  display: flex; gap: 10px; align-items: flex-start; margin-top: 16px;
  padding: 12px 16px; border-radius: 10px;
  background: rgba(254, 243, 199, 0.5); border: 1px solid rgba(253, 224, 71, 0.4);
}
.note--tip {
  background: rgba(219, 234, 254, 0.4); border-color: rgba(147, 197, 253, 0.4);
}
.note-icon { width: 18px; height: 18px; color: #d97706; flex-shrink: 0; margin-top: 1px; }
.note--tip .note-icon { color: #2563eb; }
.note p { font-size: 13px; color: #92400e; line-height: 1.5; }
.note--tip p { color: #1e40af; }

/* FAQ */
.faq-list { display: flex; flex-direction: column; gap: 8px; }
.faq-item {
  background: rgba(255, 255, 255, 0.78); backdrop-filter: blur(18px);
  border: 1px solid rgba(226, 232, 240, 0.7); border-radius: 12px; overflow: hidden;
}
.faq-question {
  padding: 16px 20px; font-size: 15px; font-weight: 600;
  cursor: pointer; list-style: none; display: flex; align-items: center;
}
.faq-question::before { content: '▸'; margin-right: 10px; transition: transform 0.2s ease; color: #94a3b8; }
.faq-item[open] .faq-question::before { transform: rotate(90deg); }
.faq-answer { padding: 0 20px 16px 30px; font-size: 14px; color: #475569; line-height: 1.7; }
.faq-answer code {
  background: rgba(241, 245, 249, 0.9); padding: 2px 6px; border-radius: 4px;
  font-size: 12px; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

@media (max-width: 640px) {
  .docs-layout { padding: 0 16px; }
  .page-title { font-size: 28px; }
  .endpoint-row { grid-template-columns: 1fr; gap: 4px; }
  .endpoint-header { display: none; }
  .model-grid { grid-template-columns: 1fr; }
}
</style>

<style>
html.dark .docs-shell { background: #0a0f1a url('/dark-dashboard-bg.png') no-repeat center center / cover; color: #e2e8f0; }
html.dark .docs-nav { border-left-color: rgba(71, 85, 105, 0.5); }
html.dark .docs-nav-item { color: #64748b; }
html.dark .docs-nav-item:hover { color: #e2e8f0; }
html.dark .docs-nav-item.active { color: #60a5fa; border-left-color: #60a5fa; }
html.dark .docs-nav-dot { background: #475569; }
html.dark .docs-nav-item.active .docs-nav-dot { background: #60a5fa; }
html.dark .page-title { color: #f8fafc; }
html.dark .page-desc { color: #94a3b8; }
html.dark .page-badge { background: rgba(30, 64, 175, 0.25); color: #93c5fd; border-color: rgba(59, 130, 246, 0.3); }
html.dark .section-title { color: #f8fafc; border-bottom-color: rgba(51, 65, 85, 0.5); }
html.dark .sub-heading { color: #e2e8f0; }
html.dark .doc-card { background: rgba(15, 23, 42, 0.7); border-color: rgba(71, 85, 105, 0.5); }
html.dark .doc-text { color: #cbd5e1; }
html.dark .doc-text-sm { color: #94a3b8; }
html.dark .doc-text-sm code { background: rgba(51, 65, 85, 0.6); color: #e2e8f0; }
html.dark .doc-hint { color: #64748b; }
html.dark .doc-link { color: #60a5fa; }
html.dark .step-content h4 { color: #f8fafc; }
html.dark .step-content p { color: #94a3b8; }
html.dark .model-platform { color: #f8fafc; }
html.dark .model-list li { color: #94a3b8; }
html.dark .endpoint-path code { color: #e2e8f0; }
html.dark .endpoint-desc { color: #94a3b8; }
html.dark .endpoint-row { border-bottom-color: rgba(51, 65, 85, 0.5); }
html.dark .tabs { background: rgba(15, 23, 42, 0.5); border-color: rgba(71, 85, 105, 0.5); }
html.dark .tab-btn { color: #94a3b8; }
html.dark .tab-btn:hover { color: #f8fafc; }
html.dark .tab-btn.active { background: rgba(51, 65, 85, 0.8); color: #f8fafc; }
html.dark .note { background: rgba(120, 53, 15, 0.2); border-color: rgba(217, 119, 6, 0.3); }
html.dark .note p { color: #fbbf24; }
html.dark .note-icon { color: #fbbf24; }
html.dark .note--tip { background: rgba(30, 64, 175, 0.15); border-color: rgba(59, 130, 246, 0.3); }
html.dark .note--tip p { color: #93c5fd; }
html.dark .note--tip .note-icon { color: #60a5fa; }
html.dark .faq-item { background: rgba(15, 23, 42, 0.7); border-color: rgba(71, 85, 105, 0.5); }
html.dark .faq-question { color: #f8fafc; }
html.dark .faq-answer { color: #94a3b8; }
html.dark .faq-answer code { background: rgba(51, 65, 85, 0.6); color: #e2e8f0; }
</style>
