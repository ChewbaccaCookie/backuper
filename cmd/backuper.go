package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/spf13/viper"
)

func fatal(message string, err error) {
	fmt.Printf("ERROR: %s\n", message)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(1)
}

func loadConfig(date time.Time) Config {
	var config Config
	viper.Unmarshal(&config)

	if len(os.Args) > 1 {
		config.SubPath = os.Args[1]
	}

	if config.Timeout == 0 {
		config.Timeout = 120
	}

	// Check tmpPath
	if config.TmpPath == "" {
		config.TmpPath = "/tmp/backup"
	}
	uuid, _ := uuid.NewV4()
	config.TmpPath = fmt.Sprintf("%s/%s", config.TmpPath, uuid)

	err := os.MkdirAll(config.TmpPath, os.ModePerm)
	if err != nil {
		fatal("Temporary directory cannot be opened!", err)
	}

	// Exchange vars
	for i, cmd := range config.Commands {
		cmd.Command = strings.ReplaceAll(cmd.Command, "##path##", config.TmpPath)
		cmd.Command = strings.ReplaceAll(cmd.Command, "##date##", date.Format("02_01_2006_15_04_05"))
		config.Commands[i] = cmd
	}

	return config

}
func printHeader(date time.Time) {
	println("-------------------------------------------------------------------")
	fmt.Printf("\n              BACKUP STARTED (%s)\n\n", date.Format("02.01.2006 15:04:05"))
	println("-------------------------------------------------------------------")
}
func printFooter(date time.Time) {
	println("-------------------------------------------------------------------")
	fmt.Printf("\n              BACKUP FINISHED (%s)\n\n", time.Since(date))
	println("-------------------------------------------------------------------")
}
func execCommand(config Config, command Command, wg *sync.WaitGroup) {
	startDate := time.Now()
	cmd := exec.Command("bash", "-c", command.Command)
	var buf bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &buf
	cmd.Stderr = &stderr
	cmd.Start()

	done := make(chan error)
	go func() { done <- cmd.Wait() }()
	timeout := time.After(time.Duration(config.Timeout) * time.Second)

	select {
	case <-timeout:
		// Timeout happened first, kill the process and print a message.
		cmd.Process.Kill()
		fmt.Printf("ERROR: Backup script '%s' execution terminated by timeout\n", command.Name)
		wg.Done()
	case err := <-done:
		if err != nil {
			fmt.Printf("ERROR: Backup script '%s' execution failed.\nCommand: %s\n", command.Name, command.Command)
			fmt.Println(err)
			fmt.Println(stderr.String())
		} else {
			fmt.Printf("SUCCESS: Backup script '%s' succeeded in %s.\n", command.Name, time.Since(startDate))
		}

		println("-------------------------------------------------------------------")

		wg.Done()
	}

}
