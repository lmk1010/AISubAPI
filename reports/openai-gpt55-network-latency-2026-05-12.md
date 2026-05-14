# OpenAI GPT-5.5 渠道网络测试报告

测试时间：2026-05-12 22:20-22:30 CST
测试模型：`gpt-5.5`
测试账号：`2 fluxnode-org`、`3 pixel-try-chatapi`、`9 charitydoing-new`
请求方式：Responses API，`stream=true`，小 prompt：`请只回复：OK`
每项样本：3 次

> 说明：报告不记录 API Key。`api.jiminaishop.com` 路径使用后台账号测试接口，按 account id 指定渠道，不走分组随机调度。

## 结论

1. `charitydoing-new` 是本轮最适合 GPT-5.5 的渠道。生产机直连中位总耗时 `2.06s`，生产本机后端中位 `1.83s`，我本机经 `api.jiminaishop.com` 中位 `2.34s`。
2. `pixel-try-chatapi` 明显偏慢。生产机直连源头时 DNS/TLS 已经偏高，DNS 中位 `301ms`，TLS 中位 `471ms`；本机经网关中位总耗时 `5.09s`。
3. `fluxnode-org` 在本机直连源头有严重长尾，单次总耗时到 `23.62s`；经服务器转发后中位明显好很多。
4. `api.jiminaishop.com` 本身不是主瓶颈。`/admin/accounts` 页面本机访问首包约 `269-495ms`，完整下载中位约 `496ms`。GPT 请求慢主要来自后端到上游、上游自身生成/流式返回抖动，而不是本机到你的站点这段。
5. 本机直连三个源头整体都比生产机直连慢，尤其 TLS：本机 TLS 中位 `0.81s-1.46s`，生产机 TLS 中位 `0.17s-0.47s`。这说明让用户经 `api.jiminaishop.com` 走你的服务器转发，通常比用户直连源头更稳。

## 汇总表

### 生产机直连源头 URL

| 渠道 | 成功率 | DNS 中位 | TLS 中位 | TTFB 中位 | 总耗时中位 | 平均总耗时 |
|---|---:|---:|---:|---:|---:|---:|
| fluxnode-org | 3/3 | 71ms | 174ms | 959ms | 2730ms | 2656ms |
| pixel-try-chatapi | 3/3 | 301ms | 471ms | 1753ms | 3175ms | 3457ms |
| charitydoing-new | 3/3 | 64ms | 188ms | 893ms | 2062ms | 2197ms |

### 生产机本机后端路径

测试路径：生产宿主机 `127.0.0.1:18080 -> sub2api -> 上游`

| 渠道 | 成功率 | 后端首包中位 | 总耗时中位 | 平均总耗时 |
|---|---:|---:|---:|---:|
| fluxnode-org | 3/3 | 12ms | 2018ms | 2098ms |
| pixel-try-chatapi | 3/3 | 13ms | 2726ms | 2726ms |
| charitydoing-new | 3/3 | 11ms | 1830ms | 1864ms |

> 后端首包很低是因为测试接口会先 flush `test_start` SSE，不代表上游首 token。

### 我本机经过 api.jiminaishop.com

测试路径：本机 -> `https://api.jiminaishop.com` -> sub2api -> 上游

| 渠道 | 成功率 | 首内容中位 | 总耗时中位 | 平均总耗时 |
|---|---:|---:|---:|---:|
| fluxnode-org | 3/3 | 3308ms | 4039ms | 3919ms |
| pixel-try-chatapi | 3/3 | 4673ms | 5094ms | 4608ms |
| charitydoing-new | 3/3 | 1716ms | 2344ms | 2439ms |

`/admin/accounts` 页面网络基线：

| 样本 | HTTP | 首包 | 总耗时 |
|---|---:|---:|---:|
| 1 | 200 | 269ms | 270ms |
| 2 | 200 | 494ms | 498ms |
| 3 | 200 | 495ms | 496ms |

### 我本机直连源头 URL

| 渠道 | 成功率 | DNS 中位 | TLS 中位 | TTFB 中位 | 总耗时中位 | 平均总耗时 |
|---|---:|---:|---:|---:|---:|---:|
| fluxnode-org | 3/3 | 3ms | 812ms | 2479ms | 5388ms | 10821ms |
| pixel-try-chatapi | 3/3 | 3ms | 1456ms | 4334ms | 5887ms | 5892ms |
| charitydoing-new | 3/3 | 2ms | 1217ms | 2225ms | 4884ms | 4811ms |

## 原始数据

### 生产机直连源头

```text
prod_direct 2 fluxnode-org 1 http=200 dns=0.071302 tcp=0.086101 tls=0.166991 ttfb=1.085948 total=2.729630 ok=1
prod_direct 2 fluxnode-org 2 http=200 dns=0.076347 tcp=0.091406 tls=0.204629 ttfb=0.835401 total=2.840755 ok=1
prod_direct 2 fluxnode-org 3 http=200 dns=0.065063 tcp=0.066740 tls=0.173638 ttfb=0.958782 total=2.397313 ok=1
prod_direct 3 pixel-try-chatapi 1 http=200 dns=0.304803 tcp=0.354384 tls=0.487004 ttfb=1.752911 total=3.175158 ok=1
prod_direct 3 pixel-try-chatapi 2 http=200 dns=0.300580 tcp=0.363890 tls=0.468224 ttfb=2.038295 total=4.740439 ok=1
prod_direct 3 pixel-try-chatapi 3 http=200 dns=0.294939 tcp=0.358872 tls=0.471099 ttfb=1.550960 total=2.455841 ok=1
prod_direct 9 charitydoing-new 1 http=200 dns=0.050425 tcp=0.051316 tls=0.110607 ttfb=0.821185 total=2.061522 ok=1
prod_direct 9 charitydoing-new 2 http=200 dns=0.076955 tcp=0.077941 tls=0.233507 ttfb=1.271568 total=2.479746 ok=1
prod_direct 9 charitydoing-new 3 http=200 dns=0.064364 tcp=0.065293 tls=0.188104 ttfb=0.893400 total=2.050238 ok=1
```

### 生产机本机后端路径

```text
prod_gateway 2 fluxnode-org 1 http=200 dns=0.000112 tcp=0.000375 ttfb=0.011536 total=2.565363 ok=1
prod_gateway 2 fluxnode-org 2 http=200 dns=0.000035 tcp=0.000650 ttfb=0.008788 total=2.018294 ok=1
prod_gateway 2 fluxnode-org 3 http=200 dns=0.000108 tcp=0.000498 ttfb=0.013761 total=1.711667 ok=1
prod_gateway 3 pixel-try-chatapi 1 http=200 dns=0.000066 tcp=0.000378 ttfb=0.012974 total=3.678356 ok=1
prod_gateway 3 pixel-try-chatapi 2 http=200 dns=0.000057 tcp=0.000730 ttfb=0.013698 total=2.725522 ok=1
prod_gateway 3 pixel-try-chatapi 3 http=200 dns=0.000053 tcp=0.000358 ttfb=0.010382 total=1.775254 ok=1
prod_gateway 9 charitydoing-new 1 http=200 dns=0.000064 tcp=0.001007 ttfb=0.010994 total=2.032663 ok=1
prod_gateway 9 charitydoing-new 2 http=200 dns=0.000106 tcp=0.000465 ttfb=0.008947 total=1.830063 ok=1
prod_gateway 9 charitydoing-new 3 http=200 dns=0.000054 tcp=0.000307 ttfb=0.012179 total=1.728603 ok=1
```

### 我本机经过 api.jiminaishop.com

```text
page#1 http=200 headers=269ms total=270ms
page#2 http=200 headers=494ms total=498ms
page#3 http=200 headers=495ms total=496ms
fluxnode-org#1 ok http=200 headers=261ms ttft=1973ms total=2386ms
fluxnode-org#2 ok http=200 headers=1524ms ttft=4816ms total=5333ms
fluxnode-org#3 ok http=200 headers=1867ms ttft=3308ms total=4039ms
pixel-try-chatapi#1 ok http=200 headers=4481ms ttft=4673ms total=5094ms
pixel-try-chatapi#2 ok http=200 headers=2155ms ttft=4868ms total=5516ms
pixel-try-chatapi#3 ok http=200 headers=1699ms ttft=2779ms total=3215ms
charitydoing-new#1 ok http=200 headers=293ms ttft=1716ms total=2344ms
charitydoing-new#2 ok http=200 headers=1421ms ttft=1879ms total=2666ms
charitydoing-new#3 ok http=200 headers=739ms ttft=1549ms total=2306ms
```

### 我本机直连源头

```text
local_direct 2 fluxnode-org 1 http=200 dns=0.256799 tcp=0.257244 tls=1.983336 ttfb=2.726941 total=23.615539 ok=1
local_direct 2 fluxnode-org 2 http=200 dns=0.003058 tcp=0.003423 tls=0.803594 ttfb=2.479063 total=5.388218 ok=1
local_direct 2 fluxnode-org 3 http=200 dns=0.001722 tcp=0.002141 tls=0.812052 ttfb=1.999590 total=3.459057 ok=1
local_direct 3 pixel-try-chatapi 1 http=200 dns=0.040996 tcp=0.041424 tls=2.692693 ttfb=4.464754 total=5.887025 ok=1
local_direct 3 pixel-try-chatapi 2 http=200 dns=0.001920 tcp=0.002368 tls=1.455843 ttfb=2.619521 total=6.492105 ok=1
local_direct 3 pixel-try-chatapi 3 http=200 dns=0.003107 tcp=0.003553 tls=1.143665 ttfb=4.334204 total=5.296236 ok=1
local_direct 9 charitydoing-new 1 http=200 dns=0.002172 tcp=0.002563 tls=2.629888 ttfb=3.634415 total=5.814580 ok=1
local_direct 9 charitydoing-new 2 http=200 dns=0.001833 tcp=0.002321 tls=1.217468 ttfb=2.200270 total=4.883731 ok=1
local_direct 9 charitydoing-new 3 http=200 dns=0.002702 tcp=0.003152 tls=0.888116 ttfb=2.225138 total=3.734020 ok=1
```

## 定位判断

### 1. 公网域名不是主要瓶颈

我本机访问 `api.jiminaishop.com/admin/accounts` 的首包在 `269-495ms`，这段只能解释几百毫秒，不足以解释 GPT 请求里的 `3s-5s`。

### 2. 上游网络质量排序

从生产机直连源头看：

1. `charitydoing-new`：最快，生产机到源头路径质量最好。
2. `fluxnode-org`：可用，整体中等。
3. `pixel-try-chatapi`：生产机 DNS/TLS 都偏高，整体最慢。

### 3. 用户直连源头更差

我本机直连源头时，三家 TLS 都明显高于生产机直连，说明这些源头对国内/本地线路并不好。让用户走你的 `api.jiminaishop.com` 由美国源站转发，整体更合理。

### 4. 当前最优路由策略

GPT-5.5 建议优先级：

| 优先级 | 渠道 | 原因 |
|---:|---|---|
| 1 | charitydoing-new | 网关路径最快，生产源站路径也最快 |
| 2 | fluxnode-org | 生产路径稳定，但本机直连有长尾 |
| 3 | pixel-try-chatapi | DNS/TLS 和总耗时都偏高 |

注意：`charitydoing-new` 当前没有绑定任何分组；如果要投入真实调度，需要单独绑定目标 OpenAI 分组并设置优先级。

## 建议动作

1. 如果要优化用户 GPT-5.5 体验，先把 `charitydoing-new` 放到目标 GPT 分组第一优先级，小流量观察。
2. `pixel-try-chatapi` 不建议放高优先级，除非它价格或额度优势明显。
3. `direct.jiminaishop.com` 这类绕 CF 入口最多优化几百毫秒；对 GPT-5.5 这种请求，主要收益不在这里。
4. 后台账号测试接口生产环境目前没有返回服务端 `duration_ms/ttft_ms` 字段；本报告的经网关 TTFT 是客户端侧首个 content SSE 到达时间。若要长期观测，建议部署后台测试延时字段改动。
