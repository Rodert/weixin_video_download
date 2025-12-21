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
	Run: func(cmd *cobra.Command, args []string) {
		if batchPoints <= 0 || batchDays <= 0 || batchCount <= 0 {
			fmt.Println("参数错误：points / days / count 都必须大于 0")
			return
		}
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
			fmt.Println(enc)
		}
	},
}

func init() {
	batchGenerateCmd.Flags().Int64Var(&batchPoints, "points", 10, "单组积分数量")
	batchGenerateCmd.Flags().Int64Var(&batchDays, "days", 7, "单组有效天数")
	batchGenerateCmd.Flags().IntVar(&batchCount, "count", 100, "生成组数")
	Register(batchGenerateCmd)
}