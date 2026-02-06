# WellCMS Go Headless · Code Review Spec（给 Codex 的最终审查规范）

> **用途声明（最高优先级）**
> 本文档用于指导 Codex/CodeAI 对当前仓库进行代码审查与修复建议输出。
> 审查目标是：**符合 WellCMS 的极端克制、高性能、执行路径收敛思想**。
> CodeAI 必须严格按此规范输出，不得扩展架构、不得增加新模块。

---

## 0. 审查范围

对整个 Go 项目进行全面审查，包括：

* `cmd/api/main.go`
* `internal/api/*`
* `internal/service/*`
* `internal/repository/*`
* `internal/pkg/*`
* `internal/core/*`
* `config.yaml`
* `go.mod/go.sum`

---

## 1. 审查输出格式（必须严格遵守）

CodeAI 输出必须分 4 大部分：

### 1.1 🔴 必改问题（Bug / 安全 / 逻辑错误）

* 必须指出文件路径
* 必须指出具体函数名/代码段
* 必须给出修复方案（可直接贴代码 diff 或替换片段）

### 1.2 🟡 性能问题（热点路径 alloc / GC / IO）

* 必须指出是否影响 QPS
* 必须指出是否会造成 GC 压力
* 必须指出是否违背 “执行路径收敛”

### 1.3 🟢 已符合规范的优秀实现

* 只列关键点（不要废话）
* 用于确认该实现不需要改

### 1.4 📌 最终修复清单（可执行 TODO）

输出一个 checklist：

* [ ] 修改 xxx.go
* [ ] 删除 xxx
* [ ] 修复 xxx

---

## 2. 核心铁律（审查必须以此为最高标准）

### 2.1 执行路径收敛

* 请求期必须简单、固定、可预测
* 禁止运行期扫描目录/插件注册/动态 hook
* 禁止反射分发
* 禁止配置热加载 Watch

### 2.2 缓存优先于数据库

* 高频读必须 L1/L2
* DB 只承担最终一致性存储
* 禁止热点请求直接回源 DB

### 2.3 禁止过度设计

任何 “未来可能用到” 的能力一律判定为错误：

* gzip 中间件预留
* refresh token
* plugin/hook
* 多余 health endpoint
* 多余 middleware

---

## 3. 分层结构铁律（必须检查）

### 3.1 Handler 层

❌ 禁止写 SQL
❌ 禁止操作 Redis
❌ 禁止写缓存逻辑
只负责：

* 参数解析
* 调用 service
* 返回 response

### 3.2 Repository 层

❌ 禁止写业务判断
❌ 禁止缓存
只负责：

* SQL 查询
* 映射 model

### 3.3 Service 层

必须负责：

* 缓存策略（L1/L2 TTL）
* singleflight 防击穿
* DTO 构造
* 序列化/反序列化

---

## 4. 数据结构与性能规则（必须检查）

### 4.1 禁止结构

❌ map[string]interface{}
❌ interface{} 作为业务返回
❌ runtime reflect

### 4.2 热路径 GC 铁律

目标：

* L1 命中路径 alloc 越少越好（允许 json.Unmarshal 的合理分配）
* 禁止 map 缓存对象图

### 4.3 L1 缓存必须使用 bigcache（零 GC）

必须确认：

* L1 Set/Get 都是 `[]byte`
* 禁止 L1 缓存 `*DTO`
* 禁止隐式 copy

---

## 5. 缓存策略审查（必须）

### 5.1 二级缓存结构

必须符合：

* L1: bigcache（进程内）
* L2: redis（分布式）
* DB: mysql（最终一致性）

### 5.2 singleflight

必须确认：

* 只用于热点 key
* L1/L2 miss 才进入 singleflight
* singleflight 内必须写回缓存

---

## 6. MySQL 设计检查（必须）

必须确认：

* 主表瘦表（thread）
* 大字段拆表（thread_data）
* 禁止 join
* 列表查询必须命中索引
* 索引设计合理（fid+lastpost / fid+dateline）

如果发现 join 或大字段在主表，判定为严重错误。

---

## 7. Redis 使用检查（必须）

必须确认：

* redis key 命名统一（thread:123）
* TTL 合理（30~120s 列表强缓存；详情软缓存）
* 失败不影响主业务（redis down 应回源 DB）

---

## 8. main.go 启动流程检查（必须）

必须确认启动顺序：

1. config init
2. logger init
3. mysql init
4. redis init + ping
5. bigcache init
6. snowflake init
7. runtime warmup
8. router init
9. server run

必须检查：

* baseURL 不能写死 localhost（SEO 严重错误）
* pprof 是否启用且不暴露公网

---

## 9. SEO 模块检查（必须）

### 9.1 必须实现

* `/robots.txt`
* `/sitemap.xml`
* sitemap index + thread sitemap 分片
* canonical（header 或 middleware）

### 9.2 必须确认

* sitemap `<loc>` 必须使用真实域名 baseURL（来自 config）
* sitemap 最大 50000 URL/文件
* sitemap 必须缓存（避免每次打 DB）
* robots.txt 必须包含 sitemap url

### 9.3 禁止事项

❌ SEO 伪静态 `.html`
❌ rewrite 装饰
❌ SEO 插件系统

---

## 10. 监控与调试检查（必须）

必须确认：

* `/metrics` 可用（prometheus）
* pprof 可用（localhost 或内网）
* 日志 zap 不阻塞请求

---

## 11. 安全规则（必须）

必须检查：

* JWT 是否无 refresh token（除非明确要求）
* mgt API 是否强制 IP 白名单
* rate limit 是否生效
* redis/mysql 密码不输出日志

---

## 12. 过度实现检测（必须列出）

审查时必须主动寻找：

* gzip middleware
* cors 配置化过多
* context timeout 过度包装
* 自定义框架/IOC容器
* plugin/hook 目录

发现后必须列入 🟡 或 🔴。

---

## 13. CodeAI 输出要求（必须）

### 13.1 不允许

* 提议引入新框架（fiber、echo、grpc）
* 提议引入 ORM（gorm）
* 提议引入插件系统
* 提议抽象成复杂模块

### 13.2 允许

* 删除冗余代码
* 修复 bug
* 优化 key 拼接
* 调整 TTL / redis key 规则
* 修复 sitemap baseURL 错误

---

## 14. 最终目标定义

审查后的仓库必须满足：

* 执行路径收敛（无动态机制）
* bigcache L1 + redis L2 正确
* singleflight 防击穿正确
* SEO sitemap/robots/canonical 正确
* main.go 启动流程正确
* 不存在明显过度实现

---

## 15. 最终审查任务指令（直接给 Codex）

你现在要做的是：

1. 全仓库代码扫描
2. 找出违反上述铁律的代码
3. 输出结构化审查报告
4. 给出明确可执行的修复建议
5. 不要增加任何未要求功能

输出必须包含：

* 必改问题
* 性能问题
* 已符合点
* TODO 清单

---

# 结束

> 本文档是唯一审查标准。
> 除非明确要求，否则禁止提出任何“扩展性设计”。
> 目标是 **个人自用、极端克制、高性能 WellCMS Go 引擎**。
