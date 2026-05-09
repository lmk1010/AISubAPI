<template>
  <div class="blog-shell">
    <MarketingHeader />

    <main class="blog-main">
      <header class="page-hero">
        <span class="page-badge">Blog</span>
        <h1 class="page-title">博客 & 公告</h1>
        <p class="page-desc">产品动态、使用技巧和平台公告</p>
      </header>

      <div class="blog-grid">
        <article v-for="post in posts" :key="post.id" class="blog-card" @click="openPost(post)">
          <div class="blog-card__tag">{{ post.tag }}</div>
          <h2 class="blog-card__title">{{ post.title }}</h2>
          <p class="blog-card__excerpt">{{ post.excerpt }}</p>
          <div class="blog-card__meta">
            <time>{{ post.date }}</time>
            <span class="blog-card__read">阅读 →</span>
          </div>
        </article>
      </div>

      <!-- Post Detail Modal -->
      <Teleport to="body">
        <Transition name="modal">
          <div v-if="activePost" class="modal-overlay" @click.self="activePost = null">
            <div class="modal-content">
              <button class="modal-close" @click="activePost = null">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="w-5 h-5"><path stroke-linecap="round" d="M18 6L6 18M6 6l12 12" /></svg>
              </button>
              <div class="modal-tag">{{ activePost.tag }}</div>
              <h1 class="modal-title">{{ activePost.title }}</h1>
              <time class="modal-date">{{ activePost.date }}</time>
              <div class="modal-body" v-html="activePost.content"></div>
            </div>
          </div>
        </Transition>
      </Teleport>
    </main>

    <MarketingFooter />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import MarketingHeader from '@/components/marketing/MarketingHeader.vue'
import MarketingFooter from '@/components/marketing/MarketingFooter.vue'

interface BlogPost {
  id: number
  tag: string
  title: string
  excerpt: string
  date: string
  content: string
}

const activePost = ref<BlogPost | null>(null)

function openPost(post: BlogPost) {
  activePost.value = post
}

const posts = ref<BlogPost[]>([
  {
    id: 1,
    tag: '产品更新',
    title: '支持 GPT-5.5 和 Claude Opus 4.7 最新模型',
    excerpt: '我们已第一时间上线 OpenAI GPT-5.5 系列（含 Pro 版）和 Anthropic Claude Opus 4.7 旗舰模型，API 调用方式不变。',
    date: '2026-05-09',
    content: `<p>我们很高兴地宣布，以下最新模型已在平台上线并可立即使用：</p>
<h3>OpenAI 新模型</h3>
<ul>
<li><strong>gpt-5.5</strong> — OpenAI 最新旗舰，推理和代码能力大幅提升</li>
<li><strong>gpt-5.5-pro</strong> — 增强版，支持更大上下文和更强推理</li>
<li><strong>gpt-5.1-codex-max</strong> — 面向编程的专用模型，Codex CLI 推荐</li>
</ul>
<h3>Anthropic 新模型</h3>
<ul>
<li><strong>claude-opus-4.7</strong> — Anthropic 最新旗舰，长上下文推理能力行业领先</li>
<li><strong>claude-sonnet-4.5</strong> — 性价比之选，适合日常编码和写作</li>
</ul>
<p>所有模型通过标准 OpenAI API 格式调用，无需更改代码。查看 <a href="/pricing">定价页面</a> 了解各模型的详细计费。</p>`
  },
  {
    id: 2,
    tag: '教程',
    title: '使用 EasyAIConfig 一键配置 Codex CLI',
    excerpt: 'OpenAI Codex CLI 是强大的终端编程助手，通过 EasyAIConfig 可以一键配置中转 API 接入。',
    date: '2026-05-08',
    content: `<p>OpenAI Codex CLI 是一款基于终端的 AI 编程助手，可以直接在命令行中使用 GPT 模型辅助编码。</p>
<h3>快速接入步骤</h3>
<ol>
<li>下载 <a href="https://github.com/lmk1010/EasyAIConfig" target="_blank">EasyAIConfig</a> 桌面端</li>
<li>输入本站 API Base URL 和你的 API Key</li>
<li>点击「保存并启动」，自动写入 <code>~/.codex/config.toml</code></li>
</ol>
<h3>手动配置</h3>
<p>也可以直接编辑 <code>~/.codex/config.toml</code>：</p>
<pre><code>[provider]
name = "openai"
base_url = "你的Base URL/v1"

[auth]
api_key = "sk-your-api-key"</code></pre>
<p>配置完成后，在终端输入 <code>codex</code> 即可开始使用。</p>`
  },
  {
    id: 3,
    tag: '公告',
    title: 'Gemini 3.1 Pro 和 Grok 4 已上线',
    excerpt: 'Google Gemini 3.1 Pro 以极具竞争力的价格提供旗舰级推理能力，xAI Grok 4 同步上线。',
    date: '2026-05-06',
    content: `<p>本次更新带来两个重量级模型：</p>
<h3>Google Gemini 3.1 Pro</h3>
<p>Gemini 3.1 Pro 在 SWE-bench 评测中达到 80.6% 的得分，同时定价仅为 Opus 级别的 1/4，是目前性价比最高的旗舰编码模型之一。</p>
<h3>xAI Grok 4</h3>
<p>xAI 的最新模型 Grok 4 在推理基准测试中表现优异，特别在 GPQA Diamond 上取得了 93.6 的高分。</p>
<p>两个模型均已通过标准 API 格式接入，可在 <a href="/pricing">定价页面</a> 查看详细价格。</p>`
  },
  {
    id: 4,
    tag: '教程',
    title: 'Claude Code 配置中转 API 指南',
    excerpt: 'Anthropic 的 Claude Code 是强大的 AI 编程工具，以下是如何将它接入本站 API 的教程。',
    date: '2026-05-04',
    content: `<p>Claude Code 是 Anthropic 推出的 AI 编程助手，支持深度代码理解和多文件编辑。</p>
<h3>配置步骤</h3>
<p>编辑 <code>~/.claude/settings.json</code>：</p>
<pre><code>{
  "api_base": "你的Base URL",
  "api_key": "sk-your-api-key"
}</code></pre>
<p>保存后重启 Claude Code 即可。推荐使用 <a href="https://github.com/lmk1010/EasyAIConfig" target="_blank">EasyAIConfig</a> 一键配置，支持 Codex 和 Claude Code 双工具。</p>`
  },
  {
    id: 5,
    tag: '产品更新',
    title: 'Responses API 兼容支持',
    excerpt: 'OpenAI 新版 Responses API (/v1/responses) 现已完全兼容，Codex CLI 等工具可直接使用。',
    date: '2026-05-01',
    content: `<p>我们现已支持 OpenAI 新推出的 Responses API 格式：</p>
<ul>
<li>端点：<code>POST /v1/responses</code></li>
<li>完全兼容 Codex CLI、ChatGPT Plugins 等使用新格式的工具</li>
<li>支持流式输出</li>
</ul>
<p>现有 <code>/v1/chat/completions</code> 端点继续正常工作，两种格式并行支持。</p>`
  },
  {
    id: 6,
    tag: '公告',
    title: '平台正式上线',
    excerpt: '欢迎使用我们的 AI API 聚合平台！统一接口接入所有主流大模型，按量计费，简单易用。',
    date: '2026-04-28',
    content: `<p>我们的 AI API 聚合平台正式上线！</p>
<h3>平台特色</h3>
<ul>
<li><strong>统一接口</strong> — OpenAI API 兼容格式，一个 Key 调用所有模型</li>
<li><strong>多模型支持</strong> — OpenAI、Anthropic、Google、xAI 等主流模型全覆盖</li>
<li><strong>按量计费</strong> — 透明定价，按 Token 计费，没有月费</li>
<li><strong>工具兼容</strong> — 原生支持 Codex CLI、Claude Code、Cursor 等 AI 编程工具</li>
</ul>
<p>立即 <a href="/register">注册账号</a> 开始使用吧！</p>`
  }
])
</script>

<style scoped>
.blog-shell {
  position: relative;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f1f3fb url('/light-bg.png') no-repeat center center / cover;
  color: #0f172a;
}

.blog-main {
  flex: 1;
  position: relative;
  z-index: 1;
  max-width: 1080px;
  margin: 0 auto;
  width: 100%;
  padding: 40px 24px 80px;
}

.page-hero {
  text-align: center;
  margin-bottom: 48px;
}

.page-badge {
  display: inline-flex;
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
}

/* Blog Grid */
.blog-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 20px;
}

.blog-card {
  background: rgba(255, 255, 255, 0.72);
  backdrop-filter: blur(16px) saturate(180%);
  -webkit-backdrop-filter: blur(16px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.8);
  border-radius: 16px;
  padding: 24px;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
}
.blog-card:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.08);
  border-color: rgba(147, 197, 253, 0.5);
}

.blog-card__tag {
  display: inline-block;
  width: fit-content;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(219, 234, 254, 0.7);
  color: #1d4ed8;
  margin-bottom: 14px;
}

.blog-card__title {
  font-size: 18px;
  font-weight: 700;
  line-height: 1.4;
  margin-bottom: 10px;
  letter-spacing: -0.01em;
}

.blog-card__excerpt {
  font-size: 14px;
  color: #475569;
  line-height: 1.65;
  flex: 1;
  margin-bottom: 16px;
}

.blog-card__meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 13px;
  color: #94a3b8;
}

.blog-card__read {
  color: #2563eb;
  font-weight: 500;
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.modal-content {
  position: relative;
  background: #fff;
  border-radius: 20px;
  max-width: 720px;
  width: 100%;
  max-height: 85vh;
  overflow-y: auto;
  padding: 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.modal-close {
  position: absolute;
  top: 16px;
  right: 16px;
  width: 36px;
  height: 36px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #64748b;
  background: rgba(241, 245, 249, 0.8);
  transition: all 0.15s ease;
}
.modal-close:hover {
  background: #e2e8f0;
  color: #0f172a;
}

.modal-tag {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(219, 234, 254, 0.7);
  color: #1d4ed8;
  margin-bottom: 12px;
}

.modal-title {
  font-size: 26px;
  font-weight: 800;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
}

.modal-date {
  font-size: 13px;
  color: #94a3b8;
  display: block;
  margin-bottom: 24px;
}

.modal-body {
  font-size: 15px;
  line-height: 1.8;
  color: #334155;
}

.modal-body h3 {
  font-size: 17px;
  font-weight: 700;
  margin: 20px 0 10px;
  color: #0f172a;
}

.modal-body p {
  margin-bottom: 12px;
}

.modal-body ul, .modal-body ol {
  padding-left: 20px;
  margin-bottom: 12px;
}

.modal-body li {
  margin-bottom: 6px;
}

.modal-body code {
  background: #f1f5f9;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
}

.modal-body pre {
  background: #0f172a;
  border-radius: 10px;
  padding: 16px;
  overflow-x: auto;
  margin-bottom: 12px;
}

.modal-body pre code {
  background: none;
  padding: 0;
  color: #e2e8f0;
  font-size: 13px;
  white-space: pre;
}

.modal-body a {
  color: #2563eb;
  text-decoration: none;
  font-weight: 500;
}
.modal-body a:hover {
  text-decoration: underline;
}

/* Transition */
.modal-enter-active, .modal-leave-active {
  transition: opacity 0.2s ease;
}
.modal-enter-active .modal-content, .modal-leave-active .modal-content {
  transition: transform 0.2s ease;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
}
.modal-enter-from .modal-content {
  transform: translateY(20px) scale(0.97);
}
.modal-leave-to .modal-content {
  transform: translateY(10px) scale(0.99);
}

@media (max-width: 640px) {
  .page-title { font-size: 28px; }
  .blog-grid { grid-template-columns: 1fr; }
  .modal-content { padding: 24px; }
  .modal-title { font-size: 22px; }
}
</style>

<style>
html.dark .blog-shell { background: #0a0f1a url('/dark-dashboard-bg.png') no-repeat center center / cover; color: #e2e8f0; }
html.dark .page-title { color: #f8fafc; }
html.dark .page-desc { color: #94a3b8; }
html.dark .page-badge { background: rgba(30, 64, 175, 0.25); color: #93c5fd; border-color: rgba(59, 130, 246, 0.3); }
html.dark .blog-card {
  background: rgba(15, 23, 42, 0.7);
  border-color: rgba(71, 85, 105, 0.5);
}
html.dark .blog-card:hover {
  border-color: rgba(59, 130, 246, 0.4);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.3);
}
html.dark .blog-card__tag { background: rgba(30, 64, 175, 0.3); color: #93c5fd; }
html.dark .blog-card__title { color: #f8fafc; }
html.dark .blog-card__excerpt { color: #94a3b8; }
html.dark .blog-card__read { color: #60a5fa; }
html.dark .modal-overlay { background: rgba(0, 0, 0, 0.7); }
html.dark .modal-content { background: #1e293b; }
html.dark .modal-close { background: rgba(51, 65, 85, 0.6); color: #94a3b8; }
html.dark .modal-close:hover { background: #475569; color: #f8fafc; }
html.dark .modal-tag { background: rgba(30, 64, 175, 0.3); color: #93c5fd; }
html.dark .modal-title { color: #f8fafc; }
html.dark .modal-date { color: #64748b; }
html.dark .modal-body { color: #cbd5e1; }
html.dark .modal-body h3 { color: #f8fafc; }
html.dark .modal-body code { background: rgba(51, 65, 85, 0.6); color: #e2e8f0; }
html.dark .modal-body a { color: #60a5fa; }
</style>
