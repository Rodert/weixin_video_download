# 交付文档

本文档面向最终用户，说明如何打包程序、生成密钥以及交付内容。

---

## 目录

- [打包程序](#打包程序)
- [生成积分密钥](#生成积分密钥)
- [交付内容](#交付内容)
- [使用说明](#使用说明)

---

## 打包程序

### Windows 打包

#### 生产模式（推荐）

```powershell
# 不打印日志，适合正式发布
go build -ldflags="-s -w" -o wx_video_download.exe
```

#### 调试模式

```powershell
# 打印详细日志，便于调试
go build -ldflags="-s -w -X main.EnableLogs=true" -o wx_video_download.exe
```

**参数说明**：
- `-s`：去除符号表
- `-w`：去除调试信息
- `-X main.EnableLogs=true`：启用日志打印（仅调试模式需要）

### macOS 打包

```bash
CGO_ENABLED=1 GOOS=darwin SDKROOT=$(xcrun --sdk macosx --show-sdk-path) go build -trimpath -ldflags="-s -w" -o wx_video_download
```

### Linux 打包

```bash
go build -ldflags="-s -w" -o wx_video_download
```

---

## 生成积分密钥

### 1. 生成加密密钥

使用 `generate-credit` 命令生成积分密钥：

```bash
go run main.go generate-credit --points 1000 --days 7
```

**参数说明**：
- `--points`：积分数量（必填，必须大于 0，默认：1000）
- `--days`：有效天数（必填，必须大于 0，从首次使用时开始计算）

**有效期说明**：
- 有效期从用户首次使用积分下载时开始计算
- 例如：生成 7 天有效期的密钥，用户在第 5 天首次使用，则有效期到第 12 天
- 确保用户获得完整的有效期，不会因为生成时间而损失

**示例**：

```bash
# 生成 1000 积分，有效期 7 天（从首次使用时开始计算）
go run main.go generate-credit --points 1000 --days 7

# 生成 500 积分，有效期 30 天
go run main.go generate-credit --points 500 --days 30
```

**输出示例**：

```
==============================================================
积分配置生成成功！
==============================================================

请创建 credit.yaml 文件（与可执行文件同目录），内容如下：

encrypted: n2wJFTONgbBzHxNo9DMNI6KEX62E/gWdZY+OQlY5HWmJC0BM2dEVt+vCJt2M6/V6a/gLRn4N8m/8LvZhA2h5IGj4PCbOJV2z94R8eOPB+SbmIGujOi7B

或者直接创建文件：
echo encrypted: n2wJFTONgbBzHxNo9DMNI6KEX62E/gWdZY+OQlY5HWmJC0BM2dEVt+vCJt2M6/V6a/gLRn4N8m/8LvZhA2h5IGj4PCbOJV2z94R8eOPB+SbmIGujOi7B > credit.yaml

积分信息：
  积分数量: 1000
  有效天数: 7 天（从首次使用时开始计算）
  激活状态: 未激活（首次使用时自动激活）

==============================================================
```

**积分消耗规则**：
- 下载视频：消耗 5 积分
- 下载封面：消耗 1 积分

**防重复使用机制**：
- 每次下载后，原始密钥会被记录到 `.use` 文件（与可执行文件同目录）
- 每次下载前，会检查密钥是否已被使用
- 即使积分还有剩余，一旦使用过就无法再次使用相同的密钥
- 防止用户通过替换密钥文件来重复使用已用完的密钥

### 2. 创建密钥文件

**方法一：使用模板文件（推荐）**

1. 复制模板文件到可执行文件目录：
   ```bash
   # Windows (PowerShell)
   Copy-Item config\credit.template.yaml .\credit.yaml
   
   # Linux/macOS
   cp config/credit.template.yaml ./credit.yaml
   ```

2. 编辑 `credit.yaml`，将生成的 `encrypted` 值填入：
   ```yaml
   encrypted: "生成的加密字符串"
   ```

**方法二：手动创建**

在与可执行文件同目录下创建 `credit.yaml` 文件，内容为：

```yaml
encrypted: "生成的加密字符串"
```

### 3. 验证密钥

启动程序后，如果积分功能正常启用，会在下载列表中显示积分信息。

---

## 交付内容

### 必需文件

1. **可执行文件**
   - Windows: `wx_video_download.exe`
   - macOS: `wx_video_download`
   - Linux: `wx_video_download`

2. **配置文件模板**
   - `config/config.template.yaml` - 主配置模板
   - `config/credit.template.yaml` - 积分密钥模板

3. **文档**
   - `README.md` - 用户使用说明
   - `DELIVERY.md` - 本交付文档（可选，面向开发者）

### 可选文件

1. **开发文档**（面向开发者）
   - `DEVELOPER.md` - 详细的技术文档和开发指南

2. **其他资源**
   - `global.js` - 全局用户脚本（如需要）
   - 图标文件（如需要）

### 交付清单

```
交付包/
├── wx_video_download.exe          # 可执行文件（Windows）
├── config/
│   ├── config.template.yaml      # 主配置模板
│   └── credit.template.yaml       # 积分密钥模板
├── README.md                      # 用户使用说明
├── DELIVERY.md                    # 交付文档（本文件）
└── DEVELOPER.md                  # 开发文档（可选）
```

---

## 使用说明

### 首次使用

1. **解压交付包**到目标目录

2. **创建配置文件**（可选）
   - 如果需要自定义配置，复制 `config/config.template.yaml` 为 `config.yaml`
   - 编辑 `config.yaml` 进行配置

3. **配置积分密钥**（如需要）
   - 使用 `generate-credit` 命令生成密钥
   - 创建 `credit.yaml` 文件并填入生成的 `encrypted` 值
   - 或使用模板文件：复制 `config/credit.template.yaml` 为 `credit.yaml`，然后填入密钥

4. **运行程序**
   - Windows: 双击 `wx_video_download.exe` 或命令行运行
   - macOS/Linux: `./wx_video_download`

### 积分续期

当客户积分用完或到期后，需要续期：

1. **生成新密钥**
   ```bash
   go run main.go generate-credit --points 1000 --days 7
   ```

2. **提供给客户**
   - 将生成的 `encrypted` 值发送给客户
   - 或直接提供新的 `credit.yaml` 文件

3. **客户更新**
   - 客户将新的 `encrypted` 值更新到 `credit.yaml` 文件
   - 重启程序，新配置生效
   - 首次使用时自动激活，从使用时间开始计算有效期

**注意事项**：
- 每个密钥只能使用一次，使用后会被记录到 `.use` 文件
- 如果客户尝试重复使用已用完的密钥，系统会拒绝
- 续期时需要提供全新的密钥，不能使用之前已使用过的密钥

### 配置文件位置

- **主配置文件**：`config.yaml`（与可执行文件同目录）
- **积分密钥文件**：`credit.yaml`（与可执行文件同目录）
- **已使用密钥记录**：`.use`（与可执行文件同目录，自动生成，用于防止重复使用）
- **全局脚本**：`global.js`（与可执行文件同目录，可选）

---

## 注意事项

### 打包注意事项

1. **生产模式**：正式发布时使用生产模式打包（不启用日志），减少文件大小
2. **调试模式**：仅在调试时使用，会打印详细日志
3. **跨平台**：如需打包其他平台，使用 `GOOS` 和 `GOARCH` 环境变量

### 密钥生成注意事项

1. **日期格式**：支持多种日期格式，但建议统一使用一种格式
2. **有效期**：开始日期为当天 00:00:00，结束日期为当天 23:59:59
3. **积分数量**：每次下载消耗 5 积分，生成时需考虑使用频率
4. **安全性**：密钥使用 AES-256-GCM 加密，密钥硬编码在代码中

### 交付注意事项

1. **文件完整性**：确保所有必需文件都已包含
2. **文档完整性**：提供必要的使用说明文档
3. **测试验证**：交付前测试打包后的程序是否正常运行
4. **版本信息**：建议在文件名或文档中标注版本号和打包日期

---

## 常见问题

### Q: 打包后的程序无法运行？

A: 检查以下几点：
- 是否缺少必要的系统库（Windows 可能需要 Visual C++ 运行库）
- 是否以管理员权限运行（Windows 需要管理员权限设置系统代理）
- 检查控制台是否有错误信息

### Q: 如何查看程序版本？

A: 程序启动时会在控制台显示版本信息（格式：`版本: YYYY-MM-DD`）

### Q: 积分密钥生成失败？

A: 检查：
- 日期格式是否正确
- 开始日期是否早于结束日期
- 积分数量是否大于 0

### Q: 客户如何续期积分？

A: 
1. 使用 `generate-credit` 命令生成新的密钥
2. 将新的 `encrypted` 值提供给客户
3. 客户更新 `credit.yaml` 文件中的 `encrypted` 字段
4. 重启程序

### Q: 配置文件在哪里？

A: 
- 主配置文件：`config.yaml`（与可执行文件同目录）
- 积分密钥文件：`credit.yaml`（与可执行文件同目录）
- 如果不存在，程序会使用默认配置

---

## 技术支持

如有问题，请参考：
- `README.md` - 用户使用说明
- `DEVELOPER.md` - 开发文档（详细技术说明）

---

**文档版本**：1.0  
**最后更新**：2025-12-06

