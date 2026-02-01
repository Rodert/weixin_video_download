package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"wx_channel/pkg/credit"
)

var (
	creditPoints int64
	creditDays   int64 // 有效天数（从首次使用时开始计算）
)

var generateCreditCmd = &cobra.Command{
	Use:   "generate-credit",
	Short: "生成积分配置",
	Long:  "生成加密的积分配置，包含积分数量和有效期（从首次使用时开始计算）",
	Run: func(cmd *cobra.Command, args []string) {
		if creditPoints <= 0 {
			fmt.Println("错误: 积分数量必须大于 0")
			return
		}
		if creditDays <= 0 {
			fmt.Println("错误: 有效天数必须大于 0 (--days)")
			return
		}

		// 生成积分信息（从首次使用时开始计算）
		info, err := credit.GenerateCreditInfoByDays(creditPoints, creditDays)
		if err != nil {
			fmt.Printf("生成失败: %v\n", err)
			return
		}

		// 加密
		encrypted, err := credit.EncryptCreditInfo(info)
		if err != nil {
			fmt.Printf("加密失败: %v\n", err)
			return
		}

		// 输出结果
		fmt.Println("=" + strings.Repeat("=", 60) + "=")
		fmt.Println("积分配置生成成功！")
		fmt.Println("=" + strings.Repeat("=", 60) + "=")
		fmt.Println()
		fmt.Println("请创建 credit.txt 文件（与可执行文件同目录），内容如下：")
		fmt.Println()
		fmt.Println("encrypted=" + encrypted)
		fmt.Println()
		fmt.Println("或者直接创建文件：")
		fmt.Printf("echo encrypted=%s > credit.txt\n", encrypted)
		fmt.Println()
		fmt.Println("积分信息：")
		fmt.Printf("  积分数量: %d\n", info.Points)
		fmt.Printf("  有效天数: %d 天（从首次使用时开始计算）\n", creditDays)
		fmt.Printf("  版本号: %s\n", info.Version)
		fmt.Printf("  激活状态: 未激活（首次使用时自动激活）\n")
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 60) + "=")
	},
}

func init() {
	generateCreditCmd.Flags().Int64Var(&creditPoints, "points", 1000, "积分数量")
	generateCreditCmd.Flags().Int64Var(&creditDays, "days", 0, "有效天数（从首次使用时开始计算）")
	Register(generateCreditCmd)
}
