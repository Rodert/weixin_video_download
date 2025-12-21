# Windows 图标打包指南

本文档说明如何将图标打包进 Windows 可执行文件（`.exe`）。

---

## 目录

- [方案说明](#方案说明)
- [快速开始](#快速开始)
- [详细步骤](#详细步骤)
- [验证图标](#验证图标)
- [常见问题](#常见问题)
- [自动化脚本](#自动化脚本)

---

## 方案说明

本项目使用 **rsrc** 工具将图标打包进 Windows 可执行文件。

**rsrc** 是一个轻量级的 Go 工具，专门用于将 Windows 资源（如图标）嵌入到 Go 程序中。

---

## 快速开始

### 一键操作

```powershell
# 1. 安装 rsrc 工具（首次使用）
go install github.com/akavel/rsrc@latest

# 2. 生成资源文件
rsrc -ico winres/icon.ico -o rsrc_windows_amd64.syso

# 3. 构建可执行文件
go build -ldflags="-s -w" -o wx_video_download.exe
```

---

## 详细步骤

### 步骤 1：安装 rsrc 工具

#### 方法一：使用 go install（推荐）

```powershell
go install github.com/akavel/rsrc@latest
```

#### 方法二：检查是否已安装

```powershell
# 检查 rsrc 是否可用
rsrc -h
```

如果提示找不到命令，需要将 Go 的 bin 目录添加到 PATH：

```powershell
# 查看 Go 的 bin 目录
go env GOPATH

# 临时添加到 PATH（当前会话有效）
$env:PATH += ";$env:USERPROFILE\go\bin"
```

#### 方法三：使用完整路径

如果 `rsrc` 不在 PATH 中，可以使用完整路径：

```powershell
& "$env:USERPROFILE\go\bin\rsrc.exe" -ico winres/icon.ico -o rsrc_windows_amd64.syso
```

### 步骤 2：准备图标文件

确保图标文件存在：

- **文件位置**：`winres/icon.ico`
- **格式要求**：必须是 `.ico` 格式（不是 `.png`）
- **文件检查**：
  ```powershell
  # 检查文件是否存在
  Test-Path winres/icon.ico
  ```

**注意**：如果只有 PNG 格式的图标，需要先转换为 ICO 格式。可以使用在线工具或图像编辑软件转换。

### 步骤 3：生成资源文件

在项目根目录执行：

```powershell
rsrc -ico winres/icon.ico -o rsrc_windows_amd64.syso
```

**参数说明**：
- `-ico`：指定图标文件路径
- `-o`：指定输出的 `.syso` 文件名

**输出**：
- 成功：生成 `rsrc_windows_amd64.syso` 文件
- 失败：检查错误信息

### 步骤 4：构建可执行文件

生成 `.syso` 文件后，正常构建即可：

```powershell
# 生产模式（不打印日志）
go build -ldflags="-s -w" -o wx_video_download.exe

# 调试模式（打印日志）
go build -ldflags="-s -w -X main.EnableLogs=true" -o wx_video_download.exe
```

Go 构建时会自动包含 `.syso` 文件中的资源。

### 步骤 5：清理（可选）

构建完成后，可以删除 `.syso` 文件（下次构建前需要重新生成）：

```powershell
Remove-Item rsrc_windows_amd64.syso
```

**注意**：`.syso` 文件已在 `.gitignore` 中，不会被提交到 Git 仓库。

---

## 验证图标

### 方法一：文件资源管理器

1. 打开文件资源管理器
2. 找到生成的 `wx_video_download.exe` 文件
3. 查看文件图标是否显示为自定义图标

### 方法二：文件属性

1. 右键点击 `wx_video_download.exe`
2. 选择"属性"
3. 查看"详细信息"标签页，确认图标已嵌入

### 方法三：运行程序

运行程序后，在任务栏和窗口标题栏查看图标是否正确显示。

---

## 常见问题

### Q1: `rsrc: 无法将"rsrc"项识别为 cmdlet...`

**原因**：`rsrc` 工具未安装或不在 PATH 中。

**解决方案**：

1. **安装工具**：
   ```powershell
   go install github.com/akavel/rsrc@latest
   ```

2. **添加到 PATH**（临时）：
   ```powershell
   $env:PATH += ";$env:USERPROFILE\go\bin"
   ```

3. **使用完整路径**：
   ```powershell
   & "$env:USERPROFILE\go\bin\rsrc.exe" -ico winres/icon.ico -o rsrc_windows_amd64.syso
   ```

### Q2: `open winres/icon.ico: The system cannot find the file specified.`

**原因**：图标文件不存在或路径错误。

**解决方案**：

1. 检查文件是否存在：
   ```powershell
   Test-Path winres/icon.ico
   ```

2. 确认当前目录是项目根目录：
   ```powershell
   pwd
   ```

3. 检查文件格式：必须是 `.ico` 格式，不是 `.png`

### Q3: 构建后图标没有显示

**可能原因**：

1. **没有生成 `.syso` 文件**：确保执行了 `rsrc` 命令
2. **`.syso` 文件位置不对**：必须在项目根目录
3. **图标文件损坏**：检查 `icon.ico` 文件是否有效
4. **Windows 缓存**：尝试刷新文件资源管理器（F5）或重启

**解决方案**：

1. 确认 `.syso` 文件存在：
   ```powershell
   Test-Path rsrc_windows_amd64.syso
   ```

2. 重新生成资源文件：
   ```powershell
   rsrc -ico winres/icon.ico -o rsrc_windows_amd64.syso
   ```

3. 清理并重新构建：
   ```powershell
   Remove-Item wx_video_download.exe -ErrorAction SilentlyContinue
   go build -ldflags="-s -w" -o wx_video_download.exe
   ```

### Q4: 如何将 PNG 转换为 ICO？

**在线工具**：
- https://convertio.co/zh/png-ico/
- https://www.icoconverter.com/

**使用 ImageMagick**（如果已安装）：
```powershell
magick convert icon.png -define icon:auto-resize=256,128,64,48,32,16 icon.ico
```

### Q5: 可以同时支持多个图标尺寸吗？

`rsrc` 工具支持多尺寸图标。需要创建一个包含多个尺寸的 ICO 文件：

1. 准备多个尺寸的 PNG 图标（如 16x16, 32x32, 48x48, 256x256）
2. 使用工具将它们合并成一个 ICO 文件
3. 使用合并后的 ICO 文件生成资源

---

## 自动化脚本

### PowerShell 脚本（build.ps1）

创建 `build.ps1` 文件：

```powershell
# 构建脚本 - 自动打包图标
param(
    [switch]$Debug
)

Write-Host "开始构建..." -ForegroundColor Green

# 1. 检查 rsrc 工具
Write-Host "检查 rsrc 工具..." -ForegroundColor Yellow
$rsrcPath = "$env:USERPROFILE\go\bin\rsrc.exe"
if (-not (Test-Path $rsrcPath)) {
    Write-Host "rsrc 工具未安装，正在安装..." -ForegroundColor Yellow
    go install github.com/akavel/rsrc@latest
    if ($LASTEXITCODE -ne 0) {
        Write-Host "安装 rsrc 失败！" -ForegroundColor Red
        exit 1
    }
}

# 2. 检查图标文件
Write-Host "检查图标文件..." -ForegroundColor Yellow
if (-not (Test-Path "winres\icon.ico")) {
    Write-Host "错误：找不到 winres\icon.ico 文件！" -ForegroundColor Red
    exit 1
}

# 3. 生成资源文件
Write-Host "生成资源文件..." -ForegroundColor Yellow
& $rsrcPath -ico winres/icon.ico -o rsrc_windows_amd64.syso
if ($LASTEXITCODE -ne 0) {
    Write-Host "生成资源文件失败！" -ForegroundColor Red
    exit 1
}

# 4. 构建可执行文件
Write-Host "构建可执行文件..." -ForegroundColor Yellow
if ($Debug) {
    go build -ldflags="-s -w -X main.EnableLogs=true" -o wx_video_download.exe
} else {
    go build -ldflags="-s -w" -o wx_video_download.exe
}

if ($LASTEXITCODE -ne 0) {
    Write-Host "构建失败！" -ForegroundColor Red
    exit 1
}

Write-Host "构建完成！" -ForegroundColor Green
```

**使用方法**：

```powershell
# 生产模式
.\build.ps1

# 调试模式
.\build.ps1 -Debug
```

### 批处理脚本（build.bat）

创建 `build.bat` 文件：

```batch
@echo off
echo 开始构建...

REM 1. 生成资源文件
echo 生成资源文件...
rsrc -ico winres/icon.ico -o rsrc_windows_amd64.syso
if errorlevel 1 (
    echo 生成资源文件失败！
    pause
    exit /b 1
)

REM 2. 构建可执行文件
echo 构建可执行文件...
go build -ldflags="-s -w" -o wx_video_download.exe
if errorlevel 1 (
    echo 构建失败！
    pause
    exit /b 1
)

echo 构建完成！
pause
```

**使用方法**：双击运行 `build.bat` 或在命令行执行。

---

## 完整构建流程

```powershell
# ============================================
# Windows 图标打包完整流程
# ============================================

# 1. 安装 rsrc 工具（首次使用）
go install github.com/akavel/rsrc@latest

# 2. 将 Go bin 目录添加到 PATH（如果未添加）
$env:PATH += ";$env:USERPROFILE\go\bin"

# 3. 生成资源文件
rsrc -ico winres/icon.ico -o rsrc_windows_amd64.syso

# 4. 构建可执行文件
go build -ldflags="-s -w" -o wx_video_download.exe

# 5. 验证图标（查看文件资源管理器中的图标）

# 6. 清理（可选）
# Remove-Item rsrc_windows_amd64.syso
```

---

## 文件说明

### 相关文件

- `winres/icon.ico` - 图标源文件（ICO 格式）
- `rsrc_windows_amd64.syso` - 生成的资源文件（构建时自动包含）
- `wx_video_download.exe` - 最终生成的可执行文件

### 文件位置

```
项目根目录/
├── winres/
│   └── icon.ico          # 图标源文件
├── rsrc_windows_amd64.syso  # 资源文件（构建时生成）
└── wx_video_download.exe    # 最终可执行文件
```

---

## 注意事项

1. **图标格式**：必须使用 `.ico` 格式，不支持 `.png`
2. **文件位置**：`.syso` 文件必须在项目根目录
3. **每次构建**：如果修改了图标，需要重新生成 `.syso` 文件
4. **Git 忽略**：`.syso` 文件已在 `.gitignore` 中，不会提交到仓库
5. **跨平台**：`rsrc` 只用于 Windows，其他平台不需要此步骤

---

## 相关文档

- [README.md](README.md) - 项目说明
- [DELIVERY.md](DELIVERY.md) - 交付文档
- [DEVELOPER.md](DEVELOPER.md) - 开发文档

---

**最后更新**：2025-12-06

