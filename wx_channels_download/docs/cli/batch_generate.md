# 批量生成密钥

批量生成积分密钥，适用于需要一次性生成多个密钥的场景。

## 命令格式

```bash
go run . batch-generate --points <积分数量> --days <有效天数> --count <生成数量>
```

## 参数说明

- `--points`：单组积分数量（必填，必须大于 0，默认：10）
- `--days`：单组有效天数（必填，必须大于 0，从首次使用时开始计算，默认：7）
- `--count`：生成组数（必填，必须大于 0，默认：100）

## 版本兼容性

**重要**：所有生成的密钥都会自动包含当前版本号（从 `version.txt` 读取），确保：

- ✅ 生成的密钥只能在相同版本的应用程序中使用
- ✅ 不同版本的密钥互不兼容，防止跨版本使用
- ✅ 升级版本后，旧版本密钥将无法使用，需要重新生成

**版本号来源**：
- 优先使用嵌入在可执行文件中的版本号（打包时从 `version.txt` 嵌入）
- 如果未找到嵌入版本，则从文件系统读取 `version.txt`
- 如果都读取失败，默认使用 `v1`

## 使用示例

### 基本用法

```bash
# 生成 100 个密钥，每个密钥包含 10 积分，有效期 7 天
go run . batch-generate --points 10 --days 7 --count 100
```

### 保存到文件

```bash
# 生成密钥并保存到文件（只保存密钥，不包含提示信息）
go run . batch-generate --points 10 --days 7 --count 100 > keys.txt
```

### 同时查看信息和保存密钥

```bash
# 生成密钥，同时查看信息并保存到文件
go run . batch-generate --points 10 --days 7 --count 100 2>&1 | tee keys.txt
```

### 生成不同规格的密钥

```bash
# 生成 50 个密钥，每个包含 100 积分，有效期 30 天
go run . batch-generate --points 100 --days 30 --count 50 > keys_100_30.txt

# 生成 200 个密钥，每个包含 5 积分，有效期 3 天
go run . batch-generate --points 5 --days 3 --count 200 > keys_5_3.txt
```

## 输出格式

### 标准输出（stdout）

每行一个加密的密钥字符串，可以直接保存到文件：

```
n2wJFTONgbBzHxNo9DMNI6KEX62E/gWdZY+OQlY5HWmJC0BM2dEVt+vCJt2M6/V6a/gLRn4N8m/8LvZhA2h5IGj4PCbOJV2z94R8eOPB+SbmIGujOi7B
n2wJFTONgbBzHxNo9DMNI6KEX62E/gWdZY+OQlY5HWmJC0BM2dEVt+vCJt2M6/V6a/gLRn4N8m/8LvZhA2h5IGj4PCbOJV2z94R8eOPB+SbmIGujOi7C
...
```

### 标准错误输出（stderr）

包含生成信息和提示：

```
批量生成积分配置
版本号: v3
单组积分: 10
单组有效天数: 7 天（从首次使用时开始计算）
生成数量: 100
----------------------------------------
开始生成（以下为密钥列表，每行一个）:

----------------------------------------
批量生成完成！共生成 100 个密钥
所有密钥版本: v3

使用示例：
  # 生成密钥并保存到文件
  go run . batch-generate --points 10 --days 7 --count 100 > keys.txt
  
  # 生成密钥并保存到文件（同时查看信息）
  go run . batch-generate --points 10 --days 7 --count 100 2>&1 | tee keys.txt
```

## 密钥使用

生成的密钥可以：

1. **直接提供给用户**：将密钥字符串发送给用户，用户创建 `credit.txt` 文件：
   ```txt
   encrypted=密钥字符串
   ```

2. **批量分发**：将多个密钥保存到文件，按需分发给不同用户

3. **版本管理**：所有密钥都包含版本号，确保只能在对应版本使用

## 版本升级注意事项

当应用程序升级到新版本时（例如从 v3 升级到 v4）：

1. **旧版本密钥失效**：v3 版本的密钥无法在 v4 版本中使用
2. **需要重新生成**：必须使用新版本重新生成密钥
3. **版本号自动包含**：新生成的密钥会自动包含新版本号（v4）

**示例**：
```bash
# v3 版本时生成的密钥
go run . batch-generate --points 10 --days 7 --count 100 > keys_v3.txt

# 升级到 v4 后，需要重新生成
go run . batch-generate --points 10 --days 7 --count 100 > keys_v4.txt
```

## 与单密钥生成的区别

| 特性 | `generate-credit` | `batch-generate` |
|------|-------------------|-----------------|
| 生成数量 | 1 个 | 多个（可指定） |
| 输出格式 | 详细说明 + 密钥 | 每行一个密钥 |
| 适用场景 | 单个用户 | 批量分发 |
| 版本兼容 | ✅ 自动包含版本号 | ✅ 自动包含版本号 |
| 重定向 | 需要手动提取 | 可直接重定向到文件 |

## 常见问题

### Q: 如何确保生成的密钥版本正确？

A: 密钥生成时会自动读取当前版本号（从 `version.txt` 或嵌入的版本号），无需手动指定。确保 `version.txt` 文件存在且内容正确即可。

### Q: 批量生成的密钥可以混用吗？

A: 可以。只要版本号相同，所有密钥都可以在同一个版本的应用程序中使用。每个密钥独立使用，互不影响。

### Q: 如何验证密钥的版本号？

A: 使用 `generate-credit` 命令生成单个密钥时，会显示版本号信息。批量生成的密钥版本号与当前应用程序版本一致。

### Q: 升级版本后，旧密钥还能用吗？

A: 不能。版本升级后，旧版本的密钥无法在新版本中使用。需要重新生成新版本的密钥。

### Q: 如何查看当前版本号？

A: 程序启动时会显示版本号，或查看 `version.txt` 文件内容。

---

**相关命令**：
- [`generate-credit`](./generate_credit.md) - 生成单个密钥（带详细说明）
- [`version`](./version.md) - 查看版本信息

