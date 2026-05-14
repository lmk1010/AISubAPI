#!/usr/bin/env node
import { performance } from 'node:perf_hooks';

const baseURL = stripTrailingSlash(process.env.BASE_URL || 'https://api.jiminaishop.com');
const runs = positiveInt(process.env.RUNS, 3);
const timeoutMs = positiveInt(process.env.TIMEOUT_MS, 90000);
const model = process.env.MODEL || 'gpt-5.3-codex';
const prompt = process.env.PROMPT || '请只回复：OK';
const accounts = parseAccounts(process.env.ACCOUNTS || process.argv.slice(2).join(','));

if (accounts.length === 0) {
  fail('Set ACCOUNTS, for example: ACCOUNTS="2=fluxnode,3=pixel,4=mkcd" node tools/prod_openai_latency_probe.mjs');
}

const authHeaders = await resolveAdminAuthHeaders();

console.log(`# OpenAI account latency probe`);
console.log(`base_url: ${baseURL}`);
console.log(`model: ${model}`);
console.log(`runs_per_account: ${runs}`);
console.log('');

await probePage();

const allRows = [];
for (const account of accounts) {
  for (let i = 1; i <= runs; i += 1) {
    const row = await probeAccount(account, i);
    allRows.push(row);
    const status = row.ok ? 'ok' : 'fail';
    console.log(
      `${account.label}#${i} ${status} ` +
        `http=${row.httpStatus ?? '-'} headers=${fmt(row.headersMs)}ms ` +
        `ttft=${fmt(row.ttftMs)}ms server=${fmt(row.serverDurationMs)}ms total=${fmt(row.totalMs)}ms` +
        (row.error ? ` error=${row.error}` : ''),
    );
  }
}

console.log('');
console.log(`| account | ok/runs | median ttft ms | median server ms | median total ms | avg total ms | errors |`);
console.log(`|---|---:|---:|---:|---:|---:|---|`);
for (const account of accounts) {
  const rows = allRows.filter((row) => row.accountId === account.id);
  const okRows = rows.filter((row) => row.ok);
  const errors = rows
    .filter((row) => !row.ok)
    .map((row) => row.error || `http ${row.httpStatus}`)
    .slice(0, 2)
    .join('; ');
  console.log(
    `| ${escapeCell(account.label)} | ${okRows.length}/${rows.length} | ` +
      `${fmt(median(okRows.map((row) => row.ttftMs)))} | ` +
      `${fmt(median(okRows.map((row) => row.serverDurationMs)))} | ` +
      `${fmt(median(okRows.map((row) => row.totalMs)))} | ` +
      `${fmt(avg(okRows.map((row) => row.totalMs)))} | ${escapeCell(errors)} |`,
  );
}

async function probePage() {
  const rows = [];
  for (let i = 0; i < Math.min(runs, 3); i += 1) {
    const started = performance.now();
    let headersAt = null;
    try {
      const resp = await fetch(`${baseURL}/admin/accounts`, {
        method: 'GET',
        signal: AbortSignal.timeout(timeoutMs),
      });
      headersAt = performance.now();
      await resp.arrayBuffer();
      rows.push({
        status: resp.status,
        headersMs: headersAt - started,
        totalMs: performance.now() - started,
      });
    } catch (err) {
      rows.push({ status: 'ERR', headersMs: headersAt == null ? null : headersAt - started, totalMs: performance.now() - started, error: err.message });
    }
  }
  console.log(`## page probe /admin/accounts`);
  for (const [idx, row] of rows.entries()) {
    console.log(
      `page#${idx + 1} http=${row.status} headers=${fmt(row.headersMs)}ms total=${fmt(row.totalMs)}ms` +
        (row.error ? ` error=${redact(row.error)}` : ''),
    );
  }
  console.log('');
}

async function probeAccount(account, run) {
  const started = performance.now();
  const body = JSON.stringify({ model_id: model, prompt });
  let headersAt = null;
  const result = {
    accountId: account.id,
    account: account.label,
    run,
    ok: false,
    httpStatus: null,
    headersMs: null,
    ttftMs: null,
    serverDurationMs: null,
    totalMs: null,
    error: '',
  };

  try {
    const resp = await fetch(`${baseURL}/api/v1/admin/accounts/${account.id}/test`, {
      method: 'POST',
      headers: {
        ...authHeaders,
        'content-type': 'application/json',
        accept: 'text/event-stream',
      },
      body,
      signal: AbortSignal.timeout(timeoutMs),
    });
    headersAt = performance.now();
    result.httpStatus = resp.status;
    result.headersMs = headersAt - started;

    if (resp.status !== 200) {
      result.error = redact(truncate(await resp.text(), 220));
      result.totalMs = performance.now() - started;
      return result;
    }

    const events = await readSSE(resp, started);
    result.totalMs = performance.now() - started;
    for (const event of events) {
      if (event.type === 'content' && result.ttftMs == null) {
        result.ttftMs = event.ttft_ms ?? event._client_ms ?? null;
      }
      if (event.type === 'test_complete') {
        result.ok = event.success === true;
        result.serverDurationMs = event.duration_ms ?? null;
        result.ttftMs = result.ttftMs ?? event.ttft_ms ?? null;
      }
      if (event.type === 'error') {
        result.error = redact(truncate(event.error || 'account test error', 220));
        result.serverDurationMs = event.duration_ms ?? null;
        result.ttftMs = result.ttftMs ?? event.ttft_ms ?? null;
      }
    }
    if (!result.ok && !result.error) {
      result.error = 'missing test_complete event';
    }
    return result;
  } catch (err) {
    result.headersMs = headersAt == null ? null : headersAt - started;
    result.totalMs = performance.now() - started;
    result.error = redact(formatError(err));
    return result;
  }
}

async function readSSE(resp, started) {
  const reader = resp.body?.getReader();
  if (!reader) {
    return [];
  }
  const decoder = new TextDecoder();
  const events = [];
  let buffer = '';
  for (;;) {
    const { done, value } = await reader.read();
    if (done) {
      break;
    }
    buffer += decoder.decode(value, { stream: true });
    let newline;
    while ((newline = buffer.indexOf('\n')) >= 0) {
      const line = buffer.slice(0, newline).trim();
      buffer = buffer.slice(newline + 1);
      if (!line.startsWith('data:')) {
        continue;
      }
      const raw = line.slice(5).trim();
      if (!raw || raw === '[DONE]') {
        continue;
      }
      try {
        const event = JSON.parse(raw);
        event._client_ms = performance.now() - started;
        events.push(event);
      } catch {
        // Ignore malformed SSE lines; the final error/missing-complete check will catch failed probes.
      }
    }
  }
  return events;
}

async function resolveAdminAuthHeaders() {
  if (process.env.ADMIN_API_KEY) {
    return { 'x-api-key': process.env.ADMIN_API_KEY };
  }
  if (process.env.ADMIN_TOKEN) {
    return { authorization: `Bearer ${process.env.ADMIN_TOKEN}` };
  }

  const email = process.env.ADMIN_EMAIL || 'admin@sub2api.local';
  const password = process.env.ADMIN_PASSWORD;
  if (!password) {
    fail('Set ADMIN_TOKEN, ADMIN_API_KEY, or ADMIN_PASSWORD for admin account login.');
  }

  const resp = await fetch(`${baseURL}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ email, password }),
    signal: AbortSignal.timeout(timeoutMs),
  });
  const text = await resp.text();
  if (resp.status !== 200) {
    fail(`admin login failed: http=${resp.status} body=${redact(truncate(text, 220))}`);
  }
  let payload;
  try {
    payload = JSON.parse(text);
  } catch {
    fail(`admin login returned non-json body: ${redact(truncate(text, 220))}`);
  }
  const token = payload?.data?.access_token;
  if (!token) {
    fail(`admin login did not return access_token: ${redact(truncate(text, 220))}`);
  }
  return { authorization: `Bearer ${token}` };
}

function parseAccounts(raw) {
  return raw
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
    .map((item) => {
      const sep = item.includes('=') ? '=' : item.includes(':') ? ':' : null;
      const idRaw = sep ? item.slice(0, item.indexOf(sep)).trim() : item;
      const labelRaw = sep ? item.slice(item.indexOf(sep) + 1).trim() : '';
      const id = Number(idRaw);
      if (!Number.isInteger(id) || id <= 0) {
        fail(`Invalid account id in ACCOUNTS: ${item}`);
      }
      return { id, label: labelRaw || `account-${id}` };
    });
}

function median(values) {
  const nums = values.filter((value) => Number.isFinite(value)).sort((a, b) => a - b);
  if (nums.length === 0) {
    return null;
  }
  const mid = Math.floor(nums.length / 2);
  return nums.length % 2 === 0 ? (nums[mid - 1] + nums[mid]) / 2 : nums[mid];
}

function avg(values) {
  const nums = values.filter((value) => Number.isFinite(value));
  if (nums.length === 0) {
    return null;
  }
  return nums.reduce((sum, value) => sum + value, 0) / nums.length;
}

function fmt(value) {
  if (!Number.isFinite(value)) {
    return '-';
  }
  return Math.round(value).toString();
}

function positiveInt(raw, fallback) {
  const n = Number(raw);
  return Number.isInteger(n) && n > 0 ? n : fallback;
}

function stripTrailingSlash(value) {
  return String(value).replace(/\/+$/, '');
}

function truncate(value, max) {
  const text = String(value || '');
  return text.length > max ? `${text.slice(0, max)}...` : text;
}

function redact(value) {
  return String(value || '').replace(/sk-[A-Za-z0-9_-]{12,}/g, 'sk-***');
}

function formatError(err) {
  const message = err?.message || String(err);
  const cause = err?.cause;
  const detail = cause?.code || cause?.name || cause?.message;
  return detail ? `${message} (${detail})` : message;
}

function escapeCell(value) {
  return String(value || '').replaceAll('|', '\\|').replace(/\s+/g, ' ').trim();
}

function fail(message) {
  console.error(message);
  process.exit(1);
}
