# WellCMS Go 2.0 详细开发文档

## 目录
1. [项目概述](#1-项目概述)
2. [技术架构](#2-技术架构)
3. [目录结构](#3-目录结构)
4. [核心模块分析](#4-核心模块分析)
5. [数据库设计](#5-数据库设计)
6. [路由系统](#6-路由系统)
7. [模型层](#7-模型层)
8. [插件机制](#8-插件机制)
9. [视图模板](#9-视图模板)
10. [后台管理](#10-后台管理)
11. [配置系统](#11-配置系统)
12. [缓存机制](#12-缓存机制)
13. [安全机制](#13-安全机制)
14. [API接口](#14-api接口)
15. [开发规范](#15-开发规范)
16. [扩展开发](#16-扩展开发)

---

## 1. 项目概述

### 1.1 项目简介
WellCMS Go 是一款具备亿级负载能力的高性能CMS系统，采用MIT协议开源，倾向移动端，轻量级，具有超快反应能力的高负载CMS的Go语言重写版本。

### 1.2 核心特点
- **高负载能力**: 亿级数据处理能力，单表可承载亿级以上数据
- **高性能**: 单次请求处理时间在0.01秒级别，开启缓存可达0.003秒
- **轻量级**: 基于自研Go Web框架，只有22张表
- **多端支持**: 支持PC、Pad、手机自适应，支持独立模板绑定
- **插件机制**: 强大的插件机制，支持中间件和处理器注入
- **多语言**: 支持简体中文、繁体中文、英文

### 1.3 技术栈
- **后端**: Go 1.21+ (最低支持Go 1.18)
- **数据库**: MySQL 5.5.6+ 或 MariaDB 或 PostgreSQL
- **缓存**: 支持Redis、Memcached、Go内建缓存
- **前端**: Bootstrap 4.5 + JQuery 3.5 (保持兼容)
- **Web框架**: 自研WellHTTP框架 (类似Gin的轻量级框架)
- **ORM**: XORM或自研轻量级ORM

### 1.4 开发原则
1. 不写重复无用的代码，保证代码干净、整洁、可读性强
2. 不设计过度重复的业务逻辑，对于大数组用完即释放
3. 不在MySQL做任何运算，只把MySQL当作储存库使用
4. 对每个表尽量只查询一次，不做重复查询
5. 绝不在业务层直接使用SQL语句，使用ORM封装
6. 尽量不使用过于复杂的Go特性，保持代码可维护性

---

## 2. 技术架构

### 2.1 整体架构图
```
┌─────────────────────────────────────────────────────┐
│                     客户端                           │
│         (浏览器/APP/AJAX请求)                        │
└─────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────┐
│                    Nginx/Caddy                       │
│                    (反向代理)                        │
└─────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────┐
│                   main.go                            │
│                   (入口文件)                         │
└─────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────┐
│                   wellhttp/                          │
│                   (核心框架)                         │
└─────────────────────────────────────────────────────┘
                           │
           ┌───────────────┼───────────────┐
           ▼               ▼               ▼
     ┌──────────┐   ┌──────────────┐  ┌──────────┐
     │ handlers │   │   middleware  │  │  plugin  │
     └──────────┘   └──────────────┘  └──────────┘
           │               │               │
           ▼               ▼               ▼
     ┌──────────────────────────────────────────┐
     │              models/                      │
     │              (数据模型)                   │
     └──────────────────────────────────────────┘
           │               │
           ▼               ▼
     ┌──────────────────────────────────────────┐
     │              Database                     │
     │           (22张核心数据表)                │
     └──────────────────────────────────────────┘
```

### 2.2 请求处理流程
```
1. 用户请求 → Nginx
2. main.go 初始化
3. 加载配置文件
4. 初始化数据库连接
5. 初始化缓存系统
6. 加载中间件 (Session, 语言, 用户等)
7. 注册路由
8. 根据路由调用对应的Handler
9. 处理业务逻辑
10. 读取数据库/缓存
11. 渲染模板
12. 输出响应
```

### 2.3 核心启动参数
```go
// main.go 中定义
type Config struct {
    Debug       bool   `json:"debug"`        // 调试模式
    AppPath     string `json:"app_path"`     // 应用路径
    AdminPath   string `json:"admin_path"`   // 后台路径
    Port        int    `json:"port"`         // 监听端口
    DB          DBConfig `json:"db"`         // 数据库配置
    Cache       CacheConfig `json:"cache"`   // 缓存配置
}

type DBConfig struct {
    Host        string `json:"host"`
    Port        int    `json:"port"`
    Username    string `json:"username"`
    Password    string `json:"password"`
    Database    string `json:"database"`
    MaxOpenConn int    `json:"max_open_conn"`
    MaxIdleConn int    `json:"max_idle_conn"`
}

type CacheConfig struct {
    Type    string `json:"type"`    // redis, memcached, memory
    Host    string `json:"host"`
    Port    int    `json:"port"`
}
```

---

## 3. 目录结构

### 3.1 根目录结构
```
wellcms-go/
├── cmd/                        # 命令行工具
│   ├── server/                 # 主服务
│   │   └── main.go
│   └── migrate/                # 数据库迁移工具
├── internal/                   # 内部模块
│   ├── admin/                  # 后台管理模块
│   │   ├── handlers/           # 后台处理器
│   │   ├── middleware/         # 后台中间件
│   │   └── views/              # 后台视图
│   ├── api/                    # API接口模块
│   ├── auth/                   # 认证模块
│   ├── cache/                  # 缓存模块
│   ├── config/                 # 配置模块
│   ├── database/               # 数据库模块
│   ├── handlers/               # 前台处理器
│   ├── middleware/             # 中间件
│   ├── models/                 # 数据模型
│   ├── plugin/                 # 插件模块
│   ├── router/                 # 路由配置
│   ├── session/                # 会话管理
│   ├── template/               # 模板引擎
│   └── utils/                  # 工具函数
├── conf/                       # 配置文件目录
│   ├── conf.yaml               # 主配置文件
│   ├── conf.default.yaml       # 默认配置示例
│   └── smtp.yaml               # SMTP邮件配置
├── install/                    # 安装程序
│   ├── install.go              # 安装逻辑
│   ├── install.sql             # 数据库表结构
│   └── views/                  # 安装视图
├── lang/                       # 语言包目录
│   ├── zh-cn/                  # 简体中文
│   │   ├── lang.json           # 主语言包
│   │   ├── lang_admin.json     # 后台语言包
│   │   └── lang_install.json   # 安装语言包
│   ├── zh-tw/                  # 繁体中文
│   └── en-us/                  # 英文
├── log/                        # 日志目录
├── plugin/                     # 插件目录
├── public/                     # 静态资源
│   ├── css/                    # 样式文件
│   ├── fonts/                  # 字体文件
│   ├── js/                     # JavaScript文件
│   ├── upload/                 # 上传文件
│   │   ├── attach/             # 附件
│   │   ├── avatar/             # 头像
│   │   ├── flag/               # 属性图片
│   │   ├── forum/              # 版块图片
│   │   └── thumbnail/          # 缩略图
│   └── template/               # 模板主题目录
├── view/                       # 前台视图目录
│   ├── htm/                    # 模板文件
│   └── template/               # 模板主题
├── wellhttp/                   # 核心框架目录
├── go.mod                      # Go模块文件
├── go.sum                      # Go依赖校验
└── README.md                   # 项目说明
```

### 3.2 核心目录详解

#### 3.2.1 models/ 目录 (数据模型层)
```
internal/models/
├── user.go                     # 用户模型
├── user_group.go               # 用户组模型
├── session.go                  # 会话模型
├── forum.go                    # 版块模型
├── forum_access.go             # 版块权限模型
├── thread.go                   # 主题模型
├── thread_data.go              # 主题数据模型
├── thread_sticky.go            # 主题置顶模型
├── thread_tid.go               # 主题索引模型
├── comment.go                  # 评论模型
├── comment_pid.go              # 评论回复模型
├── flag.go                     # 属性模型
├── flag_thread.go              # 属性主题关联模型
├── tag.go                      # 标签模型
├── tag_thread.go               # 标签主题关联模型
├── attach.go                   # 附件模型
├── page.go                     # 单页模型
├── link.go                     # 链接模型
├── operate.go                  # 操作记录模型
├── kv.go                       # KV缓存模型
├── crontab.go                  # 计划任务模型
└── base.go                     # 基础模型
```

#### 3.2.2 handlers/ 目录 (前台处理器)
```
internal/handlers/
├── index.go                    # 首页处理器
├── read.go                     # 主题详情处理器
├── list.go                     # 版块列表处理器
├── category.go                 # 频道分类处理器
├── tag.go                      # 标签页处理器
├── flag.go                     # 属性页处理器
├── user.go                     # 用户中心处理器
├── home.go                     # 个人主页处理器
├── my.go                       # 我的主页处理器
├── attach.go                   # 附件处理器
├── comment.go                  # 评论处理器
└── operate.go                  # 操作处理器
```

#### 3.2.3 middleware/ 目录 (中间件)
```
internal/middleware/
├── cors.go                     # 跨域中间件
├── logger.go                   # 日志中间件
├── recovery.go                 # 异常恢复中间件
├── session.go                  # 会话中间件
├── auth.go                     # 认证中间件
├── permission.go               # 权限中间件
├── language.go                 # 语言中间件
├── cache.go                    # 缓存中间件
└── rate_limit.go               # 限流中间件
```

#### 3.2.4 wellhttp/ 目录 (核心框架)
```
wellhttp/
├── router.go                   # 路由核心
├── context.go                  # 请求上下文
├── response.go                 # 响应封装
├── middleware.go               # 中间件接口
├── binder.go                   # 参数绑定
├── renderer.go                 # 模板渲染器
├── session.go                  # 会话接口
├── cache.go                    # 缓存接口
├── logger.go                   # 日志接口
└── wellhttp.go                 # 框架入口
```

---

## 4. 核心模块分析

### 4.1 main.go (入口文件)
```go
package main

import (
    "flag"
    "log"
    "os"
    "wellcms-go/internal/config"
    "wellcms-go/internal/database"
    "wellcms-go/internal/cache"
    "wellcms-go/internal/router"
    "wellcms-go/internal/session"
    "wellcms-go/wellhttp"
)

func main() {
    // 解析命令行参数
    configPath := flag.String("conf", "conf/conf.yaml", "配置文件路径")
    flag.Parse()

    // 加载配置
    conf, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("加载配置失败: %v", err)
    }

    // 初始化数据库
    db, err := database.Init(conf.DB)
    if err != nil {
        log.Fatalf("初始化数据库失败: %v", err)
    }
    defer db.Close()

    // 初始化缓存
    cacheManager, err := cache.Init(conf.Cache)
    if err != nil {
        log.Fatalf("初始化缓存失败: %v", err)
    }

    // 初始化会话
    sessionManager := session.New(conf)

    // 创建框架实例
    app := wellhttp.New(wellhttp.Config{
        Debug:       conf.Debug,
        AppPath:     conf.AppPath,
        Session:     sessionManager,
        Cache:       cacheManager,
        DB:          db,
    })

    // 注册中间件
    app.Use(wellhttp.Logger())
    app.Use(wellhttp.Recovery())
    app.Use(sessionManager.Middleware())

    // 注册路由
    router.Register(app)

    // 启动服务
    if err := app.Run(":" + conf.Port); err != nil {
        log.Fatalf("启动服务失败: %v", err)
    }
}
```

### 4.2 wellhttp框架核心

#### 4.2.1 路由定义
```go
package wellhttp

import "net/http"

// Router 路由接口
type Router interface {
    GET(path string, handlers ...HandlerFunc)
    POST(path string, handlers ...HandlerFunc)
    PUT(path string, handlers ...HandlerFunc)
    DELETE(path string, handlers ...HandlerFunc)
    PATCH(path string, handlers ...HandlerFunc)
    Group(prefix string, handlers ...HandlerFunc) *RouterGroup
    Use(middleware ...HandlerFunc)
}

// HandlerFunc 处理器函数类型
type HandlerFunc func(c *Context)

// Context 请求上下文
type Context struct {
    Writer     ResponseWriter
    Request    *http.Request
    Params     map[string]string
    Handlers   []HandlerFunc
    Index      int
    Session    Session
    Cache      Cache
    DB         *gorm.DB
    // 模板数据
    Data       map[string]interface{}
}

// Engine 框架引擎
type Engine struct {
    router        *router
    groups        []*RouterGroup
    session       SessionManager
    cache         CacheManager
    templateRender *TemplateRender
}

// Run 启动服务
func (e *Engine) Run(addr string) error {
    return http.ListenAndServe(addr, e)
}
```

#### 4.2.2 上下文封装
```go
package wellhttp

import (
    "encoding/json"
    "html/template"
    "net/http"
    "path/filepath"
)

// JSON 返回JSON响应
func (c *Context) JSON(code int, data interface{}) {
    c.Writer.Header().Set("Content-Type", "application/json")
    c.Writer.WriteHeader(code)
    json.NewEncoder(c.Writer).Encode(data)
}

// HTML 返回HTML响应
func (c *Context) HTML(code int, name string, data interface{}) {
    c.Writer.Header().Set("Content-Type", "text/html")
    c.Writer.WriteHeader(code)
    c.templateRender.Execute(c.Writer, name, data)
}

// Param 获取URL参数
func (c *Context) Param(key string) string {
    return c.Params[key]
}

// Query 获取Query参数
func (c *Context) Query(key string) string {
    return c.Request.URL.Query().Get(key)
}

// PostForm 获取Post表单参数
func (c *Context) PostForm(key string) string {
    return c.Request.FormValue(key)
}

// Session 获取会话
func (c *Context) Session() Session {
    return c.Session
}

// Cache 获取缓存
func (c *Context) Cache() Cache {
    return c.Cache
}

// DB 获取数据库
func (c *Context) DB() *gorm.DB {
    return c.DB
}
```

### 4.3 路由注册示例
```go
package router

import "wellcms-go/wellhttp"

// Register 注册所有路由
func Register(app *wellhttp.Engine) {
    // 前台路由
    app.GET("/", handlers.Index)
    app.GET("/read-{tid}.html", handlers.Read)
    app.GET("/list-{fid}.html", handlers.List)
    app.GET("/category-{fid}.html", handlers.Category)
    app.GET("/tag-{tagname}.html", handlers.Tag)
    app.GET("/flag-{flagid}.html", handlers.Flag)

    // 用户相关
    app.GET("/user-login.html", handlers.UserLogin)
    app.POST("/user-login.html", handlers.UserLoginPost)
    app.GET("/user-logout.html", handlers.UserLogout)
    app.GET("/user-register.html", handlers.UserRegister)
    app.POST("/user-register.html", handlers.UserRegisterPost)
    app.GET("/user-center.html", handlers.UserCenter)
    app.GET("/home-{uid}.html", handlers.Home)

    // 附件
    app.GET("/attach-{aid}.html", handlers.Attach)
    app.POST("/attach-upload.html", handlers.AttachUpload)

    // 评论
    app.POST("/comment-create.html", handlers.CommentCreate)
    app.POST("/comment-delete.html", handlers.CommentDelete)

    // 操作
    app.POST("/thread-create.html", handlers.ThreadCreate)
    app.POST("/thread-update.html", handlers.ThreadUpdate)
    app.POST("/thread-delete.html", handlers.ThreadDelete)

    // API接口
    api := app.Group("/api")
    {
        api.GET("/threads", api.ThreadList)
        api.GET("/thread/{tid}", api.ThreadDetail)
        api.POST("/thread", api.ThreadCreate)
        api.GET("/forums", api.ForumList)
        api.GET("/user/{uid}", api.UserDetail)
    }

    // 后台管理
    admin := app.Group("/admin")
    admin.Use(authMiddleware)
    {
        admin.GET("/", admin.Index)
        admin.GET("/content", admin.Content)
        admin.GET("/content/create", admin.ContentCreate)
        admin.POST("/content/store", admin.ContentStore)
        admin.GET("/forum", admin.Forum)
        admin.GET("/forum/create", admin.ForumCreate)
        admin.POST("/forum/store", admin.ForumStore)
        admin.GET("/setting", admin.Setting)
        admin.POST("/setting/save", admin.SettingSave)
        // ... 更多后台路由
    }
}
```

---

## 5. 数据库设计

### 5.1 数据库概述
- **数据库引擎**: MySQL/MariaDB/PostgreSQL
- **表数量**: 22张核心表
- **表前缀**: well_ (可配置)
- **字符集**: utf8mb4

### 5.2 核心数据表

#### 5.2.1 用户相关表

**用户表 (well_user)**
```sql
CREATE TABLE `well_user` (
  `uid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户编号',
  `gid` SMALLINT(6) UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户组编号',
  `email` VARCHAR(40) NOT NULL DEFAULT '' COMMENT '邮箱',
  `username` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '用户名',
  `realname` VARCHAR(16) NOT NULL DEFAULT '' COMMENT '真实姓名',
  `idnumber` VARCHAR(19) NOT NULL DEFAULT '' COMMENT '身份证号',
  `password` VARCHAR(64) NOT NULL DEFAULT '' COMMENT '密码(BCRYPT)',
  `mobile` VARCHAR(11) NOT NULL DEFAULT '' COMMENT '手机号',
  `qq` VARCHAR(12) NOT NULL DEFAULT '' COMMENT 'QQ号',
  `articles` INT(11) NOT NULL DEFAULT '0' COMMENT '文章数',
  `comments` INT(11) NOT NULL DEFAULT '0' COMMENT '评论数',
  `credits` INT(11) NOT NULL DEFAULT '0' COMMENT '积分',
  `golds` INT(11) NOT NULL DEFAULT '0' COMMENT '金币',
  `money` DECIMAL(11,2) NOT NULL DEFAULT '0.00' COMMENT '钱包',
  `create_ip` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建IP',
  `create_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间',
  `login_ip` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '登录IP',
  `login_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '登录时间',
  `logins` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '登录次数',
  `avatar` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '头像时间戳',
  PRIMARY KEY (`uid`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_email` (`email`),
  KEY `idx_gid` (`gid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**用户组表 (well_user_group)**
```sql
CREATE TABLE `well_user_group` (
  `gid` SMALLINT(6) UNSIGNED NOT NULL COMMENT '用户组编号',
  `name` VARCHAR(20) NOT NULL DEFAULT '' COMMENT '用户组名称',
  `creditsfrom` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '积分从',
  `creditsto` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '积分到',
  `allowread` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许访问',
  `allowthread` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许发主题',
  `allowpost` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许回复',
  `allowattach` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许上传',
  `allowdown` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许下载',
  `allowtop` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许置顶',
  `allowupdate` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许编辑',
  `allowdelete` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许删除',
  `allowmove` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '允许移动',
  `allowbanuser` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '禁止用户',
  `allowdeleteuser` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '删除用户',
  `allowviewip` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '查看IP',
  `intoadmin` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '进后台',
  `managecontent` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理内容',
  `managesticky` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理置顶',
  `managecomment` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理评论',
  `manageforum` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理版块',
  `manageuser` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理用户',
  `managegroup` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理用户组',
  `manageplugin` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理插件',
  `manageother` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '管理其他',
  `managesetting` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '系统设置',
  PRIMARY KEY (`gid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**预置用户组**:
- gid=0: 游客组
- gid=1: 管理员组 (完全权限)
- gid=2: 超级版主组
- gid=4: 版主组
- gid=5: 实习版主组
- gid=6: 待验证用户组
- gid=7: 禁止用户组
- gid=101~105: 一级至五级用户组 (积分等级)

#### 5.2.2 版块相关表

**版块表 (well_forum)**
```sql
CREATE TABLE `well_forum` (
  `fid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `fup` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '上级栏目fid',
  `son` INT(11) NOT NULL DEFAULT '0' COMMENT '子栏目数',
  `type` TINYINT(1) UNSIGNED NOT NULL DEFAULT '1' COMMENT '分类:0论坛1cms',
  `model` TINYINT(2) UNSIGNED NOT NULL DEFAULT '0' COMMENT '模型:0文章2下载3咨询4视频5商城',
  `category` TINYINT(2) UNSIGNED NOT NULL DEFAULT '0' COMMENT '版块分类:0列表1频道2单页3外链',
  `name` VARCHAR(24) NOT NULL DEFAULT '' COMMENT '版块名称',
  `rank` TINYINT(3) UNSIGNED NOT NULL DEFAULT '0' COMMENT '排序',
  `threads` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '主题数',
  `tops` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '置顶数',
  `todayposts` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '今日发帖',
  `todaythreads` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '今日发主题',
  `accesson` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '开启权限控制',
  `orderby` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '列表排序:0顶贴时间1发帖时间',
  `icon` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '图标时间戳',
  `display` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '首页显示',
  `nav_display` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '导航显示',
  `index_new` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '首页显示数量',
  `channel_new` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '频道显示数量',
  `comment` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '评论开启',
  `pagesize` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '列表显示数量',
  `flags` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '属性数量',
  `create_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间',
  `flagstr` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '属性字串',
  `thumbnail` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '缩略图像素',
  `moduids` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '版主UID',
  `seo_title` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'SEO标题',
  `seo_keywords` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'SEO关键词',
  `brief` TEXT COMMENT '版块简介',
  `announcement` TEXT COMMENT '版块公告',
  PRIMARY KEY (`fid`),
  KEY `idx_fup` (`fup`),
  KEY `idx_type` (`type`),
  KEY `idx_rank` (`rank`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 5.2.3 主题相关表

**主题表 (well_thread)**
```sql
CREATE TABLE `well_thread` (
  `tid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主题ID',
  `fid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '版块ID',
  `type` TINYINT(2) UNSIGNED NOT NULL DEFAULT '0' COMMENT '主题类型:0默认10外链11单页',
  `sticky` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '置顶级别:0普通1-3置顶',
  `uid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `icon` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '缩略图时间戳',
  `userip` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '发表IP',
  `create_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '发帖时间',
  `views` INT(11) NOT NULL DEFAULT '0' COMMENT '查看次数',
  `posts` INT(11) NOT NULL DEFAULT '0' COMMENT '回复数',
  `images` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '图片数',
  `files` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '文件数',
  `mods` TINYINT(3) NOT NULL DEFAULT '0' COMMENT '版主操作次数',
  `status` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '状态:0通过1待审核2草稿10退稿11删除20下架',
  `closed` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '关闭:1关闭回复2关闭编辑',
  `lastuid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '最近参与用户',
  `last_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '最后回复时间',
  `attach_on` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '附件储存方式',
  `flags` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '绑定flag数量',
  `subject` VARCHAR(128) NOT NULL DEFAULT '' COMMENT '主题标题',
  `tag` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '标签JSON',
  `brief` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '简介',
  `keyword` VARCHAR(64) NOT NULL DEFAULT '' COMMENT 'SEO关键词',
  `description` VARCHAR(120) NOT NULL DEFAULT '' COMMENT 'SEO描述',
  `image_url` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '图床URL',
  PRIMARY KEY (`tid`),
  KEY `idx_fid` (`fid`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`),
  KEY `idx_create_date` (`create_date`),
  KEY `idx_last_date` (`last_date`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**主题数据表 (well_thread_data)**
```sql
CREATE TABLE `well_thread_data` (
  `tid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `doctype` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0',
  `attach_on` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0',
  `message` LONGTEXT NOT NULL COMMENT '主题内容',
  PRIMARY KEY (`tid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 5.2.4 评论相关表

**评论表 (well_comment)**
```sql
CREATE TABLE `well_comment` (
  `pid` INT(11) UNSIGNED NOT NULL AUTO_INCREMENT,
  `fid` INT(11) UNSIGNED NOT NULL DEFAULT '0',
  `tid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '主题ID',
  `uid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `status` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '状态',
  `create_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '创建时间',
  `userip` BIGINT UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户IP',
  `doctype` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '文档类型',
  `quotepid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '引用PID',
  `images` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '图片数',
  `files` TINYINT(2) NOT NULL DEFAULT '0' COMMENT '文件数',
  `attach_on` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '附件储存',
  `message` LONGTEXT NOT NULL COMMENT '评论内容',
  PRIMARY KEY (`pid`),
  KEY `idx_tid` (`tid`),
  KEY `idx_uid` (`uid`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

#### 5.2.5 其他核心表

**附件表 (well_attach)**
```sql
CREATE TABLE `well_attach` (
  `aid` BIGINT(20) UNSIGNED NOT NULL AUTO_INCREMENT,
  `tid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '主题ID',
  `pid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '评论ID',
  `uid` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '用户ID',
  `filesize` INT(8) UNSIGNED NOT NULL DEFAULT '0' COMMENT '文件大小',
  `width` MEDIUMINT(8) UNSIGNED NOT NULL DEFAULT '0' COMMENT '宽度',
  `height` MEDIUMINT(8) UNSIGNED NOT NULL DEFAULT '0' COMMENT '高度',
  `downloads` INT(11) NOT NULL DEFAULT '0' COMMENT '下载次数',
  `credits` INT(11) NOT NULL DEFAULT '0' COMMENT '所需积分',
  `golds` INT(11) NOT NULL DEFAULT '0' COMMENT '所需金币',
  `money` INT(11) NOT NULL DEFAULT '0' COMMENT '所需金钱',
  `isimage` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '是否图片',
  `attach_on` TINYINT(1) NOT NULL DEFAULT '0' COMMENT '储存方式:0本地1云储存2图床',
  `create_date` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '上传时间',
  `filename` VARCHAR(60) NOT NULL DEFAULT '' COMMENT '文件名',
  `orgfilename` VARCHAR(80) NOT NULL DEFAULT '' COMMENT '原文件名',
  `image_url` VARCHAR(120) NOT NULL DEFAULT '' COMMENT '云储存URL',
  `filetype` VARCHAR(7) NOT NULL DEFAULT '' COMMENT '文件类型',
  `comment` VARCHAR(100) NOT NULL DEFAULT '' COMMENT '文件注释',
  PRIMARY KEY (`aid`),
  KEY `idx_tid` (`tid`),
  KEY `idx_pid` (`pid`),
  KEY `idx_uid` (`uid`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

**KV缓存表 (well_kv)**
```sql
CREATE TABLE `well_kv` (
  `k` VARCHAR(32) NOT NULL DEFAULT '' COMMENT '键',
  `v` MEDIUMTEXT NOT NULL COMMENT '值',
  `expiry` INT(11) UNSIGNED NOT NULL DEFAULT '0' COMMENT '过期时间',
  PRIMARY KEY(`k`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

### 5.3 ER关系图
```
┌─────────────┐       ┌─────────────┐
│   user      │──────▶│ user_group  │
│ (用户表)     │ gid   │ (用户组表)   │
└─────────────┘       └─────────────┘
       │
       │ uid
       ▼
┌─────────────┐
│   session   │
│  (会话表)    │
└─────────────┘
       │
       │ tid
       ▼
┌─────────────┐       ┌─────────────┐
│   thread    │◀──────│    forum    │
│  (主题表)    │ fid   │  (版块表)    │
└─────────────┘       └─────────────┘
       │                     ▲
       │ tid                 │ fid
       ▼                     │
┌─────────────┐       ┌─────────────┐
│thread_data  │       │forum_access │
│  (数据表)    │       │(版块权限表)  │
└─────────────┘       └─────────────┘
       │
       │ tid
       ▼
┌─────────────┐       ┌─────────────┐
│  comment    │       │    flag     │
│  (评论表)    │       │  (属性表)    │
└─────────────┘       └─────────────┘
       │ pid                │ flagid
       ▼                    ▼
┌─────────────┐       ┌─────────────┐
│comment_pid  │       │flag_thread  │
│ (评论回复表) │       │(属性主题关联)│
└─────────────┘       └─────────────┘
       │
       │ tagid
       ▼
┌─────────────┐       ┌─────────────┐
│    tag      │       │   attach    │
│  (标签表)    │       │  (附件表)    │
└─────────────┘       └─────────────┘
       │
       │ tagid
       ▼
┌─────────────┐
│ tag_thread  │
│(标签主题关联)│
└─────────────┘
```

---

## 6. 路由系统

### 6.1 路由配置
```go
package router

import (
    "wellcms-go/internal/handlers"
    "wellcms-go/internal/api"
    "wellcms-go/internal/middleware"
    "wellcms-go/wellhttp"
)

// Register 注册所有路由
func Register(app *wellhttp.Engine) {
    // 首页
    app.GET("/", handlers.Index)

    // 主题相关
    app.GET("/read-{tid}.html", handlers.Read)
    app.GET("/list-{fid}.html", handlers.List)
    app.GET("/category-{fid}.html", handlers.Category)
    app.GET("/tag-{tagname}.html", handlers.Tag)
    app.GET("/flag-{flagid}.html", handlers.Flag)

    // 用户相关
    app.GET("/user-login.html", handlers.UserLogin)
    app.POST("/user-login.html", handlers.UserLoginPost)
    app.GET("/user-logout.html", handlers.UserLogout)
    app.GET("/user-register.html", handlers.UserRegister)
    app.POST("/user-register.html", handlers.UserRegisterPost)
    app.GET("/user-center.html", handlers.UserCenter)
    app.GET("/home-{uid}.html", handlers.Home)

    // 附件
    app.GET("/attach-{aid}.html", handlers.Attach)
    app.POST("/attach-upload.html", handlers.AttachUpload)

    // 评论
    app.POST("/comment-create.html", handlers.CommentCreate)
    app.POST("/comment-delete.html", handlers.CommentDelete)

    // 操作
    app.POST("/thread-create.html", handlers.ThreadCreate)
    app.POST("/thread-update.html", handlers.ThreadUpdate)
    app.POST("/thread-delete.html", handlers.ThreadDelete)

    // API接口
    apiGroup := app.Group("/api")
    {
        apiGroup.GET("/threads", api.ThreadList)
        apiGroup.GET("/thread/{tid}", api.ThreadDetail)
        apiGroup.POST("/thread", api.ThreadCreate)
        apiGroup.GET("/forums", api.ForumList)
        apiGroup.GET("/user/{uid}", api.UserDetail)
    }

    // 后台管理
    adminGroup := app.Group("/admin")
    adminGroup.Use(middleware.AuthRequired())
    {
        adminGroup.GET("/", admin.Index)
        adminGroup.GET("/content", admin.Content)
        adminGroup.GET("/content/create", admin.ContentCreate)
        adminGroup.POST("/content/store", admin.ContentStore)
        adminGroup.GET("/forum", admin.Forum)
        adminGroup.GET("/forum/create", admin.ForumCreate)
        adminGroup.POST("/forum/store", admin.ForumStore)
        adminGroup.GET("/setting", admin.Setting)
        adminGroup.POST("/setting/save", admin.SettingSave)
    }
}
```

### 6.2 路由文件说明

| 路由文件 | URL格式 | 功能说明 |
|---------|---------|---------|
| Index | / | 首页，支持3种展示模式 |
| Read | /read-{tid}.html | 主题详情页 |
| List | /list-{fid}.html | 版块主题列表 |
| Category | /category-{fid}.html | 频道页 |
| Tag | /tag-{tagname}.html | 标签页 |
| Flag | /flag-{flagid}.html | 属性页 |
| User | /user-*.html | 用户中心 |
| Home | /home-{uid}.html | 个人主页 |
| Attach | /attach-{aid}.html | 附件处理 |

### 6.3 URL格式
支持4种伪静态格式：
```
格式0: ?user-login.html
格式1: user-login.html
格式2: /user/login.html
格式3: /user/login
```

---

## 7. 模型层

### 7.1 数据模型定义

#### 7.1.1 用户模型
```go
package models

import (
    "time"
)

// User 用户模型
type User struct {
    UID         int       `json:"uid" gorm:"primaryKey;autoIncrement"`
    GID         int       `json:"gid" gorm:"index;default:0"`
    Email       string    `json:"email" gorm:"size:40;uniqueIndex"`
    Username    string    `json:"username" gorm:"size:32;uniqueIndex"`
    Realname    string    `json:"realname" gorm:"size:16"`
    IDNumber    string    `json:"idnumber" gorm:"size:19"`
    Password    string    `json:"-" gorm:"size:64"`
    Mobile      string    `json:"mobile" gorm:"size:11"`
    QQ          string    `json:"qq" gorm:"size:12"`
    Articles    int       `json:"articles" gorm:"default:0"`
    Comments    int       `json:"comments" gorm:"default:0"`
    Credits     int       `json:"credits" gorm:"default:0"`
    Golds       int       `json:"golds" gorm:"default:0"`
    Money       float64   `json:"money" gorm:"type:decimal(11,2);default:0"`
    CreateIP    uint64    `json:"create_ip" gorm:"default:0"`
    CreateDate  int       `json:"create_date" gorm:"index"`
    LoginIP     uint64    `json:"login_ip" gorm:"default:0"`
    LoginDate   int       `json:"login_date" gorm:"index"`
    Logins      int       `json:"logins" gorm:"default:0"`
    Avatar      int       `json:"avatar" gorm:"default:0"`
    CreatedAt   time.Time `json:"-"`
    UpdatedAt   time.Time `json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
    return "well_user"
}
```

#### 7.1.2 主题模型
```go
package models

import "time"

// Thread 主题模型
type Thread struct {
    TID         int       `json:"tid" gorm:"primaryKey;autoIncrement"`
    FID         int       `json:"fid" gorm:"index;default:0"`
    Type        int       `json:"type" gorm:"default:0"`
    Sticky      int       `json:"sticky" gorm:"default:0"`
    UID         int       `json:"uid" gorm:"index;default:0"`
    Icon        int       `json:"icon" gorm:"default:0"`
    UserIP      uint64    `json:"userip" gorm:"default:0"`
    CreateDate  int       `json:"create_date" gorm:"index"`
    Views       int       `json:"views" gorm:"default:0"`
    Posts       int       `json:"posts" gorm:"default:0"`
    Images      int       `json:"images" gorm:"default:0"`
    Files       int       `json:"files" gorm:"default:0"`
    Mods        int       `json:"mods" gorm:"default:0"`
    Status      int       `json:"status" gorm:"index;default:0"`
    Closed      int       `json:"closed" gorm:"default:0"`
    LastUID     int       `json:"lastuid" gorm:"default:0"`
    LastDate    int       `json:"last_date" gorm:"index"`
    AttachOn    int       `json:"attach_on" gorm:"default:0"`
    Flags       int       `json:"flags" gorm:"default:0"`
    Subject     string    `json:"subject" gorm:"size:128"`
    Tag         string    `json:"tag" gorm:"type:text"`
    Brief       string    `json:"brief" gorm:"size:120"`
    Keyword     string    `json:"keyword" gorm:"size:64"`
    Description string    `json:"description" gorm:"size:120"`
    ImageURL    string    `json:"image_url" gorm:"size:120"`
    CreatedAt   time.Time `json:"-"`
    UpdatedAt   time.Time `json:"-"`
}

// TableName 指定表名
func (Thread) TableName() string {
    return "well_thread"
}

// ThreadData 主题数据模型
type ThreadData struct {
    TID       int    `json:"tid" gorm:"primaryKey;autoIncrement"`
    Doctype   int    `json:"doctype" gorm:"default:0"`
    AttachOn  int    `json:"attach_on" gorm:"default:0"`
    Message   string `json:"message" gorm:"type:longtext"`
}

// TableName 指定表名
func (ThreadData) TableName() string {
    return "well_thread_data"
}
```

#### 7.1.3 版块模型
```go
package models

import "time"

// Forum 版块模型
type Forum struct {
    FID          int       `json:"fid" gorm:"primaryKey;autoIncrement"`
    FUP          int       `json:"fup" gorm:"index;default:0"`
    Son          int       `json:"son" gorm:"default:0"`
    Type         int       `json:"type" gorm:"default:1"`
    Model        int       `json:"model" gorm:"default:0"`
    Category     int       `json:"category" gorm:"default:0"`
    Name         string    `json:"name" gorm:"size:24"`
    Rank         int       `json:"rank" gorm:"default:0"`
    Threads      int       `json:"threads" gorm:"default:0"`
    Tops         int       `json:"tops" gorm:"default:0"`
    TodayPosts   int       `json:"todayposts" gorm:"default:0"`
    TodayThreads int       `json:"todaythreads" gorm:"default:0"`
    AccessOn     int       `json:"accesson" gorm:"default:0"`
    OrderBy      int       `json:"orderby" gorm:"default:0"`
    Icon         int       `json:"icon" gorm:"default:0"`
    Display      int       `json:"display" gorm:"default:0"`
    NavDisplay   int       `json:"nav_display" gorm:"default:0"`
    IndexNew     int       `json:"index_new" gorm:"default:0"`
    ChannelNew   int       `json:"channel_new" gorm:"default:0"`
    Comment      int       `json:"comment" gorm:"default:0"`
    PageSize     int       `json:"pagesize" gorm:"default:0"`
    Flags        int       `json:"flags" gorm:"default:0"`
    CreateDate   int       `json:"create_date" gorm:"index"`
    FlagStr      string    `json:"flagstr" gorm:"size:120"`
    Thumbnail    string    `json:"thumbnail" gorm:"size:32"`
    ModUids      string    `json:"moduids" gorm:"size:120"`
    SEOTitle     string    `json:"seo_title" gorm:"size:64"`
    SEOKeywords  string    `json:"seo_keywords" gorm:"size:64"`
    Brief        string    `json:"brief" gorm:"type:text"`
    Announcement string    `json:"announcement" gorm:"type:text"`
    CreatedAt    time.Time `json:"-"`
    UpdatedAt    time.Time `json:"-"`
}

// TableName 指定表名
func (Forum) TableName() string {
    return "well_forum"
}
```

### 7.2 模型操作封装

#### 7.2.1 用户模型操作
```go
package models

import (
    "errors"
    "golang.org/x/crypto/bcrypt"
)

// UserService 用户服务
type UserService struct {
    DB *gorm.DB
}

// Create 创建用户
func (s *UserService) Create(user *User) error {
    // 密码加密
    hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    user.Password = string(hash)

    return s.DB.Create(user).Error
}

// VerifyPassword 验证密码
func (s *UserService) VerifyPassword(user *User, password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    return err == nil
}

// GetByUsername 根据用户名获取用户
func (s *UserService) GetByUsername(username string) (*User, error) {
    var user User
    err := s.DB.Where("username = ?", username).First(&user).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// GetByUID 根据UID获取用户
func (s *UserService) GetByUID(uid int) (*User, error) {
    var user User
    err := s.DB.First(&user, uid).Error
    if err != nil {
        return nil, err
    }
    return &user, nil
}

// Find 查找用户列表
func (s *UserService) Find(cond *User, page, pageSize int, orderBy string) ([]User, int64) {
    var users []User
    var total int64

    query := s.DB.Model(&User{})
    if cond.GID > 0 {
        query = query.Where("gid = ?", cond.GID)
    }

    query.Count(&total)
    query = query.Order(orderBy)
    query = query.Offset((page - 1) * pageSize).Limit(pageSize)
    query.Find(&users)

    return users, total
}

// Update 更新用户
func (s *UserService) Update(uid int, updates map[string]interface{}) error {
    return s.DB.Model(&User{}).Where("uid = ?", uid).Updates(updates).Error
}
```

#### 7.2.2 主题模型操作
```go
package models

import (
    "errors"
)

// ThreadService 主题服务
type ThreadService struct {
    DB *gorm.DB
}

// Create 创建主题
func (s *ThreadService) Create(thread *Thread, data *ThreadData) error {
    return s.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(thread).Error; err != nil {
            return err
        }
        data.TID = thread.TID
        if err := tx.Create(data).Error; err != nil {
            return err
        }
        return nil
    })
}

// GetByTID 根据TID获取主题
func (s *ThreadService) GetByTID(tid int) (*Thread, error) {
    var thread Thread
    err := s.DB.First(&thread, tid).Error
    if err != nil {
        return nil, err
    }
    return &thread, nil
}

// GetThreadData 获取主题内容
func (s *ThreadService) GetThreadData(tid int) (*ThreadData, error) {
    var data ThreadData
    err := s.DB.First(&data, tid).Error
    if err != nil {
        return nil, err
    }
    return &data, nil
}

// Find 查找主题列表
func (s *ThreadService) Find(cond *Thread, page, pageSize int, orderBy string) ([]Thread, int64) {
    var threads []Thread
    var total int64

    query := s.DB.Model(&Thread{})
    if cond.FID > 0 {
        query = query.Where("fid = ?", cond.FID)
    }
    if cond.Status >= 0 {
        query = query.Where("status = ?", cond.Status)
    }

    query.Count(&total)
    query = query.Order(orderBy)
    query = query.Offset((page - 1) * pageSize).Limit(pageSize)
    query.Find(&threads)

    return threads, total
}

// Update 更新主题
func (s *ThreadService) Update(tid int, updates map[string]interface{}) error {
    return s.DB.Model(&Thread{}).Where("tid = ?", tid).Updates(updates).Error
}

// Delete 删除主题
func (s *ThreadService) Delete(tid int) error {
    return s.DB.Transaction(func(tx *gorm.DB) error {
        if err := tx.Where("tid = ?", tid).Delete(&Thread{}).Error; err != nil {
            return err
        }
        if err := tx.Where("tid = ?", tid).Delete(&ThreadData{}).Error; err != nil {
            return err
        }
        return nil
    })
}

// IncrView 增加浏览次数
func (s *ThreadService) IncrView(tid int) error {
    return s.DB.Model(&Thread{}).Where("tid = ?", tid).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
```

### 7.3 缓存函数

#### 7.3.1 KV缓存
```go
package cache

import (
    "encoding/json"
    "time"
)

// KVService KV缓存服务
type KVService struct {
    Cache Cache
    DB    *gorm.DB
}

// Set 设置KV
func (s *KVService) Set(key string, value interface{}, expiry time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    // 先写入缓存
    if expiry > 0 {
        s.Cache.Set("kv_"+key, data, expiry)
    }

    // 写入数据库
    expiryTime := time.Now().Add(expiry).Unix()
    if expiry <= 0 {
        expiryTime = 0
    }

    return s.DB.Exec("INSERT INTO well_kv (k, v, expiry) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE v = ?, expiry = ?",
        key, string(data), expiryTime, string(data), expiryTime).Error
}

// Get 获取KV
func (s *KVService) Get(key string, value interface{}) error {
    // 先从缓存获取
    data, err := s.Cache.Get("kv_" + key)
    if err == nil && data != nil {
        return json.Unmarshal(data.([]byte), value)
    }

    // 从数据库获取
    var result struct {
        V      string `json:"v"`
        Expiry int64  `json:"expiry"`
    }
    err = s.DB.Raw("SELECT v, expiry FROM well_kv WHERE k = ?", key).Scan(&result).Error
    if err != nil {
        return err
    }

    // 检查是否过期
    if result.Expiry > 0 && result.Expiry < time.Now().Unix() {
        return errors.New("key expired")
    }

    // 写入缓存
    if result.Expiry > 0 {
        s.Cache.Set("kv_"+key, []byte(result.V), time.Duration(result.Expiry-time.Now().Unix())*time.Second)
    }

    return json.Unmarshal([]byte(result.V), value)
}

// Delete 删除KV
func (s *KVService) Delete(key string) error {
    s.Cache.Delete("kv_" + key)
    return s.DB.Exec("DELETE FROM well_kv WHERE k = ?", key).Error
}
```

---

## 8. 插件机制

### 8.1 插件机制概述
WellCMS Go采用插件化架构，支持中间件注入、处理器覆盖和事件钩子。

### 8.2 插件文件结构
```
plugin/
└── 插件名/
    ├── conf.json               # 插件配置
    ├── main.go                 # 插件入口
    ├── hooks/                  # 钩子文件目录
    │   ├── on_user_create.go
    │   ├── on_thread_create.go
    │   └── on_request.go
    ├── handlers/               # 覆盖处理器
    │   └── custom_handler.go
    └── middleware/             # 插件中间件
        └── custom_middleware.go
```

### 8.3 插件配置示例
```json
{
    "name": "插件名称",
    "brief": "插件简介",
    "version": "1.0",
    "author": "作者",
    "hooks": [
        "on_user_create",
        "on_thread_create"
    ],
    "dependencies": []
}
```

### 8.4 钩子机制

#### 8.4.1 定义钩子点
```go
package wellhttp

import (
    "reflect"
)

// HookRegistry 钩子注册表
type HookRegistry struct {
    hooks map[string][]HookFunc
}

// HookFunc 钩子函数类型
type HookFunc func(c *Context) error

// RegisterHook 注册钩子
func (r *HookRegistry) RegisterHook(name string, fn HookFunc) {
    r.hooks[name] = append(r.hooks[name], fn)
}

// CallHooks 调用钩子
func (r *HookRegistry) CallHooks(name string, c *Context) error {
    if hooks, ok := r.hooks[name]; ok {
        for _, hook := range hooks {
            if err := hook(c); err != nil {
                return err
            }
        }
    }
    return nil
}
```

#### 8.4.2 在业务中定义钩子点
```go
package models

import (
    "wellcms-go/internal/hooks"
)

// UserService 用户服务
type UserService struct {
    DB *gorm.DB
}

// Create 创建用户
func (s *UserService) Create(user *User) error {
    // 调用创建前钩子
    if err := hooks.Call("user_create_before", user); err != nil {
        return err
    }

    if err := s.DB.Create(user).Error; err != nil {
        return err
    }

    // 调用创建后钩子
    if err := hooks.Call("user_create_after", user); err != nil {
        return err
    }

    return nil
}
```

### 8.5 插件实现示例
```go
package myplugin

import (
    "wellcms-go/internal/hooks"
    "wellcms-go/internal/models"
)

// Init 插件初始化
func Init() {
    // 注册钩子
    hooks.Register("user_create_before", UserCreateBefore)
    hooks.Register("user_create_after", UserCreateAfter)
}

// UserCreateBefore 用户创建前钩子
func UserCreateBefore(c *hooks.Context) error {
    user := c.Data["user"].(*models.User)
    // 添加额外逻辑
    user.Credits = 100 // 初始积分100
    return nil
}

// UserCreateAfter 用户创建后钩子
func UserCreateAfter(c *hooks.Context) error {
    user := c.Data["user"].(*models.User)
    // 记录用户创建日志
    // ...
    return nil
}
```

---

## 9. 视图模板

### 9.1 模板文件结构
```
view/
├── htm/                        # 模板文件
│   ├── 404.htm
│   ├── browser.htm
│   ├── comment.htm
│   ├── comment_list.inc.htm
│   ├── common.header.htm
│   ├── common.template.htm
│   ├── flag.htm
│   ├── flat.htm
│   ├── flat_category.htm
│   ├── footer.inc.htm
│   ├── footer_nav.inc.htm
│   ├── header.inc.htm
│   ├── header_nav.inc.htm
│   ├── home_article.htm
│   └── ...
└── template/                   # 主题目录
    └── demo/                   # 示例主题
        ├── conf.json
        └── ...
```

### 9.2 模板引擎封装
```go
package template

import (
    "html/template"
    "io"
    "wellcms-go/internal/config"
)

// Engine 模板引擎
type Engine struct {
    dir       string
    theme     string
    funcs     template.FuncMap
    templates *template.Template
}

// New 创建模板引擎
func New(conf *config.Config) *Engine {
    e := &Engine{
        dir:   conf.AppPath + "view/htm/",
        theme: conf.Theme,
    }

    e.funcs = template.FuncMap{
        "format_date":  e.formatDate,
        "safe_html":    e.safeHTML,
        "pagination":   e.pagination,
    }

    e.loadTemplates()
    return e
}

// loadTemplates 加载模板
func (e *Engine) loadTemplates() {
    e.templates = template.New("").Funcs(e.funcs)
    e.templates = template.Must(e.templates.ParseGlob(e.dir + "*.htm"))
}

// Execute 执行模板
func (e *Engine) Execute(wr io.Writer, name string, data interface{}) error {
    return e.templates.ExecuteTemplate(wr, name+".htm", data)
}
```

### 9.3 模板语法

#### 9.3.1 变量输出
```htm
{$title}
{$thread.subject}
{$forumlist[$fid].name}
```

#### 9.3.2 条件判断
```htm
{{if .User}}
    欢迎 {$User.Username}
{{else}}
    请登录
{{end}}
```

#### 9.3.3 循环遍历
```htm
{{range $key, $thread := .ThreadList}}
    <div class="thread-item">
        <a href="/read-{{$thread.Tid}}.html">{{$thread.Subject}}</a>
    </div>
{{end}}
```

#### 9.3.4 包含文件
```htm
{{include "header.inc.htm" .}}
{{include "footer.inc.htm" .}}
```

### 9.4 模板变量

#### 9.4.1 全局变量
```go
type GlobalData struct {
    Conf      *config.Config
    User      *models.User
    GID       int
    GroupList []*models.UserGroup
    ForumList []*models.Forum
    Runtime   *RuntimeData
    Header    map[string]string
}
```

#### 9.4.2 首页变量
```go
type IndexData struct {
    WebsiteMode  int
    TplMode      int
    ThreadList   []*models.Thread
    FlagList     []*models.Flag
    StickyList   []*models.Thread
    LinkList     []*models.Link
    Pagination   *Pagination
}
```

#### 9.4.3 详情页变量
```go
type ReadData struct {
    Thread     *models.Thread
    Data       *models.ThreadData
    Forum      *models.Forum
    AttachList []*models.Attach
    ImageList  []*models.Attach
    PostList   []*models.Comment
    Pagination *Pagination
    Access     *AccessInfo
}
```

---

## 10. 后台管理

### 10.1 后台入口
```
/admin/
```

### 10.2 后台结构
```
internal/admin/
├── handlers/              # 后台处理器
│   ├── index.go          # 首页
│   ├── content.go        # 内容管理
│   ├── column.go         # 栏目管理
│   ├── flag.go           # 属性管理
│   ├── template.go       # 模板管理
│   ├── comment.go        # 评论管理
│   ├── sticky.go         # 置顶管理
│   ├── page.go           # 单页管理
│   ├── group.go          # 用户组管理
│   ├── setting.go        # 系统设置
│   ├── other.go          # 其他设置
│   ├── user.go           # 用户管理
│   ├── style.go          # 风格管理
│   └── plugin.go         # 插件管理
├── middleware/           # 后台中间件
│   └── auth.go           # 认证中间件
└── views/                # 后台视图
    └── htm/              # 后台模板
```

### 10.3 后台认证中间件
```go
package middleware

import (
    "net/http"
    "wellcms-go/wellhttp"
    "wellcms-go/internal/models"
)

// AuthRequired 需要登录
func AuthRequired() wellhttp.HandlerFunc {
    return func(c *wellhttp.Context) error {
        user := c.Session().Get("user")
        if user == nil {
            c.Redirect("/user-login.html")
            return nil
        }

        // 检查是否是管理员
        userModel, ok := user.(*models.User)
        if !ok || userModel.GID != 1 {
            return c.String(http.StatusForbidden, "无权限访问")
        }

        c.Set("current_user", user)
        return c.Next()
    }
}
```

### 10.4 后台处理器示例
```go
package admin

import (
    "net/http"
    "wellcms-go/wellhttp"
    "wellcms-go/internal/models"
)

// Index 后台首页
func Index(c *wellhttp.Context) error {
    var stats struct {
        UserCount    int64
        ThreadCount  int64
        CommentCount int64
        TodayThreads int64
    }

    c.DB().Model(&models.User{}).Count(&stats.UserCount)
    c.DB().Model(&models.Thread{}).Count(&stats.ThreadCount)
    c.DB().Model(&models.Comment{}).Count(&stats.CommentCount)

    today := intodayUnix()
    c.DB().Model(&models.Thread{}).Where("create_date >= ?", today).Count(&stats.TodayThreads)

    return c.HTML(http.StatusOK, "admin/index", stats)
}

// Content 内容管理
func Content(c *wellhttp.Context) error {
    page := c.QueryInt("page", 1)
    pageSize := 20

    var threads []*models.Thread
    var total int64

    c.DB().Model(&models.Thread{}).Count(&total)
    c.DB().Order("tid DESC").Offset((page-1)*pageSize).Limit(pageSize).Find(&threads)

    return c.HTML(http.StatusOK, "admin/content", map[string]interface{}{
        "ThreadList": threads,
        "Pagination": NewPagination(page, pageSize, int(total)),
    })
}
```

---

## 11. 配置系统

### 11.1 配置文件结构
```yaml
# WellCMS Go 配置文件

app:
  name: "WellCMS Go"
  debug: false
  host: "0.0.0.0"
  port: 8080
  app_path: "./"
  admin_path: "admin"
  theme: "default"

database:
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  name: "wellcms"
  table_prefix: "well_"
  max_open_conn: 100
  max_idle_conn: 10

cache:
  type: "memory"  # memory, redis, memcached
  host: "localhost"
  port: 6379
  password: ""
  db: 0

session:
  type: "memory"  # memory, redis
  cookie_name: "wellcms_session"
  max_lifetime: 86400

upload:
  path: "./public/upload/"
  max_size: 10485760  # 10MB
  allowed_exts:
    - ".jpg"
    - ".jpeg"
    - ".png"
    - ".gif"
    - ".pdf"
    - ".zip"

smtp:
  host: "smtp.example.com"
  port: 465
  username: "noreply@example.com"
  password: "password"
  from: "noreply@example.com"

log:
  level: "info"
  path: "./log/"
  filename: "wellcms.log"
```

### 11.2 配置加载
```go
package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

// Config 配置结构
type Config struct {
    App       AppConfig       `yaml:"app"`
    DB        DBConfig        `yaml:"database"`
    Cache     CacheConfig     `yaml:"cache"`
    Session   SessionConfig   `yaml:"session"`
    Upload    UploadConfig    `yaml:"upload"`
    SMTP      SMTPConfig      `yaml:"smtp"`
    Log       LogConfig       `yaml:"log"`
}

// Load 加载配置
func Load(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var conf Config
    if err := yaml.Unmarshal(data, &conf); err != nil {
        return nil, err
    }

    // 环境变量覆盖
    conf.applyEnvOverrides()

    return &conf, nil
}

// applyEnvOverrides 应用环境变量覆盖
func (c *Config) applyEnvOverrides() {
    if host := os.Getenv("DB_HOST"); host != "" {
        c.DB.Host = host
    }
    if port := os.Getenv("APP_PORT"); port != "" {
        // 解析端口
    }
}
```

---

## 12. 缓存机制

### 12.1 缓存接口定义
```go
package cache

// Cache 缓存接口
type Cache interface {
    Get(key string) (interface{}, error)
    Set(key string, value interface{}, expiration time.Duration) error
    Delete(key string) error
    Clear() error
}
```

### 12.2 多缓存实现

#### 12.2.1 内存缓存
```go
package memory

import (
    "sync"
    "time"
)

// MemoryCache 内存缓存
type MemoryCache struct {
    mu    sync.RWMutex
    items map[string]cacheItem
}

type cacheItem struct {
    value      interface{}
    expiration int64
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    item, ok := c.items[key]
    if !ok {
        return nil, ErrCacheMiss
    }

    if item.expiration > 0 && time.Now().Unix() > item.expiration {
        delete(c.items, key)
        return nil, ErrCacheMiss
    }

    return item.value, nil
}

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}, expiration time.Duration) error {
    c.mu.Lock()
    defer c.mu.Unlock()

    var expiry int64 = 0
    if expiration > 0 {
        expiry = time.Now().Add(expiration).Unix()
    }

    c.items[key] = cacheItem{
        value:      value,
        expiration: expiry,
    }

    return nil
}
```

#### 12.2.2 Redis缓存
```go
package redis

import (
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
)

// RedisCache Redis缓存
type RedisCache struct {
    client *redis.Client
    prefix string
}

// New 创建Redis缓存
func New(addr, password string, db int, prefix string) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })

    return &RedisCache{
        client: client,
        prefix: prefix,
    }
}

// Get 获取缓存
func (c *RedisCache) Get(key string) (interface{}, error) {
    data, err := c.client.Get(c.ctx, c.prefix+key).Bytes()
    if err != nil {
        return nil, err
    }

    var result interface{}
    json.Unmarshal(data, &result)
    return result, nil
}

// Set 设置缓存
func (c *RedisCache) Set(key string, value interface{}, expiration time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }

    return c.client.Set(c.ctx, c.prefix+key, data, expiration).Err()
}
```

---

## 13. 安全机制

### 13.1 XSS防护
```go
package utils

import (
    "html"
    "regexp"
)

// XSSClean 清理XSS
func XSSClean(s string) string {
    // HTML转义
    s = html.EscapeString(s)

    // 移除危险标签
    dangerous := regexp.MustCompile(`(?i)<script.*?>.*?</script>`)
    s = dangerous.ReplaceAllString(s, "")

    dangerous = regexp.MustCompile(`(?i)javascript:`)
    s = dangerous.ReplaceAllString(s, "")

    return s
}
```

### 13.2 CSRF防护
```go
package middleware

import (
    "crypto/rand"
    "encoding/base64"
    "wellcms-go/wellhttp"
)

// CSRF CSRF中间件
func CSRF() wellhttp.HandlerFunc {
    return func(c *wellhttp.Context) error {
        // 生成CSRF Token
        token := generateToken()
        c.Session().Set("csrf_token", token)
        c.Set("csrf_token", token)

        return c.Next()
    }
}

func generateToken() string {
    b := make([]byte, 32)
    rand.Read(b)
    return base64.StdEncoding.EncodeToString(b)
}
```

### 13.3 密码加密
```go
package utils

import (
    "golang.org/x/crypto/bcrypt"
)

// HashPassword 加密密码
func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

// CheckPassword 验证密码
func CheckPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### 13.4 权限检查
```go
package middleware

import (
    "wellcms-go/wellhttp"
    "wellcms-go/internal/models"
)

// Permission 权限检查
func Permission(action string) wellhttp.HandlerFunc {
    return func(c *wellhttp.Context) error {
        user := c.Get("current_user").(*models.User)
        group, err := models.GetGroup(user.GID)
        if err != nil {
            return c.String(403, "无权限")
        }

        // 检查权限
        switch action {
        case "manage_content":
            if group.ManageContent != 1 {
                return c.String(403, "无权限")
            }
        case "manage_forum":
            if group.ManageForum != 1 {
                return c.String(403, "无权限")
            }
        // ... 其他权限
        }

        return c.Next()
    }
}
```

---

## 14. API接口

### 14.1 API路由注册
```go
package api

import (
    "wellcms-go/wellhttp"
)

// Register 注册API路由
func Register(app *wellhttp.Engine) {
    api := app.Group("/api/v1")
    {
        // 主题相关
        api.GET("/threads", ThreadList)
        api.GET("/thread/:tid", ThreadDetail)
        api.POST("/thread", ThreadCreate)

        // 版块相关
        api.GET("/forums", ForumList)
        api.GET("/forum/:fid", ForumDetail)

        // 用户相关
        api.GET("/user/:uid", UserDetail)
        api.GET("/user/:uid/threads", UserThreads)

        // 评论相关
        api.GET("/thread/:tid/comments", CommentList)
        api.POST("/comment", CommentCreate)
    }
}
```

### 14.2 API响应格式
```go
package api

import (
    "net/http"
    "wellcms-go/wellhttp"
)

// Response API响应结构
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// Success 成功响应
func Success(c *wellhttp.Context, data interface{}) error {
    return c.JSON(http.StatusOK, Response{
        Code:    0,
        Message: "success",
        Data:    data,
    })
}

// Error 错误响应
func Error(c *wellhttp.Context, code int, message string) error {
    return c.JSON(http.StatusOK, Response{
        Code:    code,
        Message: message,
    })
}
```

### 14.3 API接口示例
```go
package api

import (
    "net/http"
    "wellcms-go/wellhttp"
    "wellcms-go/internal/models"
)

// ThreadList 获取主题列表
func ThreadList(c *wellhttp.Context) error {
    page := c.QueryInt("page", 1)
    pageSize := c.QueryInt("pagesize", 20)
    fid := c.QueryInt("fid", 0)

    var cond models.Thread
    if fid > 0 {
        cond.FID = fid
    }

    threads, total := models.ThreadService.Find(&cond, page, pageSize, "tid DESC")

    return Success(c, map[string]interface{}{
        "threads": threads,
        "total":   total,
        "page":    page,
        "pagesize": pageSize,
    })
}

// ThreadDetail 获取主题详情
func ThreadDetail(c *wellhttp.Context) error {
    tid := c.ParamInt("tid", 0)

    thread, err := models.ThreadService.GetByTID(tid)
    if err != nil {
        return Error(c, 404, "主题不存在")
    }

    data, err := models.ThreadService.GetThreadData(tid)
    if err != nil {
        return Error(c, 404, "主题内容不存在")
    }

    // 增加浏览次数
    models.ThreadService.IncrView(tid)

    return Success(c, map[string]interface{}{
        "thread": thread,
        "data":   data,
    })
}
```

---

## 15. 开发规范

### 15.1 命名规范
- **文件命名**: 使用小写下划线命名 (snake_case)
- **包命名**: 使用小写下划线命名
- **结构体命名**: 使用大驼峰命名 (PascalCase)
- **变量命名**: 使用小驼峰命名 (camelCase)
- **常量命名**: 使用全大写下划线命名 (SCREAMING_SNAKE_CASE)
- **接口命名**: 使用大驼峰命名，以er结尾

### 15.2 代码格式
- 使用gofmt格式化代码
- 缩进使用Tab
- 行宽限制120字符
- 导入包分组: 标准库 → 第三方 → 内部

```go
import (
    // 标准库
    "encoding/json"
    "net/http"
    "time"

    // 第三方
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 内部
    "wellcms-go/internal/models"
    "wellcms-go/internal/config"
)
```

### 15.3 错误处理
- 使用wrap包装错误
- 在业务层返回业务错误码
- 在展现层返回用户友好信息

```go
func (s *UserService) GetByUID(uid int) (*User, error) {
    var user User
    err := s.DB.First(&user, uid).Error
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, ErrUserNotFound
        }
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }
    return &user, nil
}
```

### 15.4 注释规范
- 公共函数和结构体必须添加注释
- 注释使用完整句子
- 复杂逻辑需要详细解释

```go
// UserService 用户服务层
type UserService struct {
    DB *gorm.DB
}

// Create 创建用户
// 如果邮箱已存在，会返回 ErrEmailExists 错误
func (s *UserService) Create(user *User) error {
    // ...
}
```

### 15.5 日志规范
- 使用统一的日志格式
- 记录关键操作和错误
- 敏感信息不记录日志

```go
func (s *UserService) Login(username, password string) (*User, error) {
    user, err := s.GetByUsername(username)
    if err != nil {
        log.Warnf("用户登录失败: username=%s, error=%v", username, err)
        return nil, ErrInvalidCredentials
    }

    if !CheckPassword(password, user.Password) {
        log.Warnf("密码错误: username=%s", username)
        return nil, ErrInvalidCredentials
    }

    log.Infof("用户登录成功: uid=%d", user.UID)
    return user, nil
}
```

---

## 16. 扩展开发

### 16.1 添加新模型
1. 创建模型文件 `internal/models/xxx.go`
2. 定义结构体
3. 实现CRUD方法
4. 注册到服务容器

```go
package models

import "time"

// Link 链接模型
type Link struct {
    ID         int       `json:"id" gorm:"primaryKey;autoIncrement"`
    Rank       int       `json:"rank" gorm:"default:0"`
    Name       string    `json:"name" gorm:"size:24"`
    URL        string    `json:"url" gorm:"size:120"`
    CreateDate int       `json:"create_date" gorm:"index"`
    CreatedAt  time.Time `json:"-"`
    UpdatedAt  time.Time `json:"-"`
}

// TableName 指定表名
func (Link) TableName() string {
    return "well_link"
}

// LinkService 链接服务
type LinkService struct {
    DB *gorm.DB
}

// GetList 获取链接列表
func (s *LinkService) GetList() ([]Link, error) {
    var links []Link
    err := s.DB.Order("rank ASC").Find(&links).Error
    return links, err
}
```

### 16.2 添加新路由
1. 创建处理器文件 `internal/handlers/xxx.go`
2. 在 `internal/router/router.go` 中注册路由

```go
package handlers

import (
    "net/http"
    "wellcms-go/wellhttp"
    "wellcms-go/internal/models"
)

// Link 链接页
func Link(c *wellhttp.Context) error {
    links, err := models.LinkService.GetList()
    if err != nil {
        return c.String(http.StatusInternalServerError, "获取链接失败")
    }

    return c.HTML(http.StatusOK, "link", map[string]interface{}{
        "LinkList": links,
    })
}
```

### 16.3 添加新插件
1. 创建插件目录 `plugin/xxx/`
2. 创建 `main.go` 实现初始化函数
3. 注册钩子和中间件

```go
package myplugin

import (
    "wellcms-go/internal/hooks"
)

// Init 插件初始化
func Init() {
    // 注册钩子
    hooks.Register("thread_create_after", OnThreadCreate)
    hooks.Register("user_login_after", OnUserLogin)
}

// OnThreadCreate 主题创建后处理
func OnThreadCreate(c *hooks.Context) error {
    thread := c.Data["thread"].(*models.Thread)
    // 自定义逻辑
    return nil
}
```

### 16.4 添加后台菜单
修改 `internal/admin/menu.conf.go`：

```go
package admin

// Menu 后台菜单配置
var Menu = []MenuItem{
    {
        Name: "内容管理",
        Icon: "icon-content",
        Tab: []TabItem{
            {Name: "首页", URL: "/admin/content"},
            {Name: "置顶管理", URL: "/admin/sticky"},
            {Name: "评论管理", URL: "/admin/comment"},
        },
    },
    {
        Name: "栏目管理",
        Icon: "icon-forum",
        Tab: []TabItem{
            {Name: "栏目列表", URL: "/admin/forum"},
            {Name: "属性管理", URL: "/admin/flag"},
        },
    },
}
```

---

## 附录

### A. 常用命令
```bash
# 运行项目
go run cmd/server/main.go

# 构建项目
go build -o wellcms-go cmd/server/main.go

# 运行测试
go test ./...

# 代码格式化
gofmt -w .

# 代码检查
go vet ./...

# 依赖更新
go mod tidy
```

### B. 技术栈版本
- Go: 1.21+
- GORM: v2
- Redis: 7.x+
- MySQL: 8.0+
- Nginx: 1.20+

### C. 参考资源
- Go官方文档: https://golang.org/doc/
- GORM文档: https://gorm.io/docs/
- Gin框架: https://gin-gonic.com/docs/
