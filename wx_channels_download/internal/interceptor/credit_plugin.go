package interceptor

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ltaoo/echo"
	"github.com/spf13/viper"

	"wx_channel/config"
	"wx_channel/pkg/credit"
)

// getCreditBaseDir 获取积分密钥文件所在目录（用于检查 .use 文件）
// 优先使用 credit.txt 文件所在的目录，确保与加载密钥时的目录一致
func getCreditBaseDir(cfg *config.Config) string {
	baseDir := ""
	if exe, err := os.Executable(); err == nil {
		baseDir = filepath.Dir(exe)
		// 检查 credit.txt 是否在可执行文件目录
		creditPath := filepath.Join(baseDir, "credit.txt")
		if _, err := os.Stat(creditPath); err != nil {
			// 如果 credit.txt 不存在，尝试检查 credit.yaml（兼容旧版本）
			oldCreditPath := filepath.Join(baseDir, "credit.yaml")
			if _, err := os.Stat(oldCreditPath); err != nil {
				// 如果都不在可执行文件目录，尝试使用配置文件目录
				if cfg.FilePath != "" {
					baseDir = filepath.Dir(cfg.FilePath)
				}
			}
		}
	} else if cfg.FilePath != "" {
		baseDir = filepath.Dir(cfg.FilePath)
	}
	return baseDir
}

// CreateCreditPlugin 创建积分管理插件（独立模块，可选启用）
func CreateCreditPlugin(cfg *config.Config) *echo.Plugin {
	return &echo.Plugin{
		Match: "qq.com",
		OnRequest: func(ctx *echo.Context) {
			pathname := ctx.Req.URL.Path

			// 积分检查 API
			if pathname == "/__wx_channels_api/credit/check" {
				handleCreditCheck(ctx, cfg)
				return
			}

			// 积分消耗 API
			if pathname == "/__wx_channels_api/credit/consume" {
				handleCreditConsume(ctx, cfg)
				return
			}
		},
	}
}

// handleCreditCheck 处理积分检查请求
func handleCreditCheck(ctx *echo.Context, cfg *config.Config) {
	// 获取 baseDir（用于检查 .use 文件和读取最新的密钥）
	baseDir := getCreditBaseDir(cfg)

	// 从文件重新读取 encrypted 值（确保使用最新的值，而不是内存中的旧值）
	encrypted := ""
	if baseDir != "" {
		keyPath := filepath.Join(baseDir, "credit.txt")
		if _, err := os.Stat(keyPath); err == nil {
			// 读取文本文件
			data, err := os.ReadFile(keyPath)
			if err == nil {
				content := strings.TrimSpace(string(data))
				if strings.HasPrefix(content, "encrypted=") {
					encrypted = strings.TrimPrefix(content, "encrypted=")
				} else {
					// 兼容旧格式（直接是 encrypted 值）
					if idx := strings.Index(content, "\n"); idx > 0 {
						content = content[:idx]
					}
					encrypted = strings.TrimSpace(content)
				}
			}
		} else {
			// 兼容旧版本：尝试读取 credit.yaml
			oldKeyPath := filepath.Join(baseDir, "credit.yaml")
			if _, err := os.Stat(oldKeyPath); err == nil {
				viperKey := viper.New()
				viperKey.SetConfigFile(oldKeyPath)
				viperKey.SetConfigType("yaml")
				if err := viperKey.ReadInConfig(); err == nil {
					encrypted = viperKey.GetString("encrypted")
				}
			}
		}
	}

	// 如果从文件读取失败，使用内存中的值（兼容旧逻辑）
	if encrypted == "" {
		encrypted = cfg.CreditEncrypted
	}

	if encrypted == "" {
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"valid":false,"error":"未配置积分"}`)
		return
	}

	// 解析请求体，获取需要的积分数（默认为视频下载的消耗量）
	var requestBody map[string]interface{}
	var cost int64 = credit.CreditCostPerDownload
	if ctx.Req.Body != nil {
		bodyData, _ := io.ReadAll(ctx.Req.Body)
		if len(bodyData) > 0 {
			json.Unmarshal(bodyData, &requestBody)
			if costVal, ok := requestBody["cost"].(float64); ok {
				cost = int64(costVal)
			} else if costVal, ok := requestBody["cost"].(int64); ok {
				cost = costVal
			}
		}
	}

	valid, info, err := credit.CheckCreditWithBaseDir(encrypted, baseDir, cost)
	if err != nil {
		response := map[string]interface{}{
			"valid": false,
			"error": err.Error(),
		}
		if info != nil {
			response["points"] = info.Points
			response["start_at"] = info.StartAt
			response["end_at"] = info.EndAt
		}
		jsonData, _ := json.Marshal(response)
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, string(jsonData))
		return
	}

	response := map[string]interface{}{
		"valid":      valid,
		"points":     info.Points,
		"start_at":   info.StartAt,
		"end_at":     info.EndAt,
		"expires_in": info.EndAt - time.Now().Unix(),
	}

	jsonData, _ := json.Marshal(response)
	ctx.Mock(200, map[string]string{
		"Content-Type": "application/json",
	}, string(jsonData))
}

// handleCreditConsume 处理积分消耗请求
func handleCreditConsume(ctx *echo.Context, cfg *config.Config) {
	// 获取 baseDir（用于检查 .use 文件和读取最新的密钥）
	baseDir := getCreditBaseDir(cfg)

	if baseDir == "" {
		response := map[string]interface{}{
			"success": false,
			"error":   "无法确定密钥文件目录",
		}
		jsonData, _ := json.Marshal(response)
		ctx.Mock(500, map[string]string{
			"Content-Type": "application/json",
		}, string(jsonData))
		return
	}

	// 从文件重新读取 encrypted 值（确保使用最新的值，而不是内存中的旧值）
	encrypted := ""
	keyPath := filepath.Join(baseDir, "credit.txt")
	if _, err := os.Stat(keyPath); err == nil {
		// 读取文本文件
		data, err := os.ReadFile(keyPath)
		if err == nil {
			content := strings.TrimSpace(string(data))
			if strings.HasPrefix(content, "encrypted=") {
				encrypted = strings.TrimPrefix(content, "encrypted=")
			} else {
				// 兼容旧格式（直接是 encrypted 值）
				if idx := strings.Index(content, "\n"); idx > 0 {
					content = content[:idx]
				}
				encrypted = strings.TrimSpace(content)
			}
		}
	} else {
		// 兼容旧版本：尝试读取 credit.yaml
		oldKeyPath := filepath.Join(baseDir, "credit.yaml")
		if _, err := os.Stat(oldKeyPath); err == nil {
			viperKey := viper.New()
			viperKey.SetConfigFile(oldKeyPath)
			viperKey.SetConfigType("yaml")
			if err := viperKey.ReadInConfig(); err == nil {
				encrypted = viperKey.GetString("encrypted")
			}
		}
	}

	// 如果从文件读取失败，使用内存中的值（兼容旧逻辑）
	if encrypted == "" {
		encrypted = cfg.CreditEncrypted
	}

	if encrypted == "" {
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"success":false,"error":"未配置积分"}`)
		return
	}

	// 解析请求体，获取消耗量（默认为视频下载的消耗量）
	var requestBody map[string]interface{}
	var cost int64 = credit.CreditCostPerDownload
	if ctx.Req.Body != nil {
		bodyData, _ := io.ReadAll(ctx.Req.Body)
		if len(bodyData) > 0 {
			json.Unmarshal(bodyData, &requestBody)
			if costVal, ok := requestBody["cost"].(float64); ok {
				cost = int64(costVal)
			} else if costVal, ok := requestBody["cost"].(int64); ok {
				cost = costVal
			}
		}
	}

	// 消耗积分（计算新的加密数据，传入 baseDir 用于检查 .use 文件）
	newEncrypted, info, err := credit.ConsumeCreditWithBaseDir(encrypted, baseDir, cost)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		if info != nil {
			response["points"] = info.Points
			response["start_at"] = info.StartAt
			response["end_at"] = info.EndAt
		}
		jsonData, _ := json.Marshal(response)
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, string(jsonData))
		return
	}

	// 更新密钥文件（线程安全，原子操作）
	if err := credit.UpdateCreditInKeyFile(baseDir, newEncrypted); err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "更新配置失败: " + err.Error(),
		}
		jsonData, _ := json.Marshal(response)
		ctx.Mock(500, map[string]string{
			"Content-Type": "application/json",
		}, string(jsonData))
		return
	}

	// 更新内存中的配置（重要！）
	cfg.CreditEncrypted = newEncrypted

	// 返回成功响应
	response := map[string]interface{}{
		"success":  true,
		"points":   info.Points,
		"start_at": info.StartAt,
		"end_at":   info.EndAt,
	}

	jsonData, _ := json.Marshal(response)
	ctx.Mock(200, map[string]string{
		"Content-Type": "application/json",
	}, string(jsonData))
}
