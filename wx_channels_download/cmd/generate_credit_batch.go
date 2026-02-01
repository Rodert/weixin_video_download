package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"wx_channel/pkg/credit"
)

var (
	batchPoints int64
	batchDays   int64
	batchCount  int
)

var batchGenerateCmd = &cobra.Command{
	Use:   "batch-generate",
	Short: "批量生成积分配置（一行一个）",
	Long:  "批量生成加密的积分配置，每行输出一个加密字符串。所有密钥自动包含当前版本号，确保版本兼容性。",
	Run: func(cmd *cobra.Command, args []string) {
		if batchPoints <= 0 || batchDays <= 0 || batchCount <= 0 {
			fmt.Println("参数错误：points / days / count 都必须大于 0")
			return
		}

		// 获取当前版本号（用于显示）
		currentVersion := credit.GetCurrentVersion()

		// 输出批量生成信息（到 stderr，避免影响密钥输出）
		fmt.Fprintf(os.Stderr, "批量生成积分配置\n")
		fmt.Fprintf(os.Stderr, "版本号: %s\n", currentVersion)
		fmt.Fprintf(os.Stderr, "单组积分: %d\n", batchPoints)
		fmt.Fprintf(os.Stderr, "单组有效天数: %d 天（从首次使用时开始计算）\n", batchDays)
		fmt.Fprintf(os.Stderr, "生成数量: %d\n", batchCount)
		fmt.Fprintf(os.Stderr, "----------------------------------------\n")
		fmt.Fprintf(os.Stderr, "开始生成（以下为密钥列表，每行一个）:\n\n")

		// 生成密钥（输出到 stdout，便于重定向到文件）
		for i := 0; i < batchCount; i++ {
			info, err := credit.GenerateCreditInfoByDays(batchPoints, batchDays)
			if err != nil {
				fmt.Fprintf(os.Stderr, "生成失败: %v\n", err)
				return
			}
			enc, err := credit.EncryptCreditInfo(info)
			if err != nil {
				fmt.Fprintf(os.Stderr, "加密失败: %v\n", err)
				return
			}
			// 输出到 stdout（便于重定向）
			fmt.Println(enc)
		}

		// 输出完成信息（到 stderr）
		fmt.Fprintf(os.Stderr, "\n----------------------------------------\n")
		fmt.Fprintf(os.Stderr, "批量生成完成！共生成 %d 个密钥\n", batchCount)
		fmt.Fprintf(os.Stderr, "所有密钥版本: %s\n", currentVersion)
		fmt.Fprintf(os.Stderr, "\n使用示例：\n")
		fmt.Fprintf(os.Stderr, "  # 生成密钥并保存到文件\n")
		fmt.Fprintf(os.Stderr, "  go run . batch-generate --points 10 --days 7 --count 100 > keys.txt\n")
		fmt.Fprintf(os.Stderr, "  \n")
		fmt.Fprintf(os.Stderr, "  # 生成密钥并保存到文件（同时查看信息）\n")
		fmt.Fprintf(os.Stderr, "  go run . batch-generate --points 10 --days 7 --count 100 2>&1 | tee keys.txt\n")
	},
}

func init() {
	batchGenerateCmd.Flags().Int64Var(&batchPoints, "points", 10, "单组积分数量")
	batchGenerateCmd.Flags().Int64Var(&batchDays, "days", 7, "单组有效天数")
	batchGenerateCmd.Flags().IntVar(&batchCount, "count", 100, "生成组数")
	Register(batchGenerateCmd)
}