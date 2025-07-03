# OpenDDNS

现代化多云 DDNS 动态域名解析工具，支持 Cloudflare、阿里云，多 IP 提供源，日志等级与文件输出，自动检查更新。

## 特性
- 目前支持 Cloudflare、阿里云（Aliyun），全部采用官方 SDK
- 多 IP 源自动判定，支持 JSON、trace 等格式
- 日志等级支持 debug/info/warn/error
- 启动参数支持 `-c/--config` 指定配置文件，`--no-check-update` 跳过更新检查
- 首次启动自动生成默认 `config.yml`

---

## 快速开始

1. **下载**
   - 前往 [Releases](https://github.com/GloryRedstoneUnion/OpenDDNS/releases) 下载对应平台的二进制文件。
   - 或自行编译：
     ```sh
     git clone https://github.com/GloryRedstoneUnion/OpenDDNS.git
     cd OpenDDNS
     go build -o ddns.exe
     ```
   
2. **配置**
   - 首次运行会自动生成 `config.yml`并自动退出，请手动编辑配置文件（见下方配置示例）。
   
3. **运行**
   - Windows:
     ```sh
     openddns-xxx-xxx.exe
     ```
     
   - Linux/macOS:
     ```sh
     ./openddns-xxx-xxx
     ```
     
   - 可选参数：
     - `-c`/`--config` 指定配置文件
     
       ```
       ./openddns-xxx-xxx -c myconfig.yml
       ```
     
     - `--no-check-update` 跳过启动时版本检查

---

## 配置示例

```yaml
provider: "cloudflare"
domain: "example.com"
subdomain: "www"
log_level: "info"
log_file: ""
ip_sources:
  - name: "bilibili"
    url: "https://api.live.bilibili.com/xlive/web-room/v1/index/getIpInfo"
    type: "json"
    json_path: "data.addr"
  - name: "cloudflare"
    url: "https://www.cloudflare-cn.com/cdn-cgi/trace"
    type: "trace"
update_interval_minutes: 5
cloudflare:
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"
  zone_id: ""
aliyun:
  access_key_id: "YOUR_ALIYUN_ACCESS_KEY_ID"
  access_key_secret: "YOUR_ALIYUN_ACCESS_KEY_SECRET"
  endpoint: "alidns.aliyuncs.com"
```

---

## 配置项目录

- [provider](#provider)
- [domain](#domain)
- [subdomain](#subdomain)
- [log_level](#log_level)
- [log_file](#log_file)
- [ip_sources](#ip_sources)
- [update_interval_minutes](#update_interval_minutes)
- [cloudflare](#cloudflare)
- [aliyun](#aliyun)

---

### <a id="provider"></a>provider
- **类型**：string
- **说明**：选择 DNS 服务商。可选值：`cloudflare`、`aliyun`
- **示例**：`provider: "cloudflare"`

### <a id="domain"></a>domain
- **类型**：string
- **说明**：主域名，不含子域名部分。
- **示例**：`domain: "example.com"`

### <a id="subdomain"></a>subdomain
- **类型**：string

- **说明**：子域名部分。添加前会自动检查是否存在此子域名，若不存在将自动创建。

> [!WARNING]
> 若此子域名存在多个解析记录，则会将它们全部覆盖为当前公网IP。

- **示例**：`subdomain: "www"`

### <a id="log_level"></a>log_level
- **类型**：string
- **说明**：日志等级。可选：`debug`、`info`、`warn`、`error`
- **示例**：`log_level: "info"`

### <a id="log_file"></a>log_file
- **类型**：string
- **说明**：日志文件路径，留空则仅输出到控制台。
- **示例**：`log_file: ""`

### <a id="ip_sources"></a>ip_sources
- **类型**：数组

- **说明**：公网 IP 获取源列表，可填多个，建议数量为奇数。当有多个IP源存在时，将启用投票机制，多数者胜。
	
> [!NOTE]
> OpenDDNS目前仅会向设定的URL发送GET请求以获取响应，更多配置项将在后续版本添加。


- **结构**：

  - `name`：源名称

  - `url`：请求地址

  - `type`：`json` 或 `trace`

> [!NOTE]  
> 目前对trace的兼容性较差，建议使用提供json响应的API，对于trace的兼容性配置将在后续版本更新。
>
> **trace模式说明：**
>
> 1. 先将 HTTP 响应内容按行（\n）分割。
> 2. 遍历每一行，查找以 `ip=` 开头的行（如 `ip=1.2.3.4`）。
> 3. 找到后，去掉前缀 `ip=`，直接返回后面的 IP 字符串。
> 4. 如果遍历完都没找到，则返回错误。

  - `json_path`：仅 type 为 json 时必填，指定 IP 字段路径，OpenDDNS将从API响应中提取对应路径的值

- **示例**：

```yaml
ip_sources:
  - name: "bilibili"
    url: "https://api.live.bilibili.com/xlive/web-room/v1/index/getIpInfo"
    type: "json"
    json_path: "data.addr"
  - name: "cloudflare"
    url: "https://www.cloudflare-cn.com/cdn-cgi/trace"
    type: "trace"
```

### <a id="update_interval_minutes"></a>update_interval_minutes

- **类型**：int
- **说明**：检测并同步 IP 的时间间隔（分钟）。
- **示例**：`update_interval_minutes: 5`

### <a id="cloudflare"></a>cloudflare
- **类型**：对象
- **说明**：Cloudflare 账户配置。
  - `api_token`：Cloudflare API Token
  - `zone_id`：可选，留空自动获取
- **示例**：
```yaml
cloudflare:
  api_token: "YOUR_CLOUDFLARE_API_TOKEN"
  zone_id: ""
```

> [!IMPORTANT]
> 所使用的Cloudflare账户必须对相应主域有**编辑DNS权限**：

### <a id="aliyun"></a>aliyun

- **类型**：对象
- **说明**：阿里云相关配置。
  - `access_key_id`：用来进行DNS操作的阿里云账户AccessKey ID
  - `access_key_secret`：用来进行DNS操作的阿里云账户 AccessKey Secret
  - `endpoint`：可选，默认为 `alidns.aliyuncs.com`
- **示例**：
```yaml
aliyun:
  access_key_id: "YOUR_ALIYUN_ACCESS_KEY_ID"
  access_key_secret: "YOUR_ALIYUN_ACCESS_KEY_SECRET"
  endpoint: "alidns.aliyuncs.com"
```
  > [!IMPORTANT]
  > 所使用的阿里云账户必须具有以下权限：
  >
  > ```
  > alidns:DescribeDomainRecords
  > alidns:UpdateDomainRecord
  > alidns:AddDomainRecord
  > alidns:DescribeSubDomainRecords
  > ```
  >

---

## 其它说明

- **首次启动**：无 config.yml 会自动生成模板并退出

- **命令行参数**：
  - `-c`/`--config` 指定配置文件
  - `--no-check-update` 跳过启动时版本检查
  
- **关于权限**：

  由于大多数使用AccessKey/AccessSecret方法鉴权的DNS提供商所提供的Key和Secret权限等同于账号权限，且Key和Secret是明文保存在配置文件中的，所以建议使用子账户并为其分配最小化权限。
> [!CAUTION]
> OpenDDNS开发者不对任何访问密钥或身份凭证泄露事件及其造成的后果负任何责任。OpenDDNS不会试图收集你的访问密钥或身份凭证。

- **常见问题**：
  
  - 提交PR时 config.yml 泄露：已默认加入 .gitignore，请勿上传敏感配置
  - 支持平台：Windows/Linux/macOS/FreeBSD/ARM 等主流架构，**对任何32位操作系统均不提供支持**，任何因非64位操作系统/计算平台产生的issue将被关闭。

## 贡献
欢迎 PR、issue 反馈与建议！

