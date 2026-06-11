# Changelog

## [v0.1.1] - 2026-06-11

### Skill 市场与桌面端发布增强

本版本补齐 Skill 市场的后端 API、前端真实数据接入与安装链路，优化桌面端发布流程，并修复多用户数据隔离和构建稳定性问题。

- 新增 Skill 市场 API 与内置 Skill 数据模型，重构 `backend/skills` 为 server / worker 分层目录
- Skill 市场前端接入真实 API，支持内置 Skill 下载与安装流程端到端闭环
- BuiltinSource 改为 HTTP API 模式，短关键词外部搜索跳过逻辑下沉到 `SkillsShSource.Search`
- Skill 管理支持删除时清理 CLI symlink，并引入类型化 catalog 错误
- 修复用户、组织、项目、任务、会话、消息与 artifact 查询中的用户数据隔离问题
- 新增 `install.sh`，支持将 `leros` 注册为全局系统命令，并补充 Makefile 安装入口
- 优化桌面端发布 workflow，移除 COSCLI 上传链路，配置桌面端发布 API 地址注入
- 优化桌面端发布产物体积，修复 SkillMarketView 中 `offsetRef` 导致的构建问题

## [v0.1.0] - 2026-06-11

### SingerOS 首个可用版本

核心引擎、Worker 调度、CLI 工具链、桌面端与前端交互框架初步成型，支持用户组织管理、邮箱认证、审批工作流和 Skill 系统。

- 重构 native engine 与 system prompt 分层架构，Skill 架构升级为三层 + 事件驱动 handler 模型
- Worker 解耦数据库依赖，支持并发任务消费与重建流恢复
- 新增 User / Organization CRUD 接口，支持邮箱注册登录与令牌刷新
- CLI 命令架构重构，新增 project / task / session 的 get 子命令，支持 skill 管理与统一配置
- 桌面端发布流程打通，支持构建产物上传 COS 与 Windows 打包
- 前端优化左侧栏拖拽与展开收起，支持输入框内联 mention 高亮、任务进度侧栏、文件预览抽屉
- Skill 系统支持创建、编辑、删除操作，新增 Word 文档生成 Skill
- 集成交互式审批生命周期，支持 DOCX 文档预览
- 修复 Markdown 排版、数学公式渲染、ModelRouter 协议转换等多项问题
