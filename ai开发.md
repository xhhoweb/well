# WellCMS Go Headless · CodeAI Final Spec（融合修订版）

> **用途声明（非常重要）**  
> 本文档是 **唯一、完整、最终的 CodeAI 开发输入规范**。  
> 由以下三部分**融合修订**而成：  
> 1️⃣ 原 `well_cms_go版开发文档.md`（完整 WellCMS 思想与结构）  
> 2️⃣ Headless / Personal Edition（去后台、去插件、执行路径收敛）  
> 3️⃣ 针对 CodeAI 的「防过度设计」硬性约束  
>
> **CodeAI 必须严格按本文档实现，不得自行扩展架构、抽象或机制。**

---

## 0. 项目总定位（最高优先级）

### 0.1 项目是什么
- 一个 **个人自用的高性能内容 API 引擎**
- 继承 **WellCMS 原作者的极端性能与克制设计思想**
- 服务对象仅为：前端 / CLI / 自动化程序

### 0.2 项目不是什么
- ❌ 不是 CMS 产品
- ❌ 不支持插件生态
- ❌ 不支持主题系统
- ❌ 不提供 Web 后台页面
- ❌ 不考虑第三方开发者体验

> **铁律**：任何“为了以后可能用到”的设计，一律禁止。

---

## 1. 核心设计哲学（来自 WellCMS）

### 1.1 三大不可动摇原则

1. **缓存优先于数据库**  
   数据库只承担最终一致性存储角色。

2. **执行路径必须收敛**  
   - 启动期：允许复杂
   - 请求期：必须简单、固定、可预测

3. **少机制 > 好机制**  
   - 无运行期插件
   - 无反射分发
   - 无动态配置热加载

---

## 2. 技术选型（稳定优先，低门槛）

| 模块 | 选型 | 约束 |
|---|---|---|
| 语言 | Go 1.22+ | 静态编译 |
| Web | Gin（默认） / Fiber（可选） | Gin 为推荐 |
| 数据库 | MySQL 5.7+ | 推荐 8.0，不强制 |
| SQL | sqlx | ❌ 禁止 ORM |
| ID | Snowflake | 全表 BIGINT |
| 缓存 | L1: bigcache / L2: Redis | L1 必须零 GC |
| 并发防护 | singleflight | 仅限热点读 |
| 鉴权 | JWT | 仅替代 Session |
| 日志 | Zap | 结构化 |
| 调试 | pprof | 必须支持 |

---

## 3. 不可争议编码规则（CodeAI 强约束）

### 3.1 分层铁律
- ❌ Handler 禁止写 SQL
- ❌ Handler 禁止操作 Redis
- ❌ Repository 禁止业务判断
- ❌ Service 才能决定缓存策略

### 3.2 数据结构铁律
- ❌ 禁止 map[string]interface{}
- ❌ 禁止 interface{} 作为业务返回
- ❌ 禁止运行期反射

### 3.3 运行期铁律
- ❌ 禁止运行期注册 hook / plugin
- ❌ 禁止配置热加载（Watch）
- ❌ 禁止请求期 new 大对象

> **违反任意一条 = 错误实现**

---

## 4. 目录结构（融合原 WellCMS + Headless）

```
wellcms-go/
├── cmd/api/main.go
├── internal/
│   ├── api/
│   │   ├── v1/        # Public Read API
│   │   └── mgt/       # Management API（运维接口）
│   ├── core/
│   │   ├── config/
│   │   ├── database/
│   │   ├── logger/
│   │   ├── snowflake/
│   │   └── runtime/   # 预热数据
│   ├── middleware/
│   ├── model/         # 映射 WellCMS 表结构
│   ├── repository/
│   ├── service/
│   └── pkg/
│       ├── apperr/
│       ├── response/
│       ├── pool/
│       └── util/
├── scripts/
├── config.yaml
└── go.mod
```

---

## 5. 启动流程（执行路径锁死）

1. 加载配置（仅启动期）
2. 初始化 Logger
3. 初始化 MySQL
4. 初始化 Redis
5. 初始化 L1/L2 缓存
6. 初始化 Snowflake
7. 预热 runtime 数据（forum / 权限 / 配置）
8. 注册路由
9. 启动 HTTP Server

---

## 6. 数据模型设计（完全继承 WellCMS）

### 6.1 thread 主表（列表专用）

```sql
CREATE TABLE thread (
  tid BIGINT UNSIGNED PRIMARY KEY,
  fid INT UNSIGNED NOT NULL,
  uid BIGINT UNSIGNED NOT NULL,
  subject VARCHAR(120) NOT NULL,
  views INT UNSIGNED NOT NULL DEFAULT 0,
  replies INT UNSIGNED NOT NULL DEFAULT 0,
  dateline INT UNSIGNED NOT NULL,
  lastpost INT UNSIGNED NOT NULL,
  status TINYINT UNSIGNED NOT NULL DEFAULT 0,
  KEY idx_fid_lastpost (fid, lastpost),
  KEY idx_uid (uid)
) ENGINE=InnoDB;
```

### 6.2 thread_data 内容表

```sql
CREATE TABLE thread_data (
  tid BIGINT UNSIGNED PRIMARY KEY,
  message MEDIUMTEXT
) ENGINE=InnoDB;
```

**原则**：
- 主表只为列表服务
- 大字段必须拆表
- 禁止 join

---

## 7. Repository 标准模板

```go
type ThreadRepository interface {
  GetByID(ctx context.Context, tid int64) (*model.Thread, error)
}
```

---

## 8. Service 标准模板（缓存 + singleflight）

```go
func (s *ThreadService) Get(ctx context.Context, tid int64) (*ThreadDTO, error) {
  key := fmt.Sprintf("thread:%d", tid)

  if v := s.l1.Get(key); v != nil {
    return v.(*ThreadDTO), nil
  }

  if v, err := s.l2.Get(key); err == nil {
    s.l1.Set(key, v, ttl)
    return v.(*ThreadDTO), nil
  }

  v, err, _ := s.sf.Do(key, func() (any, error) {
    t, err := s.repo.GetByID(ctx, tid)
    if err != nil {
      return nil, err
    }
    dto := MapThreadDTO(t)
    s.l1.Set(key, dto, ttl)
    s.l2.Set(key, dto, ttl)
    return dto, nil
  })

  if err != nil {
    return nil, err
  }
  return v.(*ThreadDTO), nil
}
```

---

## 9. API 设计（去后台化）

### 9.1 Public API（只读）
- GET /api/v1/threads
- GET /api/v1/thread/:tid

### 9.2 Management API（运维接口）
- POST /api/mgt/login
- POST /api/mgt/thread
- PUT /api/mgt/thread/:tid
- DELETE /api/mgt/thread/:tid
- POST /api/mgt/cache/flush
- POST /api/mgt/cache/prewarm

> Management API **不是后台系统**，仅用于你自己操作。

---

## 10. 缓存与运行时优化（WellCMS 隐性核心）

### 10.1 缓存分级
- 强缓存：30~120s（首页 / 列表）
- 软缓存：详情页，可手动失效
- 预热缓存：forum / 权限 / 配置

### 10.2 Pool 复用
- sync.Pool：DTO / buffer
- 禁止热路径 new slice / map

### 10.3 GC 铁律
- 单请求 alloc < 5KB
- 禁止 interface{} 逃逸

---

## 11. 给 CodeAI 的最终执行指令

> 你正在实现的是一个 **个人自用、极端克制、高性能的内容引擎**。
>
> 不允许：
> - 过度抽象
> - 预留扩展点
> - 引入新框架或插件系统
>
> 只允许：
> - 严格按本文档
> - 按模板填充代码
> - 保持执行路径简单、可预测

---

**本文档已融合原 WellCMS Go 开发文档与 Headless 优化，是唯一推荐的 CodeAI 开发规范。**

