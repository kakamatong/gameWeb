# gameWeb 项目文档

本目录包含 gameWeb 项目的所有技术文档和说明文件。

## 📋 文档分类

### API 接口文档
- [`admin_update_api.md`](./admin_update_api.md) - 管理员信息更新接口文档
- [`get_admin_info_api.md`](./get_admin_info_api.md) - 获取管理员信息接口文档
- [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md) - 完整的API接口文档
- [`API_SEPARATION.md`](./API_SEPARATION.md) - API分离设计文档

### 数据库文档
- [`DATABASE_FIX.md`](./DATABASE_FIX.md) - 数据库修复记录
- [`DATABASE_STRUCTURE_FIX.md`](./DATABASE_STRUCTURE_FIX.md) - 数据库结构修复文档

### 安全与部署
- [`SECURITY_CLEANUP_COMPLETE.md`](./SECURITY_CLEANUP_COMPLETE.md) - 安全清理完成记录
- [`SECURITY_DEPLOYMENT.md`](./SECURITY_DEPLOYMENT.md) - 安全部署指南

### 项目管理
- [`FEATURE_CHECKLIST.md`](./FEATURE_CHECKLIST.md) - 功能开发清单

## 📖 文档使用指南

### 开发者快速入门
1. 首先阅读项目根目录的 [`README.md`](../README.md)
2. 查看 [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md) 了解完整的API设计
3. 参考具体接口文档进行开发和测试

### API 开发流程
1. 查看 [`API_SEPARATION.md`](./API_SEPARATION.md) 了解API设计原则
2. 参考现有接口文档的格式编写新的API文档
3. 使用 [`../test/`](../test/) 目录下的测试脚本进行验证

### 数据库操作
1. 查看 [`../sql/`](../sql/) 目录了解数据库表结构
2. 参考数据库相关文档进行操作
3. 遵循数据库修复文档中的最佳实践

### 安全考虑
1. 阅读安全相关文档了解安全要求
2. 按照安全部署指南进行项目部署
3. 定期检查安全清理文档中的安全检查项

## 🔄 文档维护

### 更新原则
- 每个新功能都必须有对应的文档
- API变更必须同步更新文档
- 重要的bug修复和改进要记录在相应文档中

### 文档格式
- 使用 Markdown 格式
- 遵循统一的文档结构和样式
- 包含必要的代码示例和测试用例

### 文档审查
- 新增文档需要代码审查
- 重要文档变更需要团队确认
- 定期检查文档的准确性和完整性

## 📞 联系方式

如果您对文档有任何疑问或建议，请通过以下方式联系：
- 提交 Issue 到项目仓库
- 发送邮件到项目维护者
- 在项目讨论区发起讨论

---

*最后更新时间: 2024-08-28*