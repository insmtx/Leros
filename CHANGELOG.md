# Changelog

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
