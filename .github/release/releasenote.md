# OpenDDNS Release Note

## 新功能
- 🌟 **IPv6 DDNS 支持**：新增 A/AAAA 记录类型配置，支持 IPv4 和 IPv6 双栈 DDNS
- 🔍 **智能记录类型检测**：`record_type: "auto"` 模式下自动根据 IP 地址类型选择 A 或 AAAA 记录
- 🚀 **强制网络类型**：指定 A 记录时强制通过 IPv4 网络访问所有 API，指定 AAAA 记录时强制通过 IPv6 网络访问
- 📡 **新增 text 类型 IP 源**：支持直接返回 IP 地址的 API（如 `api64.ipify.org`）
- ⚡ **阿里云ESA支持**：新增阿里云边缘安全加速（ESA）DNS服务支持

## Provider 更新
- **Aliyun ESA**：全新支持阿里云ESA DNS服务
  - 支持自动创建、更新DNS记录
  - **自动站点ID查询**：配置中留空 `site_id` 时自动根据域名查询对应的ESA站点ID
  - 智能检测记录状态，避免不必要的API调用
  - 完整的错误处理和日志记录
- **网络强制**：A记录强制IPv4，AAAA记录强制IPv6访问API

## 配置更新
- 新增 `aliyun_esa` provider 类型
- 新增 `aliyun_esa` 配置项：`access_key_id`、`access_key_secret`、`site_id`
- 新增 `record_type` 配置项，可选值：`A`（IPv4）、`AAAA`（IPv6）、`auto`（自动检测）
- 默认配置文件中添加 AliyunESA 和 IPv6 IP 源示例

## 技术改进
- 新增 `FetchIPWithNetwork()` 函数，支持强制指定网络类型
- 自定义 HTTP Transport 和 Dialer，实现真正的网络层强制
- A 记录模式：强制 IPv4 网络（tcp4），禁用 IPv6 fallback
- AAAA 记录模式：强制 IPv6 网络（tcp6）

## 配置更新
- 新增 `record_type` 配置项，可选值：`A`（IPv4）、`AAAA`（IPv6）、`auto`（自动检测）
- 默认配置文件中添加 IPv6 IP 源示例
- 启动信息中显示当前记录类型配置

## 兼容性
- 向后兼容：现有配置文件无需修改，默认使用 `auto` 模式
- Provider 接口升级：Aliyun 和 Cloudflare provider 均已支持动态记录类型

## 其他改进
- 完善 README 文档，添加 IPv6 配置说明
- 优化日志输出，显示当前使用的记录类型
