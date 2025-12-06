# 开发者文档

本文档面向开发者，详细介绍了微信视频号下载器的功能特性、架构设计和技术实现。

## 目录

- [功能特性](#功能特性)
- [架构设计](#架构设计)
- [核心技术点](#核心技术点)
- [实现细节](#实现细节)
- [开发指南](#开发指南)

---

## 功能特性

### 1. 视频下载

#### 1.1 下载原视频（多种质量）
- **功能描述**：支持下载视频号中的视频，可选择不同质量规格（如 xWT111、xWT98 等）或原始视频
- **实现方式**：
  - 通过 HTTP 代理拦截视频播放请求
  - 注入 JavaScript 捕获视频流数据（SourceBuffer）
  - 使用 FileSaver.js 在浏览器端直接下载
  - 支持加密视频的自动解密
- **相关文件**：
  - `inject/main.js` - 前端下载逻辑
  - `internal/interceptor/plugin.go` - 请求拦截和 JS 注入
  - `pkg/decrypt/decrypt.go` - 视频解密算法

#### 1.2 长视频下载
- **功能描述**：针对超过 30 分钟的长视频，提供命令行下载方式，支持多线程下载
- **实现方式**：
  - 前端提供"打印下载命令"功能
  - 使用 `download` 命令进行多线程下载
  - 下载完成后自动解密
- **相关文件**：
  - `cmd/download.go` - 命令行下载实现
  - `internal/download/download.go` - 下载和解密逻辑

### 2. 音频下载

#### 2.1 下载 MP3（前端转换）
- **功能描述**：将短视频转换为 MP3 格式下载
- **实现方式**：
  - 使用 Web Audio API 解码音频
  - 使用 recorder.min.js 转换为 WAV
  - 使用 lame.js 将 WAV 转换为 MP3
  - 完全在浏览器端完成，无需服务器
- **相关文件**：
  - `inject/main.js` - `__wx_channels_download4` 函数
  - `inject/utils.js` - `wavBlobToMP3` 函数
  - `inject/lib/recorder.min.js` - 音频录制库

#### 2.2 下载 MP3（服务器转换）
- **功能描述**：通过本地服务器将长视频转换为 MP3
- **实现方式**：
  - 前端请求本地下载服务器
  - 服务器使用 FFmpeg 进行视频转音频
  - 支持加密视频的解密和转换
- **前置条件**：需要安装 FFmpeg 并配置到 PATH
- **相关文件**：
  - `internal/download/download.go` - `convertWithDecrypt` 和 `convertOnly` 函数
  - `internal/download/server.go` - 下载服务器实现

### 3. 封面图下载

- **功能描述**：下载视频的封面图片
- **实现方式**：
  - 从视频元数据中提取封面 URL
  - 使用 fetch 下载图片
  - 使用 FileSaver.js 保存到本地
- **相关文件**：
  - `inject/main.js` - `__wx_channels_handle_download_cover` 函数

### 4. 下载列表管理

- **功能描述**：提供悬浮的下载列表，记录和管理所有下载任务
- **功能特性**：
  - 显示下载状态（下载中/已完成/失败）
  - **实时显示下载进度百分比和进度条**
  - 支持重新下载、重试、删除操作
  - 支持复制下载命令
  - 自动记录下载历史（最多 10 条）
  - 支持下载封面和 MP3（在列表头部）
- **相关文件**：
  - `inject/download_list.js` - 下载列表 UI 和逻辑
  - `inject/main.js` - 下载进度更新逻辑

### 4.1 下载进度显示

- **功能描述**：在下载过程中实时显示百分比进度
- **实现方式**：
  - 在下载列表项中显示进度条和百分比文字
  - 在视频上的 loading 转圈中显示 "下载中 XX.X%"
  - 使用节流机制（每200ms更新一次）避免闪烁
- **相关代码**：
  - `inject/utils.js` - `__wx_channel_loading()` 函数，支持动态更新文本
  - `inject/main.js` - `show_progress_or_loaded_size()` 函数，计算并更新进度
  - `inject/download_list.js` - `update_download_item_progress()` 函数，更新列表进度

### 5. 直播下载

- **功能描述**：支持下载直播流
- **实现方式**：
  - 在直播页面注入下载按钮
  - 生成 FFmpeg 下载命令
  - 用户复制命令到终端执行
- **相关文件**：
  - `inject/live.js` - 直播下载逻辑

### 6. 文件名模板

- **功能描述**：支持自定义下载文件名格式
- **实现方式**：
  - 使用模板语法（如 `{{filename}}_{{spec}}`）
  - 支持变量：filename、id、title、spec、created_at、download_at、author
  - 可通过配置文件或全局脚本自定义
- **相关文件**：
  - `inject/utils.js` - `__wx_build_filename` 函数
  - `config/config.go` - 配置加载

---

## 架构设计

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                     用户浏览器                            │
│  ┌──────────────────────────────────────────────────┐   │
│  │           微信视频号页面 (channels.weixin.qq.com) │   │
│  │  ┌────────────────────────────────────────────┐  │   │
│  │  │  注入的 JavaScript (main.js, utils.js)    │  │   │
│  │  │  - 下载按钮注入                            │  │   │
│  │  │  - 视频流捕获                              │  │   │
│  │  │  - 下载列表管理                            │  │   │
│  │  └────────────────────────────────────────────┘  │   │
│  └──────────────────────────────────────────────────┘   │
└───────────────────────┬───────────────────────────────────┘
                        │ HTTP/HTTPS 请求
                        ▼
┌─────────────────────────────────────────────────────────┐
│               HTTP 代理服务器 (Interceptor)               │
│  ┌──────────────────────────────────────────────────┐   │
│  │            Echo 代理框架                          │   │
│  │  - HTTPS 中间人攻击 (MITM)                       │   │
│  │  - 证书管理                                       │   │
│  │  - 请求/响应拦截                                  │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────┐   │
│  │         ChannelInterceptorPlugin                  │   │
│  │  - HTML 注入 JavaScript                          │   │
│  │  - JavaScript 代码修改                            │   │
│  │  - API 请求拦截                                  │   │
│  └──────────────────────────────────────────────────┘   │
└───────────────────────┬───────────────────────────────────┘
                        │
                        ▼
┌─────────────────────────────────────────────────────────┐
│              本地下载服务器 (Download Server)             │
│  ┌──────────────────────────────────────────────────┐   │
│  │         MediaProxyWithDecrypt                    │   │
│  │  - 视频下载代理                                  │   │
│  │  - 视频解密                                      │   │
│  │  - FFmpeg 转换 (MP3)                             │   │
│  └──────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────┘
```

### 核心组件

#### 1. 主程序入口 (`main.go`)
- 初始化配置
- 加载嵌入的资源文件（证书、JavaScript 文件）
- 启动命令处理

#### 2. 命令处理 (`cmd/`)
- `root.go` - 主命令，启动代理和下载服务
- `download.go` - 命令行下载命令
- `decrypt.go` - 视频解密命令
- `uninstall.go` - 卸载证书命令
- `version.go` - 版本信息命令

#### 3. 拦截器 (`internal/interceptor/`)
- `interceptor.go` - 拦截器核心逻辑
- `plugin.go` - 视频号专用插件，负责 JS 注入和请求拦截
- `server.go` - HTTP 服务器封装

#### 4. 下载服务 (`internal/download/`)
- `download.go` - 下载和解密逻辑
- `server.go` - 下载服务器实现

#### 5. 服务管理 (`internal/manager/`)
- `manager.go` - 统一管理多个服务（拦截器、下载服务器）
- `server.go` - 服务接口定义

#### 6. 前端脚本 (`inject/`)
- `main.js` - 主要功能逻辑（下载按钮、视频捕获、下载）
- `utils.js` - 工具函数（文件名生成、MP3 转换等）
- `download_list.js` - 下载列表 UI
- `live.js` - 直播下载功能
- `lib/` - 第三方库（FileSaver、JSZip、Recorder）

#### 7. 平台相关 (`pkg/`)
- `platform/` - 平台特定功能（Windows/macOS/Linux）
- `proxy/` - 系统代理设置
- `certificate/` - 证书安装和管理
- `decrypt/` - 视频解密算法

---

## 核心技术点

### 1. HTTPS 中间人攻击 (MITM)

**目的**：拦截和修改 HTTPS 请求/响应

**实现方式**：
- 使用自签名根证书（SunnyNet）
- 首次运行时自动安装证书到系统信任库
- 使用 Echo 代理框架进行 HTTPS 拦截
- 动态生成目标域名的证书

**相关代码**：
- `pkg/certificate/` - 证书管理
- `internal/interceptor/interceptor.go` - 证书安装逻辑

### 2. JavaScript 注入

**目的**：在视频号页面注入自定义功能

**实现方式**：
- 拦截 HTML 响应，在 `<head>` 标签后注入脚本
- 拦截 JavaScript 文件，修改关键代码
- 使用正则表达式匹配和替换

**关键注入点**：
1. **HTML 注入**：在页面加载时注入工具脚本
2. **SourceBuffer 拦截**：修改 `appendBuffer` 方法捕获视频流
3. **解密器捕获**：拦截视频解密器数组
4. **菜单注入**：在"更多"菜单中添加下载选项

**相关代码**：
- `internal/interceptor/plugin.go` - 注入逻辑
- `inject/main.js` - 注入的脚本

### 3. 视频流捕获

**目的**：从浏览器中捕获正在播放的视频数据

**实现方式**：
- 拦截 `SourceBuffer.appendBuffer()` 调用
- 将视频数据存储到内存缓冲区
- 下载时将所有片段合并为完整视频
- 支持加密视频的解密

**关键代码**：
```javascript
// 拦截 SourceBuffer.appendBuffer
this.sourceBuffer.appendBuffer = function(buffer) {
  if (window.__wx_channels_store__) {
    window.__wx_channels_store__.buffers.push(buffer);
  }
  // 原始调用
  originalAppendBuffer.call(this, buffer);
}
```

### 4. 视频解密算法

**目的**：解密微信视频号的加密视频

**算法**：ISAAC (Indirection, Shift, Accumulate, Add, and Count) 伪随机数生成器

**实现方式**：
- 使用视频的 `decodeKey` 初始化 ISAAC 上下文
- 生成伪随机数流作为密钥流
- 对加密数据逐字节进行 XOR 解密
- 只解密前 `encLimit` 字节（通常为 131072 字节）

**相关代码**：
- `pkg/decrypt/decrypt.go` - ISAAC 算法实现
- `internal/download/download.go` - `DecryptReader` 流式解密

### 5. 前端 MP3 转换

**目的**：在浏览器端将视频转换为 MP3

**实现流程**：
1. 使用 Web Audio API 解码视频中的音频
2. 将音频数据转换为 WAV 格式
3. 使用 lame.js 将 WAV 转换为 MP3
4. 使用 FileSaver.js 下载 MP3 文件

**相关代码**：
- `inject/utils.js` - `wavBlobToMP3` 函数
- `inject/lib/recorder.min.js` - 音频处理库

### 6. 服务器端 MP3 转换

**目的**：使用 FFmpeg 转换长视频为 MP3

**实现方式**：
- 启动本地 HTTP 服务器（默认 127.0.0.1:8080）
- 接收下载请求，包含视频 URL 和解密密钥
- 下载视频流，同时进行解密
- 使用 FFmpeg 管道式转换（`pipe:0` → `pipe:1`）
- 流式返回 MP3 数据

**FFmpeg 命令**：
```bash
ffmpeg -i pipe:0 -vn -acodec libmp3lame -ab 192k -f mp3 pipe:1
```

**相关代码**：
- `internal/download/download.go` - `convertWithDecrypt` 函数

### 7. 系统代理设置

**目的**：将系统流量转发到代理服务器

**实现方式**：
- Windows: 使用注册表修改系统代理设置
- macOS: 使用 networksetup 命令
- Linux: 修改环境变量或使用系统设置

**相关代码**：
- `pkg/proxy/` - 平台特定的代理设置

### 8. 配置管理

**目的**：灵活的配置系统

**实现方式**：
- 使用 Viper 库管理配置
- 支持 YAML 配置文件
- 支持默认值和环境变量
- 配置文件自动查找（可执行文件目录、项目根目录等）

**配置项**（`config.yaml`）：

```yaml
# 调试模式
debug: false

# 下载配置
download:
  defaultHighest: false              # 是否默认下载原始视频
  filenameTemplate: "{{filename}}_{{spec}}"  # 文件名模板
  pauseWhenDownload: false           # 下载时是否暂停视频
  localServer:
    enabled: false                   # 是否启用本地下载服务器
    addr: "127.0.0.1:8080"          # 服务器地址

# 代理配置
proxy:
  system: true                      # 是否设置系统代理
  hostname: "127.0.0.1"             # 代理主机名
  port: 2023                        # 代理端口

# 视频号配置
channel:
  disableLocationToHome: false      # 是否禁用跳转到首页
```

**文件名模板变量**：
- `{{filename}}` - 视频标题
- `{{spec}}` - 视频规格（如 xWT111）
- `{{id}}` - 视频ID
- `{{title}}` - 视频标题（同 filename）
- `{{created_at}}` - 创建时间
- `{{download_at}}` - 下载时间
- `{{author}}` - 作者昵称

**相关代码**：
- `config/config.go` - 配置加载和管理
- `config/config.template.yaml` - 配置模板

---

## 实现细节

### 0. 关键代码示例

#### 0.1 下载进度更新

**前端进度计算**（`inject/main.js`）：
```javascript
async function show_progress_or_loaded_size(response, downloadItemId, loadingInstance) {
  var total_size = parseInt(response.headers.get("Content-Length"), 10);
  var loaded_size = 0;
  var reader = response.body.getReader();
  
  while (true) {
    var { done, value } = await reader.read();
    if (done) break;
    
    loaded_size += value.length;
    if (total_size) {
      var progress = (loaded_size / total_size) * 100;
      
      // 更新下载列表进度
      if (downloadItemId && window.update_download_item_progress) {
        window.update_download_item_progress(downloadItemId, progress);
      }
      
      // 更新 loading 提示
      if (loadingInstance && loadingInstance.update) {
        loadingInstance.update("下载中 " + progress.toFixed(1) + "%");
      }
    }
  }
  
  // 下载完成
  if (loadingInstance && loadingInstance.update) {
    loadingInstance.update("下载完成");
  }
}
```

**Loading 提示更新**（`inject/utils.js`）：
```javascript
function __wx_channel_loading(initialText) {
  var currentInstance = null;
  if (window.__wx_channels_tip__ && window.__wx_channels_tip__.loading) {
    currentInstance = window.__wx_channels_tip__.loading(initialText || "下载中");
    
    return {
      update: function(newText) {
        // 隐藏旧的，创建新的显示新文本
        if (currentInstance && currentInstance.hide) {
          currentInstance.hide();
        }
        currentInstance = window.__wx_channels_tip__.loading(newText);
      },
      hide: function() {
        if (currentInstance && currentInstance.hide) {
          currentInstance.hide();
        }
      }
    };
  }
  return { hide() {}, update() {} };
}
```

#### 0.2 SourceBuffer 拦截

**JavaScript 代码修改**（`internal/interceptor/plugin.go`）：
```go
// 拦截 SourceBuffer.appendBuffer 调用
replace_str1 := `(() => {
    if (window.__wx_channels_store__) {
        window.__wx_channels_store__.buffers.push($1);
    }
})(),this.sourceBuffer.appendBuffer($1),`
js_script = jsSourceBufferReg.ReplaceAllString(js_script, replace_str1)
```

#### 0.3 视频解密

**ISAAC 算法解密**（`pkg/decrypt/decrypt.go`）：
```go
func DecryptData(data []byte, encLen uint32, key uint64) {
    aaInst := CreateISAacInst(key)
    for i := uint32(0); i < encLen; i += 8 {
        randNumber := aaInst.ISAacRandom()
        tempNumber := make([]byte, 8)
        binary.BigEndian.PutUint64(tempNumber, randNumber)
        for j := 0; j < 8; j++ {
            if i + uint32(j) >= encLen {
                return
            }
            data[i + uint32(j)] ^= tempNumber[j]
        }
    }
}
```

### 1. 视频下载流程

```
用户点击下载按钮
    ↓
包装函数拦截（download_list.js）：
    - 添加到下载列表（状态：downloading）
    - 设置 downloadItemId 到 profile
    ↓
调用原始下载函数（main.js）：
    - 显示 loading 提示（"下载中 0.0%"）
    - 开始下载视频流
    ↓
下载过程中（show_progress_or_loaded_size）：
    - 读取视频流数据块
    - 计算进度百分比
    - 每200ms更新一次：
      * 下载列表进度条和百分比
      * loading 提示文字
    ↓
下载完成：
    - 合并所有数据块为 Blob
    - 更新进度为 100%
    - 显示 "下载完成"
    ↓
检查是否需要解密：
    - 如果有 key，使用 ISAAC 算法解密
    - 对前 131072 字节进行 XOR 解密
    ↓
使用 FileSaver.js 下载视频文件
    ↓
更新下载列表状态为 "completed"
    ↓
隐藏 loading 提示
```

### 2. JavaScript 注入流程

```
HTTP 请求 → 代理服务器
    ↓
拦截 HTML 响应（channels.weixin.qq.com）
    ↓
在 <head> 后注入：
    - utils.js（工具函数）
    - 配置对象（__wx_channels_config__）
    - main.js（主要功能）
    - download_list.js（下载列表）
    ↓
拦截 JavaScript 文件
    ↓
修改关键代码：
    - SourceBuffer.appendBuffer（捕获视频流）
    - 解密器数组（保存解密密钥）
    - 菜单项（添加下载选项）
    ↓
返回修改后的响应
```

### 3. 长视频下载流程

```
用户点击"打印下载命令"
    ↓
前端生成下载命令：
    download --url "视频URL" --key 解密密钥 --filename "文件名"
    ↓
用户在终端执行命令
    ↓
命令行工具：
    - 多线程下载视频
    - 下载完成后自动解密
    - 保存到 Downloads 目录
```

### 4. MP3 转换流程（服务器端）

```
用户点击"下载MP3"
    ↓
前端请求：http://127.0.0.1:8080/download?url=...&key=...&mp3=1
    ↓
下载服务器：
    - 下载视频流
    - 使用 DecryptReader 解密
    - 通过管道传递给 FFmpeg
    ↓
FFmpeg：
    - 从 pipe:0 读取视频
    - 提取音频并转换为 MP3
    - 输出到 pipe:1
    ↓
服务器：
    - 从 FFmpeg 读取 MP3 数据
    - 流式返回给浏览器
    ↓
浏览器下载 MP3 文件
```

### 5. 下载列表实现

```
页面加载时初始化下载列表
    ↓
用户点击下载按钮
    ↓
添加到下载列表（状态：下载中，进度：0%）
    ↓
开始下载，显示 loading 提示（"下载中 0.0%"）
    ↓
下载过程中：
    - 每200ms更新一次进度
    - 更新下载列表中的进度条和百分比
    - 更新 loading 提示文字（"下载中 XX.X%"）
    ↓
下载完成/失败
    ↓
更新列表项状态（已完成/失败，进度：100%/0%）
    ↓
隐藏 loading 提示
    ↓
提供操作按钮：
    - 重新下载
    - 复制命令
    - 重试
    - 删除
```

### 6. 前端全局状态管理

**全局变量**：

1. **`window.__wx_channels_config__`** - 配置对象
   - 从后端注入，包含所有配置项
   - 包括：下载配置、代理配置、调试配置等

2. **`window.__wx_channels_store__`** - 数据存储
   ```javascript
   {
     profile: null,        // 当前视频的元数据
     profiles: [],        // 所有视频的元数据列表
     keys: {},            // 解密密钥映射（key -> decryptor_array）
     buffers: []          // 视频流数据缓冲区
   }
   ```

3. **`window.__wx_channels_tip__`** - 提示工具
   - `loading(text)` - 显示加载动画（支持更新文本）
   - `toast(msg, duration)` - 显示提示消息

4. **`window.__wx_channels_download_list__`** - 下载列表状态
   ```javascript
   {
     list: [],           // 下载项列表
     container: null,    // DOM 容器
     isExpanded: false,  // 是否展开
     maxItems: 10        // 最大显示数量
   }
   ```

**下载项数据结构**：
```javascript
{
  id: "unique_id",           // 唯一标识
  profile: {...},            // 视频元数据
  spec: {...},               // 视频规格（可选）
  status: "downloading",     // 状态：downloading/completed/failed
  filename: "文件名.mp4",     // 文件名
  timestamp: 1234567890,     // 时间戳
  url: "视频URL",            // 下载URL
  key: "解密密钥",           // 解密密钥（可选）
  progress: 45.2             // 下载进度（0-100）
}
```

---

## 开发指南

### 环境要求

- Go 1.19+
- FFmpeg（仅 MP3 转换功能需要）
- 管理员权限（用于安装证书和设置系统代理）
- Windows/macOS/Linux 系统

### 开发模式检测

程序会自动检测是否为开发模式：

```go
func isDevMode() bool {
    // 1. 检查构建时标志
    if EnableLogs == "true" {
        return true
    }
    if EnableLogs == "false" {
        return false  // 明确禁用日志
    }
    
    // 2. 检查可执行文件路径
    exe, _ := os.Executable()
    exeLower := strings.ToLower(exe)
    
    // 开发模式特征：
    // - 路径包含 go-build（go run 临时目录）
    // - 文件名是 main.exe
    // - 路径包含 temp 或 tmp
    isDev := strings.Contains(exeLower, "go-build") ||
             strings.Contains(exeLower, "main.exe") ||
             strings.Contains(exeLower, "\\temp\\") ||
             strings.Contains(exeLower, "/tmp/")
    
    return isDev
}
```

**开发模式行为**：
- 打印详细日志（包括代理地址、配置信息等）
- 显示前端提示信息
- 显示视频标题、下载进度等

**生产模式行为**：
- 不打印日志（除非设置 `EnableLogs=true`）
- 只显示必要的错误信息

### 项目结构

```
wx_channels_download/
├── main.go                 # 程序入口
├── cmd/                    # 命令实现
│   ├── root.go            # 主命令
│   ├── download.go        # 下载命令
│   ├── decrypt.go         # 解密命令
│   └── ...
├── config/                 # 配置管理
│   ├── config.go          # 配置加载
│   └── config.template.yaml
├── internal/              # 内部模块
│   ├── interceptor/       # 请求拦截
│   ├── download/          # 下载服务
│   └── manager/           # 服务管理
├── inject/                # 前端脚本
│   ├── main.js           # 主要功能
│   ├── utils.js          # 工具函数
│   ├── download_list.js  # 下载列表
│   └── lib/              # 第三方库
├── pkg/                   # 公共包
│   ├── platform/         # 平台相关
│   ├── proxy/            # 代理设置
│   ├── certificate/      # 证书管理
│   └── decrypt/          # 解密算法
└── docs/                  # 文档
```

### 开发模式

运行 `go run main.go` 时：
- 自动检测为开发模式
- 打印详细日志
- 显示代理地址和配置信息

### 构建

**Windows:**
```bash
# 生产模式（不打印日志）
go build -ldflags="-s -w" -o wx_video_download.exe

# 调试模式（打印日志）
go build -ldflags="-s -w -X main.EnableLogs=true" -o wx_video_download.exe
```

**macOS:**
```bash
CGO_ENABLED=1 GOOS=darwin SDKROOT=$(xcrun --sdk macosx --show-sdk-path) go build -trimpath -ldflags="-s -w" -o wx_video_download
```

### 调试技巧

1. **开启调试模式**：
   - 在 `config.yaml` 中设置 `debug: true`
   - 打包时使用 `-X main.EnableLogs=true`

2. **查看注入的脚本**：
   - 浏览器开发者工具 → Network → 查看 HTML 响应
   - 搜索 `__wx_channels_config__` 或 `__wx_channels_store__`

3. **查看控制台日志**：
   - 浏览器开发者工具 → Console
   - 查看 `__wx_log` 输出的信息

4. **查看后端日志**：
   - 开发模式下会打印详细日志
   - 包括：视频标题、下载进度、API 请求等

5. **调试下载进度**：
   - 在控制台执行：`window.update_download_item_progress('item_id', 50)`
   - 检查 `__wx_channels_download_list__.list` 查看下载项状态

6. **调试状态管理**：
   - 在控制台查看：`window.__wx_channels_store__`
   - 查看当前视频：`window.__wx_channels_store__.profile`
   - 查看下载列表：`window.__wx_channels_download_list__.list`

7. **使用 PageSpy**（如果启用）：
   - 在 `config.yaml` 中配置 `pagespyServerAPI`
   - 可以远程调试页面状态

### 常见问题

1. **证书安装失败**：需要管理员权限
2. **代理设置失败**：检查是否有其他代理软件冲突
3. **FFmpeg 未找到**：确保 FFmpeg 已安装并添加到 PATH
4. **下载按钮不显示**：检查 JavaScript 注入是否成功
5. **下载进度不显示**：
   - 检查 `downloadItemId` 是否正确传递
   - 检查 `update_download_item_progress` 函数是否已暴露到全局
   - 检查浏览器控制台是否有错误
6. **打包后仍有日志输出**：
   - 确保使用 `go build -ldflags="-s -w"` 打包（不设置 EnableLogs）
   - 检查 `isDevMode()` 函数逻辑是否正确
7. **下载列表不显示**：
   - 检查 `download_list.js` 是否正确注入
   - 检查是否有 JavaScript 错误
   - 尝试右键点击悬浮下载按钮或长按显示列表

---

## 数据流和状态转换

### 下载状态转换图

```
初始状态
    ↓
[用户点击下载]
    ↓
downloading (进度: 0%)
    ↓
[下载中，实时更新进度]
    ↓
[下载完成]
    ↓
completed (进度: 100%)
    ↓
[用户点击重新下载] → downloading
    ↓
[下载失败]
    ↓
failed (进度: 0%)
    ↓
[用户点击重试] → downloading
```

### 关键数据流

1. **视频元数据流**：
   ```
   微信页面 JavaScript
       ↓ (拦截并修改)
   注入代码捕获 profile
       ↓
   __wx_channels_store__.profile
       ↓
   下载函数使用 profile
   ```

2. **视频流数据流**：
   ```
   SourceBuffer.appendBuffer()
       ↓ (拦截)
   __wx_channels_store__.buffers.push(buffer)
       ↓
   下载时合并所有 buffers
       ↓
   new Blob(buffers)
       ↓
   FileSaver.js 下载
   ```

3. **进度更新流**：
   ```
   show_progress_or_loaded_size()
       ↓ (计算进度)
   update_download_item_progress(id, progress)
       ↓
   更新下载列表 UI
       ↓
   loadingInstance.update("下载中 XX%")
       ↓
   更新 loading 提示
   ```

## 重要注意事项

### 1. 日志控制

- **开发模式**：自动检测 `go run` 或临时目录，打印所有日志
- **生产模式**：默认不打印日志，除非设置 `EnableLogs=true`
- **关键点**：`isDevMode()` 函数决定是否打印日志

### 2. 下载进度实现

- **节流机制**：每200ms更新一次，避免频繁更新导致闪烁
- **双重显示**：同时在下载列表和 loading 提示中显示进度
- **状态同步**：确保 `downloadItemId` 正确传递到下载函数

### 3. 前端状态管理

- **全局变量**：使用 `window` 对象存储全局状态
- **数据持久化**：下载列表数据仅保存在内存中，刷新页面后清空
- **状态更新**：通过函数调用更新状态，然后重新渲染 UI

### 4. 错误处理

- **下载失败**：捕获异常，更新状态为 `failed`，显示错误信息
- **解密失败**：提示用户使用命令行解密
- **网络错误**：显示错误提示，允许重试

## 参考资源

- [Echo 代理框架](https://github.com/ltaoo/echo)
- [ISAAC 算法](https://en.wikipedia.org/wiki/ISAAC_(cipher))
- [WechatSphDecrypt](https://github.com/Hanson/WechatSphDecrypt) - 解密算法参考
- [WechatVideoSniffer](https://github.com/kanadeblisst00/WechatVideoSniffer2.0) - 前端解密参考

## 更新日志

### 最新功能（2025-12-06）

1. **下载进度显示**：
   - 在下载列表中显示实时进度条和百分比
   - 在 loading 提示中显示 "下载中 XX%"
   - 支持节流更新，避免闪烁

2. **日志控制优化**：
   - 通过 `EnableLogs` 构建参数控制日志输出
   - 生产模式默认不打印日志
   - 开发模式自动检测并打印详细日志

3. **下载列表增强**：
   - 支持下载封面和 MP3（在列表头部）
   - 显示下载进度和状态
   - 支持重新下载、重试、删除等操作

---

## 许可证

本项目为开源项目，仅用于技术交流学习和研究的目的。

