package interceptor

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ltaoo/echo"

	"wx_channel/config"
	"wx_channel/pkg/credit"
)

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
	encrypted := cfg.CreditEncrypted
	if encrypted == "" {
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"valid":false,"error":"未配置积分"}`)
		return
	}

	valid, info, err := credit.CheckCredit(encrypted)
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
		"valid":     valid,
		"points":    info.Points,
		"start_at":  info.StartAt,
		"end_at":    info.EndAt,
		"expires_in": info.EndAt - time.Now().Unix(),
	}

	jsonData, _ := json.Marshal(response)
	ctx.Mock(200, map[string]string{
		"Content-Type": "application/json",
	}, string(jsonData))
}

// handleCreditConsume 处理积分消耗请求
func handleCreditConsume(ctx *echo.Context, cfg *config.Config) {
	encrypted := cfg.CreditEncrypted
	if encrypted == "" {
		ctx.Mock(200, map[string]string{
			"Content-Type": "application/json",
		}, `{"success":false,"error":"未配置积分"}`)
		return
	}

	// 消耗积分（计算新的加密数据）
	newEncrypted, info, err := credit.ConsumeCredit(encrypted)
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
	// 获取配置文件所在目录（如果 FilePath 为空，使用可执行文件目录）
	baseDir := ""
	if cfg.FilePath != "" {
		baseDir = filepath.Dir(cfg.FilePath)
	} else {
		// 如果配置文件路径为空，尝试获取可执行文件目录
		if exe, err := os.Executable(); err == nil {
			baseDir = filepath.Dir(exe)
		}
	}
	
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
		"success": true,
		"points":  info.Points,
		"start_at": info.StartAt,
		"end_at":   info.EndAt,
	}

	jsonData, _ := json.Marshal(response)
	ctx.Mock(200, map[string]string{
		"Content-Type": "application/json",
	}, string(jsonData))
}

