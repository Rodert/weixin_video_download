package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"wx_channel/config"
	"wx_channel/internal/download"
	"wx_channel/internal/interceptor"
	"wx_channel/internal/manager"
)

var (
	Version        string
	device         string
	hostname       string
	port           int
	debug          bool
	cert_files     *interceptor.ServerCertFiles
	channel_files  *interceptor.ChannelInjectedFiles
	cert_file_name string
	cfg            *config.Config
	isDevMode      bool // 是否是开发模式
)

var root_cmd = &cobra.Command{
	Use:   "wx_video_download",
	Short: "启动下载程序",
	Long:  "\n启动后将对网络请求进行代理，在微信视频号详情页面注入下载按钮",
	PreRun: func(cmd *cobra.Command, args []string) {
	},
	Run: func(cmd *cobra.Command, args []string) {
		cfg.Debug = viper.GetBool("debug")
		root_command(RootCommandArg{
			InterceptorConfig: interceptor.InterceptorConfig{
				Version:        Version,
				SetSystemProxy: viper.GetBool("proxy.system"),
				Device:         device,
				Hostname:       viper.GetString("proxy.hostname"),
				Port:           viper.GetInt("proxy.port"),
				Debug:          cfg.Debug,
				CertFiles:      cert_files,
				CertFileName:   cert_file_name,
				ChannelFiles:   channel_files,
				Cfg:            cfg,
				IsDevMode:      isDevMode,
			},
		})
	},
}

func init() {
	root_cmd.PersistentFlags().StringVar(&device, "dev", "", "代理服务器网络设备")
	root_cmd.PersistentFlags().StringVar(&hostname, "hostname", "127.0.0.1", "代理服务器主机名")
	root_cmd.PersistentFlags().IntVar(&port, "port", 2023, "代理服务器端口")
	root_cmd.PersistentFlags().BoolVar(&debug, "debug", false, "是否开启调试")

	viper.BindPFlag("proxy.port", root_cmd.PersistentFlags().Lookup("port"))
	viper.BindPFlag("proxy.hostname", root_cmd.PersistentFlags().Lookup("hostname"))
	viper.BindPFlag("debug", root_cmd.PersistentFlags().Lookup("debug"))
}

func Execute(app_ver string, cert_filename string, files1 *interceptor.ChannelInjectedFiles, files2 *interceptor.ServerCertFiles, c *config.Config, devMode bool) error {
	cobra.MousetrapHelpText = ""

	Version = app_ver
	cert_file_name = cert_filename
	channel_files = files1
	cert_files = files2
	cfg = c
	isDevMode = devMode

	return root_cmd.Execute()
}
func Register(cmd *cobra.Command) {
	root_cmd.AddCommand(cmd)
}

type RootCommandArg struct {
	interceptor.InterceptorConfig
}

func root_command(args RootCommandArg) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signal_chan := make(chan os.Signal, 1)
	err_chan := make(chan error, 1)

	signal.Notify(signal_chan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(signal_chan)

	if isDevMode {
		fmt.Printf("\nv%v\n", Version)
		if args.Cfg.FilePath != "" {
			fmt.Printf("配置文件 %s\n", args.Cfg.FilePath)
		}
	}

	mgr := manager.NewServerManager()

	// 初始化拦截服务
	interceptorServer, err := interceptor.NewInterceptorServer(args.InterceptorConfig)
	if err != nil {
		fmt.Printf("ERROR 初始化代理服务失败: %v\n", err.Error())
		os.Exit(1)
	}
	mgr.RegisterServer(interceptorServer)

	// 初始化下载服务
	downloadServer := download.NewDownloadServer(cfg.DownloadLocalServerAddr)
	mgr.RegisterServer(downloadServer)

	cleanup := func() {
		fmt.Printf("\n正在关闭服务...\n")
		if err := mgr.StopServer("interceptor"); err != nil {
			fmt.Printf("⚠️ 关闭代理服务失败: %v\n", err)
		}
		if err := mgr.StopServer("download"); err != nil {
			fmt.Printf("⚠️ 关闭下载服务失败: %v\n", err)
		}
	}

	if args.Cfg.DownloadLocalServerEnabled {
		// 启动下载服务
		if err := mgr.StartServer("download"); err != nil {
			fmt.Printf("ERROR 启动下载服务失败: %v\n", err.Error())
			cleanup()
			os.Exit(1)
		}
		if isDevMode {
			color.Green(fmt.Sprintf("下载服务启动成功，地址: %s", args.Cfg.DownloadLocalServerAddr))
		} else {
			color.Green("下载服务启动成功")
		}
	}
	// 启动代理服务
	if err := mgr.StartServer("interceptor"); err != nil {
		fmt.Printf("ERROR 启动代理服务失败: %v\n", err.Error())
		cleanup()
		os.Exit(1)
	}
	if isDevMode {
		proxyAddr := fmt.Sprintf("%s:%d", args.Hostname, args.Port)
		color.Green(fmt.Sprintf("代理服务启动成功，地址: %s", proxyAddr))
	} else {
		color.Green("代理服务启动成功")
	}

	if isDevMode {
		proxyAddr := fmt.Sprintf("%s:%d", args.Hostname, args.Port)
		if !args.SetSystemProxy {
			color.Red(fmt.Sprintf("当前未设置系统代理,请通过软件将流量转发至 %s", proxyAddr))
			color.Red("设置成功后再打开视频号页面下载")
		} else {
			color.Green(fmt.Sprintf("已修改系统代理为 %s", proxyAddr))
			color.Green("请打开需要下载的视频号页面进行下载")
		}
	}
	fmt.Println("\n按 Ctrl+C 退出...")

	select {
	case <-signal_chan:
		cleanup()
	case err := <-err_chan:
		fmt.Printf("ERROR %v\n", err.Error())
		cleanup()
		os.Exit(1)
	case <-ctx.Done():
		cleanup()
	}
}
