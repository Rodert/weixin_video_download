package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"wx_channel/pkg/credit"
)

var (
	creditPoints    int64
	creditStartDate string
	creditEndDate   string
)

var generateCreditCmd = &cobra.Command{
	Use:   "generate-credit",
	Short: "生成积分配置",
	Long:  "生成加密的积分配置，包含积分数量和有效期",
	Run: func(cmd *cobra.Command, args []string) {
		if creditPoints <= 0 {
			fmt.Println("错误: 积分数量必须大于 0")
			return
		}
		if creditStartDate == "" {
			fmt.Println("错误: 必须指定开始日期 (--start-date)")
			return
		}
		if creditEndDate == "" {
			fmt.Println("错误: 必须指定结束日期 (--end-date)")
			return
		}

		// 生成积分信息
		info, err := credit.GenerateCreditInfo(creditPoints, creditStartDate, creditEndDate)
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
		fmt.Println("请创建 credit.yaml 文件（与可执行文件同目录），内容如下：")
		fmt.Println()
		fmt.Println("encrypted: " + encrypted)
		fmt.Println()
		fmt.Println("或者直接创建文件：")
		fmt.Printf("echo encrypted: %s > credit.yaml\n", encrypted)
		fmt.Println()
		fmt.Println("积分信息：")
		fmt.Printf("  积分数量: %d\n", info.Points)
		fmt.Printf("  开始时间: %s (00:00:00)\n", time.Unix(info.StartAt, 0).Format("2006-01-02"))
		fmt.Printf("  结束时间: %s (23:59:59)\n", time.Unix(info.EndAt, 0).Format("2006-01-02"))
		fmt.Println()
		fmt.Println("=" + strings.Repeat("=", 60) + "=")
	},
}

func init() {
	generateCreditCmd.Flags().Int64Var(&creditPoints, "points", 1000, "积分数量")
	generateCreditCmd.Flags().StringVar(&creditStartDate, "start-date", "", "开始日期（格式: 2006.01.02 或 2006-01-02）")
	generateCreditCmd.Flags().StringVar(&creditEndDate, "end-date", "", "结束日期（格式: 2006.01.02 或 2006-01-02）")
	Register(generateCreditCmd)
}
