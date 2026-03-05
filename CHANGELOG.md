# Changelog

所有项目的显著变更都将记录在此文件中。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
并且该项目遵循 [语义化版本](https://semver.org/lang/zh-CN/)。

## [v0.0.9] - 2024-07-??

### 新增

- 优化日期过期检测逻辑

### 修复

- 修复未配置证书保存路径时默认存储路径为空的问题
- 修复 CDN 证书信息过期时间检查不一致问题

### 依赖更新

- 升级 `mikepenz/release-changelog-builder-action` 从 4 到 5
- 升级 `docker/build-push-action` 从 5 到 6

## [v0.0.8] - 2024-06-??

### 新增

- 添加调试模式环境变量开关 `DEBUG`
- 完善 Debug 日志输出

## [v0.0.7] - 2024-06-??

### 新增

- 启动时立即执行证书检测
- 添加 webhook 通知兼容性支持
- 添加 Dependabot 配置

## [v0.0.6] - 2024-06-??

### 修复

- 修复 lego 注册 ACME 次数限制问题

## [v0.0.5] - 2024-05-??

### 新增

- 证书续期时间支持可配置
- 添加 Docker Hub 镜像支持
- ACME 相关配置支持环境变量

### 修复

- 添加 CI 读取 PR 权限

## [v0.0.4] - 2024-05-??

### 新增

- 优化开发体验，添加开发证书申请环境变量支持
- 优化 Changelog 生成

### 文档

- 更新说明文档

## [v0.0.3] - 2024-05-??

### 新增

- 更新 CI 配置
- 添加 Changelog 构建配置

## [v0.0.2] - 2024-04-??

### 新增

- 添加 Dockerfile 支持容器化部署
- 添加 CI/CD 自动发布配置
- 添加 Webhook 通知功能（企业微信、钉钉等）
- 添加阿里云 DCDN 支持
- 实现证书申请续签和 OSS 证书更新功能
- 添加 OSS 自定义域名证书关联测试
- 添加文件操作工具函数
- 切换日志库为 charmbracelet/log
- 添加 UML 架构图

### 文档

- 完善项目说明文档
- 添加 LICENSE

## [v0.0.1] - 2024-04-??

### 新增

- 项目初始化
- 实现基础的阿里云 OSS 证书管理功能
- 集成 Let's Encrypt ACME 客户端 (lego)
- 实现配置加载功能
- 添加阿里云 OpenAPI 集成

[未发布]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.9...HEAD
[v0.0.9]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.8...v0.0.9
[v0.0.8]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.7...v0.0.8
[v0.0.7]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.6...v0.0.7
[v0.0.6]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.5...v0.0.6
[v0.0.5]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.4...v0.0.5
[v0.0.4]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/nekoimi/oss-auto-cert/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/nekoimi/oss-auto-cert/releases/tag/v0.0.1
