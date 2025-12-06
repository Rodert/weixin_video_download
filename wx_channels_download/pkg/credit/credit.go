package credit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"
)

const (
	// CreditCostPerDownload 每次下载消耗的积分
	CreditCostPerDownload = 5
)

var (
	// EncryptionKey 加密密钥（32字节，硬编码在代码中）
	// 注意：必须是32字节（AES-256）
	EncryptionKey = []byte("wx_channels_credit_key_2025_32b!")
)

// CreditInfo 积分信息
type CreditInfo struct {
	Points  int64 `json:"points"`   // 积分数量
	StartAt int64 `json:"start_at"` // 开始时间戳（Unix时间戳，当天0点）
	EndAt   int64 `json:"end_at"`   // 结束时间戳（Unix时间戳，当天23:59:59）
}

var (
	// creditMutex 保护配置更新的互斥锁
	creditMutex sync.Mutex
)

// EncryptCreditInfo 加密积分信息
func EncryptCreditInfo(info *CreditInfo) (string, error) {
	// 序列化为 JSON
	data, err := json.Marshal(info)
	if err != nil {
		return "", fmt.Errorf("序列化失败: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher([]byte(EncryptionKey))
	if err != nil {
		return "", fmt.Errorf("创建cipher失败: %w", err)
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %w", err)
	}

	// 生成 nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("生成nonce失败: %w", err)
	}

	// 加密
	ciphertext := gcm.Seal(nonce, nonce, data, nil)

	// Base64 编码
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptCreditInfo 解密积分信息
func DecryptCreditInfo(encrypted string) (*CreditInfo, error) {
	if encrypted == "" {
		return nil, errors.New("加密数据为空")
	}

	// Base64 解码
	ciphertext, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, fmt.Errorf("Base64解码失败: %w", err)
	}

	// 创建 AES cipher
	block, err := aes.NewCipher(EncryptionKey)
	if err != nil {
		return nil, fmt.Errorf("创建cipher失败: %w", err)
	}

	// 创建 GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("创建GCM失败: %w", err)
	}

	// 提取 nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("加密数据格式错误")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// 解密
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("解密失败: %w", err)
	}

	// 反序列化
	var info CreditInfo
	if err := json.Unmarshal(plaintext, &info); err != nil {
		return nil, fmt.Errorf("反序列化失败: %w", err)
	}

	return &info, nil
}

// CheckCredit 检查积分是否足够
func CheckCredit(encrypted string) (bool, *CreditInfo, error) {
	if encrypted == "" {
		return false, nil, errors.New("未配置积分")
	}

	info, err := DecryptCreditInfo(encrypted)
	if err != nil {
		return false, nil, err
	}

	// 检查当前时间是否在有效区间内
	now := time.Now().Unix()
	if now < info.StartAt {
		startTime := time.Unix(info.StartAt, 0)
		return false, info, fmt.Errorf("积分尚未生效，生效时间: %s", startTime.Format("2006-01-02 15:04:05"))
	}
	if now > info.EndAt {
		endTime := time.Unix(info.EndAt, 0)
		return false, info, fmt.Errorf("积分已过期，过期时间: %s", endTime.Format("2006-01-02 15:04:05"))
	}

	// 检查积分是否足够
	if info.Points < CreditCostPerDownload {
		return false, info, fmt.Errorf("积分不足，当前: %d，需要: %d", info.Points, CreditCostPerDownload)
	}

	return true, info, nil
}

// ConsumeCredit 消耗积分并返回新的加密数据
func ConsumeCredit(encrypted string) (string, *CreditInfo, error) {
	if encrypted == "" {
		return "", nil, errors.New("未配置积分")
	}

	info, err := DecryptCreditInfo(encrypted)
	if err != nil {
		return "", nil, err
	}

	// 检查当前时间是否在有效区间内
	now := time.Now().Unix()
	if now < info.StartAt {
		startTime := time.Unix(info.StartAt, 0)
		return "", info, fmt.Errorf("积分尚未生效，生效时间: %s", startTime.Format("2006-01-02 15:04:05"))
	}
	if now > info.EndAt {
		endTime := time.Unix(info.EndAt, 0)
		return "", info, fmt.Errorf("积分已过期，过期时间: %s", endTime.Format("2006-01-02 15:04:05"))
	}

	// 检查积分是否足够
	if info.Points < CreditCostPerDownload {
		return "", info, fmt.Errorf("积分不足，当前: %d，需要: %d", info.Points, CreditCostPerDownload)
	}

	// 扣除积分
	oldPoints := info.Points
	info.Points -= CreditCostPerDownload

	// 重新加密
	newEncrypted, err := EncryptCreditInfo(info)
	if err != nil {
		// 加密失败，恢复原积分
		info.Points = oldPoints
		return "", nil, fmt.Errorf("加密失败: %w", err)
	}

	return newEncrypted, info, nil
}

// UpdateCreditInKeyFile 线程安全地更新独立的密钥文件
func UpdateCreditInKeyFile(baseDir string, newEncrypted string) error {
	creditMutex.Lock()
	defer creditMutex.Unlock()

	if baseDir == "" {
		return errors.New("基础目录路径为空")
	}

	// 密钥文件路径
	keyPath := filepath.Join(baseDir, "credit.yaml")

	// 1. 读取现有密钥文件（如果存在）
	viperKey := viper.New()
	viperKey.SetConfigFile(keyPath)
	viperKey.SetConfigType("yaml")

	var existingData map[string]interface{}
	if _, err := os.Stat(keyPath); err == nil {
		if err := viperKey.ReadInConfig(); err == nil {
			existingData = viperKey.AllSettings()
		}
	}

	// 2. 备份原密钥文件（可选，用于回滚）
	backupPath := keyPath + ".backup"
	if _, err := os.Stat(keyPath); err == nil {
		_ = copyFile(keyPath, backupPath) // 备份失败不影响主流程
	}

	// 3. 更新积分
	if existingData == nil {
		existingData = make(map[string]interface{})
	}
	existingData["encrypted"] = newEncrypted

	// 4. 创建新的 viper 实例用于写入
	viperWrite := viper.New()
	viperWrite.SetConfigType("yaml")
	for key, value := range existingData {
		viperWrite.Set(key, value)
	}

	// 5. 原子性写入（先写临时文件，再重命名）
	// 使用 .tmp.yaml 扩展名，确保 Viper 能识别为 YAML 格式
	tempPath := strings.TrimSuffix(keyPath, ".yaml") + ".tmp.yaml"
	if err := viperWrite.WriteConfigAs(tempPath); err != nil {
		return fmt.Errorf("写入临时密钥文件失败: %w", err)
	}

	// 6. 原子性替换（重命名是原子操作）
	if err := os.Rename(tempPath, keyPath); err != nil {
		// 如果失败，尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			_ = os.Rename(backupPath, keyPath)
		}
		return fmt.Errorf("更新密钥文件失败: %w", err)
	}

	// 7. 删除备份（成功后才删除）
	_ = os.Remove(backupPath)

	return nil
}

// UpdateCreditInConfig 线程安全地更新配置文件中的积分（已废弃，保留用于兼容）
// 建议使用 UpdateCreditInKeyFile
func UpdateCreditInConfig(configPath string, newEncrypted string) error {
	// 获取配置文件所在目录
	baseDir := filepath.Dir(configPath)
	return UpdateCreditInKeyFile(baseDir, newEncrypted)
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ParseDate 解析日期字符串，支持 "2006.01.02" 和 "2006-01-02" 格式
func ParseDate(dateStr string) (time.Time, error) {
	// 尝试多种日期格式
	formats := []string{
		"2006.01.02",
		"2006-01-02",
		"2006/01/02",
		"20060102",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("无法解析日期格式: %s，支持的格式: 2006.01.02, 2006-01-02, 2006/01/02, 20060102", dateStr)
}

// GenerateCreditInfo 生成新的积分信息（用于生成配置）
// startDate: 开始日期，格式 "2006.01.02" 或 "2006-01-02"
// endDate: 结束日期，格式 "2006.01.02" 或 "2006-01-02"
func GenerateCreditInfo(points int64, startDate, endDate string) (*CreditInfo, error) {
	// 解析开始日期，设置为当天的 00:00:00
	startTime, err := ParseDate(startDate)
	if err != nil {
		return nil, fmt.Errorf("解析开始日期失败: %w", err)
	}
	startAt := time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 0, 0, 0, 0, time.Local).Unix()

	// 解析结束日期，设置为当天的 23:59:59
	endTime, err := ParseDate(endDate)
	if err != nil {
		return nil, fmt.Errorf("解析结束日期失败: %w", err)
	}
	endAt := time.Date(endTime.Year(), endTime.Month(), endTime.Day(), 23, 59, 59, 999999999, time.Local).Unix()

	// 验证日期区间
	if startAt > endAt {
		return nil, errors.New("开始日期不能晚于结束日期")
	}

	return &CreditInfo{
		Points:  points,
		StartAt: startAt,
		EndAt:   endAt,
	}, nil
}

// GetCreditInfo 获取积分信息（不检查有效性，仅解密）
func GetCreditInfo(encrypted string) (*CreditInfo, error) {
	if encrypted == "" {
		return nil, errors.New("未配置积分")
	}
	return DecryptCreditInfo(encrypted)
}
