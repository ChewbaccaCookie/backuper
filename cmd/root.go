package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/spf13/cobra"

	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "backuper",
	Short: "Diese Anwendung dient zur Backuperstellung und zum Upload des Backups auf AWS Server.",
	Long:  `Diese Anwendung dient zur Backuperstellung und zum Upload des Backups auf AWS Server.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		date := time.Now()
		config := loadConfig(date)

		printHeader(date)

		execDone := make(chan bool)
		var wg sync.WaitGroup
		wg.Add(len(config.Commands))

		for _, command := range config.Commands {
			execCommand(config, command, &wg)
		}

		go func() {
			wg.Wait()
			close(execDone)
		}()

		<-execDone

		fmt.Printf("Backup creation finished in %s.\n", time.Since(date))
		fmt.Println("Backup uploading started...")
		rcloneDir := fmt.Sprintf("%s:%s", config.ConnectionName, config.Bucket)
		serverPath := fmt.Sprintf("%s/%s/%s", rcloneDir, config.SubPath, time.Now().Format("2006/01/02"))
		fmt.Printf("Uploading %s -> %s\n", config.TmpPath, serverPath)

		c := exec.Command("bash", "-c", fmt.Sprintf("rclone move %s %s --delete-empty-src-dirs", config.TmpPath, serverPath))
		err := c.Run()

		if err != nil {
			fatal("Data uploading failed!", err)
		}
		printFooter(date)

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .backuper)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		path, err := os.Executable()
		cobra.CheckErr(err)

		// Search config in home directory with name ".backuper" (without extension).
		viper.AddConfigPath(filepath.Dir(path))
		viper.SetConfigType("yaml")
		viper.SetConfigName(".backuper")
	}

	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	if os.IsNotExist(err) {
		panic("A config file cannot be found!")
	}

}
