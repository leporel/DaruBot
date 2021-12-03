package cmd

import (
	"DaruBot/internal/config"
	"DaruBot/pkg/logger"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
)

var (
	logo = `$$$$$$$\                                      $$$$$$$\             $$\     
$$  __$$\                                     $$  __$$\            $$ |    
$$ |  $$ | $$$$$$\   $$$$$$\  $$\   $$\       $$ |  $$ | $$$$$$\ $$$$$$\   
$$ |  $$ | \____$$\ $$  __$$\ $$ |  $$ |      $$$$$$$\ |$$  __$$\\_$$  _|  
$$ |  $$ | $$$$$$$ |$$ |  \__|$$ |  $$ |      $$  __$$\ $$ /  $$ | $$ |    
$$ |  $$ |$$  __$$ |$$ |      $$ |  $$ |      $$ |  $$ |$$ |  $$ | $$ |$$\ 
$$$$$$$  |\$$$$$$$ |$$ |      \$$$$$$  |      $$$$$$$  |\$$$$$$  | \$$$$  |
\_______/  \_______|\__|       \______/       \_______/  \______/   \____/ `

	DebugMode  = false
	Ver        = "unknown"
	BuildDate  = "unknown"
	GitCommit  = "unknown"
	ConfigFile = ""

	rootCmd = &cobra.Command{}
)

func init() {
	rootCmd = &cobra.Command{
		Use:     "",
		Short:   "",                                   // TODO
		Long:    `https://github.com/leporel/DaruBot`, // TODO
		Version: Ver,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("%s \n\n", logo)
			fmt.Printf("\t Version: %s, Date: %s, Git: %s\n\n", Ver, BuildDate, GitCommit)
			fmt.Printf("\t To close program correctly, use Ctrl+C\n\n\n")

			cfg := initConfig()

			logLevel := logger.InfoLevel
			if cfg.IsDebug() {
				logLevel = logger.DebugLevel
			}

			log := logger.New(os.Stdout, logLevel)
			log.Debug("DebugMode enabled")

			sigs := make(chan os.Signal, 1)
			done := make(chan bool, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

			go func() {
				sig := <-sigs
				log.Infof("SIG: %v, shutdown...", sig)
				done <- true
			}()

			rootCtx := context.Background()
			_, cancelFn := context.WithCancel(rootCtx)
			//ctx, cancelFn := context.WithCancel(rootCtx)

			//core.Run(ctx)

			<-done
			cancelFn()

			//core.Shutdown()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&ConfigFile, "config", "c", "", "config file (e.g. ./config.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&DebugMode, "debug", "d", false, "debug mode")
}

func Run() {
	// TODO CMD list strategies

	// TODO CMD test strategies

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func initConfig() config.Configurations {
	viper.SetConfigName("default_config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
	}

	cfg := config.GetDefaultConfig()

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		panic(err)
	}
	cfg.SetDebug(DebugMode)

	// TODO viper.UnmarshalKey() dynamic strategy
	// custom_name - strategy and params

	return cfg
}
