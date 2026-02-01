package credit

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	// CreditCostPerDownload 每次下载视频消耗的积分
	CreditCostPerDownload = 5
	// CreditCostPerCover 每次下载封面消耗的积分
	CreditCostPerCover = 1
)

var (
	// EncryptionKey 加密密钥（32字节，硬编码在代码中）
	// 注意：必须是32字节（AES-256）
	EncryptionKey = []byte("wx_channels_credit_key_2025_32b!")
)

// CreditInfo 积分信息
type CreditInfo struct {
	Version string `json:"version"` // 版本号（用于版本隔离，v1, v2, v3, v4, v5...）
	Points  int64  `json:"points"`  // 积分数量
	StartAt int64  `json:"start_at"` // 开始时间戳（Unix时间戳，当天0点）
	EndAt   int64  `json:"end_at"`   // 结束时间戳（Unix时间戳，当天23:59:59）
}

var (
	// creditMutex 保护配置更新的互斥锁
	creditMutex sync.Mutex
)

// ReadVersion 从 version.txt 读取版本号，失败时返回默认版本 v1
// 支持任意版本号格式：v1, v2, v3, v4, v5...
// 在开发模式下（go run），会尝试从当前工作目录和源代码目录读取
func ReadVersion(baseDir string) (string, error) {
	if baseDir == "" {
		if exe, err := os.Executable(); err == nil {
			baseDir = filepath.Dir(exe)
		}
	}

	// 尝试读取版本文件的路径列表（按优先级）
	var versionPaths []string

	// 1. 如果指定了 baseDir，优先使用
	if baseDir != "" {
		versionPaths = append(versionPaths, filepath.Join(baseDir, "version.txt"))
	}

	// 2. 尝试从当前工作目录读取（开发模式：go run）
	if wd, err := os.Getwd(); err == nil {
		versionPaths = append(versionPaths, filepath.Join(wd, "version.txt"))
		// 如果在子目录中运行，尝试从父目录读取
		parentDir := filepath.Dir(wd)
		if parentDir != wd {
			versionPaths = append(versionPaths, filepath.Join(parentDir, "version.txt"))
		}
	}

	// 3. 尝试从可执行文件目录读取（如果 baseDir 为空）
	if baseDir == "" {
		if exe, err := os.Executable(); err == nil {
			exeDir := filepath.Dir(exe)
			versionPaths = append(versionPaths, filepath.Join(exeDir, "version.txt"))
		}
	}

	// 按优先级尝试读取
	for _, versionPath := range versionPaths {
		data, err := os.ReadFile(versionPath)
		if err == nil {
			version := strings.TrimSpace(string(data))
			if version != "" {
				// 支持任意版本号格式（v1, v2, v3, v4, v5...）
				return version, nil
			}
		}
	}

	// 所有路径都读取失败，返回默认版本 v1
	return "v1", nil
}

// GetCurrentVersion 获取当前版本号（读取失败时返回 v1）
func GetCurrentVersion() string {
	version, _ := ReadVersion("") // 忽略错误，失败时返回 v1
	return version
}

// EncryptCreditInfo 加密积分信息（自动包含版本号）
func EncryptCreditInfo(info *CreditInfo) (string, error) {
	// 如果版本号为空，自动读取当前版本号
	if info.Version == "" {
		info.Version = GetCurrentVersion() // 失败时返回 v1
	}

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

	// 如果版本号为空（兼容旧版本），默认使用 v1
	if info.Version == "" {
		info.Version = "v1"
	}

	// 读取当前版本号（失败时返回 v1）
	currentVersion := GetCurrentVersion()

	// 验证版本号是否匹配（支持任意版本号：v1, v2, v3, v4, v5...）
	if info.Version != currentVersion {
		return nil, fmt.Errorf("版本不匹配：密钥版本 %s，当前版本 %s", info.Version, currentVersion)
	}

	return &info, nil
}

// hashKey 对密钥进行哈希编码，用于存储到 .use 文件（二进制格式，不可读）
func hashKey(encrypted string) string {
	h := sha256.Sum256([]byte(encrypted))
	return hex.EncodeToString(h[:])
}

// isKeyUsed 检查密钥是否已被使用（在 C:\.use 文件中）
func isKeyUsed(encrypted string) (bool, error) {
	if encrypted == "" {
		return false, nil
	}

	useFilePath := `C:\.use`

	if _, err := os.Stat(useFilePath); os.IsNotExist(err) {
		return false, nil
	}

	data, err := os.ReadFile(useFilePath)
	if err != nil {
		return false, err
	}

	keyHash := hashKey(encrypted)

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == keyHash {
			return true, nil
		}
	}

	return false, nil
}

// recordUsedKey 将已用完或失效的密钥记录到 .use 文件
func recordUsedKey(baseDir, encrypted string) error {
	if encrypted == "" {
		return nil
	}

	// 如果 baseDir 为空，尝试使用可执行文件目录
	checkBaseDir := baseDir
	if checkBaseDir == "" {
		if exe, err := os.Executable(); err == nil {
			checkBaseDir = filepath.Dir(exe)
		}
	}

	if checkBaseDir == "" {
		return nil // 无法确定目录，跳过记录
	}

	useFilePath := "C:\\.use"

	// 检查是否已存在
	used, err := isKeyUsed(encrypted)
	if err != nil {
		return fmt.Errorf("检查密钥使用状态失败: %w", err)
	}
	if used {
		return nil // 已存在，不需要重复记录
	}

	// 对密钥进行哈希编码
	keyHash := hashKey(encrypted)

	// 追加到文件（每行一个哈希值）
	file, err := os.OpenFile(useFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开 .use 文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(keyHash + "\n")
	if err != nil {
		return fmt.Errorf("写入 .use 文件失败: %w", err)
	}

	return nil
}

// CheckCredit 检查积分是否足够（兼容旧版本，不检查 .use 文件）
func CheckCredit(encrypted string) (bool, *CreditInfo, error) {
	return CheckCreditWithBaseDir(encrypted, "", CreditCostPerDownload)
}

// CheckCreditWithBaseDir 检查积分是否足够（带 baseDir 参数，用于检查 .use 文件）
// cost: 需要的积分数，默认为 CreditCostPerDownload
func CheckCreditWithBaseDir(encrypted, baseDir string, cost int64) (bool, *CreditInfo, error) {
	if cost <= 0 {
		cost = CreditCostPerDownload
	}
	if encrypted == "" {
		return false, nil, errors.New("未配置积分")
	}

	// 检查密钥是否已被使用（必须在解密之前检查，且即使 baseDir 为空也要尝试检查）
	// 如果 baseDir 为空，尝试使用可执行文件目录
	checkBaseDir := baseDir
	if checkBaseDir == "" {
		if exe, err := os.Executable(); err == nil {
			checkBaseDir = filepath.Dir(exe)
		}
	}

	if checkBaseDir != "" {
		used, err := isKeyUsed(encrypted)
		if err != nil {
			return false, nil, fmt.Errorf("检查密钥使用状态失败: %w", err)
		}
		if used {
			return false, nil, errors.New("该密钥已被使用，无法重复使用")
		}
	}

	info, err := DecryptCreditInfo(encrypted)
	if err != nil {
		return false, nil, err
	}

	now := time.Now().Unix()

	// 如果 StartAt 还未生效（未激活状态），从当前时间开始计算，保持原有时长
	// 这样可以在检查时就判断激活后的有效期
	checkEndAt := info.EndAt
	if now < info.StartAt {
		duration := info.EndAt - info.StartAt
		checkEndAt = now + duration
	}

	// 检查当前时间是否在有效区间内（使用激活后的时间）
	if now > checkEndAt {
		// 积分已过期，记录到 .use 文件（即使 baseDir 为空也会尝试记录）
		_ = recordUsedKey(baseDir, encrypted) // 记录失败不影响返回错误
		endTime := time.Unix(checkEndAt, 0)
		return false, info, fmt.Errorf("积分已过期，过期时间: %s", endTime.Format("2006-01-02 15:04:05"))
	}

	// 检查积分是否足够
	if info.Points < cost {
		// 积分不足，如果积分为0，记录到 .use 文件（即使 baseDir 为空也会尝试记录）
		if info.Points <= 0 {
			_ = recordUsedKey(baseDir, encrypted) // 记录失败不影响返回错误
		}
		return false, info, fmt.Errorf("积分不足，当前: %d，需要: %d", info.Points, cost)
	}

	return true, info, nil
}

// ConsumeCredit 消耗积分并返回新的加密数据（兼容旧版本，不检查 .use 文件）
func ConsumeCredit(encrypted string) (string, *CreditInfo, error) {
	return ConsumeCreditWithBaseDir(encrypted, "", CreditCostPerDownload)
}

// ConsumeCreditWithBaseDir 消耗积分并返回新的加密数据
// 功能：
// 1. 从用户开始使用的时间算起（如果 StartAt 还未生效，从当前时间开始计算，保持原有时长）
// 2. 积分用完或失效时，记录到 .use 文件（使用哈希值，不可读）
// cost: 消耗的积分数，默认为 CreditCostPerDownload
func ConsumeCreditWithBaseDir(encrypted, baseDir string, cost int64) (string, *CreditInfo, error) {
	if cost <= 0 {
		cost = CreditCostPerDownload
	}
	if encrypted == "" {
		return "", nil, errors.New("未配置积分")
	}

	// 检查密钥是否已被使用（必须在解密之前检查，且即使 baseDir 为空也要尝试检查）
	// 如果 baseDir 为空，尝试使用可执行文件目录
	checkBaseDir := baseDir
	if checkBaseDir == "" {
		if exe, err := os.Executable(); err == nil {
			checkBaseDir = filepath.Dir(exe)
		}
	}

	if checkBaseDir != "" {
		used, err := isKeyUsed(encrypted)
		if err != nil {
			return "", nil, fmt.Errorf("检查密钥使用状态失败: %w", err)
		}
		if used {
			return "", nil, errors.New("该密钥已被使用，无法重复使用")
		}
	}

	info, err := DecryptCreditInfo(encrypted)
	if err != nil {
		return "", nil, err
	}

	now := time.Now().Unix()
	originalStartAt := info.StartAt
	originalEndAt := info.EndAt

	// 如果 StartAt 还未生效，从当前时间开始计算，保持原有时长
	if now < info.StartAt {
		duration := info.EndAt - info.StartAt
		info.StartAt = now
		info.EndAt = now + duration
	}

	// 检查当前时间是否在有效区间内
	if now > info.EndAt {
		// 积分已过期，记录到 .use 文件（即使 baseDir 为空也会尝试记录）
		_ = recordUsedKey(baseDir, encrypted) // 记录失败不影响返回错误
		endTime := time.Unix(info.EndAt, 0)
		return "", info, fmt.Errorf("积分已过期，过期时间: %s", endTime.Format("2006-01-02 15:04:05"))
	}

	// 检查积分是否足够
	if info.Points < cost {
		// 积分不足，如果积分为0，记录到 .use 文件（即使 baseDir 为空也会尝试记录）
		if info.Points <= 0 {
			_ = recordUsedKey(baseDir, encrypted) // 记录失败不影响返回错误
		}
		return "", info, fmt.Errorf("积分不足，当前: %d，需要: %d", info.Points, cost)
	}

	// 保存原始加密字符串（用于记录到 .use 文件）
	originalEncrypted := encrypted

	// 扣除积分
	oldPoints := info.Points
	info.Points -= cost

	// 重新加密
	newEncrypted, err := EncryptCreditInfo(info)
	if err != nil {
		// 加密失败，恢复原积分和时间
		info.Points = oldPoints
		info.StartAt = originalStartAt
		info.EndAt = originalEndAt
		return "", nil, fmt.Errorf("加密失败: %w", err)
	}

	// 每次下载后，都记录原始密钥到 .use 文件（防止重复使用）
	// 使用 checkBaseDir 确保与检查时使用相同的目录
	if err := recordUsedKey(checkBaseDir, originalEncrypted); err != nil {
		// 记录失败不影响主流程，只记录错误
	}

	return newEncrypted, info, nil
}

// UpdateCreditInKeyFile 线程安全地更新独立的密钥文件（改为 credit.txt）
func UpdateCreditInKeyFile(baseDir string, newEncrypted string) error {
	creditMutex.Lock()
	defer creditMutex.Unlock()

	if baseDir == "" {
		return errors.New("基础目录路径为空")
	}

	// 密钥文件路径（改为 credit.txt）
	keyPath := filepath.Join(baseDir, "credit.txt")

	// 备份原密钥文件（可选，用于回滚）
	backupPath := keyPath + ".backup"
	if _, err := os.Stat(keyPath); err == nil {
		_ = copyFile(keyPath, backupPath) // 备份失败不影响主流程
	}

	// 原子性写入（先写临时文件，再重命名）
	tempPath := keyPath + ".tmp"

	// 写入简单文本格式：encrypted=xxx
	content := "encrypted=" + newEncrypted + "\n"
	if err := os.WriteFile(tempPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("写入临时密钥文件失败: %w", err)
	}

	// 原子性替换（重命名是原子操作）
	if err := os.Rename(tempPath, keyPath); err != nil {
		// 如果失败，尝试恢复备份
		if _, err2 := os.Stat(backupPath); err2 == nil {
			_ = os.Rename(backupPath, keyPath)
		}
		_ = os.Remove(tempPath) // 清理临时文件
		return fmt.Errorf("更新密钥文件失败: %w", err)
	}

	// 删除备份（成功后才删除）
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

// GenerateCreditInfoByDays 生成新的积分信息（用于生成配置，从首次使用时开始计算）
// days: 有效天数（从首次使用时开始计算）
func GenerateCreditInfoByDays(points int64, days int64) (*CreditInfo, error) {
	if days <= 0 {
		return nil, errors.New("有效天数必须大于 0")
	}

	// 将 StartAt 设置为一个未来的时间（9999-12-31），表示未激活
	// 第一次使用时，ConsumeCreditWithBaseDir 会检测到 now < info.StartAt
	// 然后从当前时间开始计算，保持原有时长
	futureTime := time.Date(9999, 12, 31, 0, 0, 0, 0, time.Local)
	startAt := futureTime.Unix()

	// 计算结束时间：StartAt + 天数（秒数）
	// 第一次使用时，会从当前时间开始，加上这个天数
	durationSeconds := days * 24 * 60 * 60
	endAt := startAt + durationSeconds

	// 读取版本号（失败时返回 v1）
	version := GetCurrentVersion()

	return &CreditInfo{
		Version: version, // 添加版本号
		Points:  points,
		StartAt: startAt,
		EndAt:   endAt,
	}, nil
}

// GenerateCreditInfo 生成新的积分信息（用于生成配置，兼容旧版本）
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

	// 读取版本号（失败时返回 v1）
	version := GetCurrentVersion()

	return &CreditInfo{
		Version: version, // 添加版本号
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
