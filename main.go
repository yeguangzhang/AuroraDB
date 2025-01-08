package main

import (
	"embed"
	"log"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	// 设置日志文件
	setupLogging()

	// 创建应用实例
	app := NewApp()

	// 运行应用
	err := wails.Run(&options.App{
		Title:              "AuroraDB",
		Width:              1280,
		Height:             700,
		Assets:             assets,
		BackgroundColour:   &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		LogLevel:           logger.DEBUG,
		LogLevelProduction: logger.ERROR,
		OnStartup:          app.startup,
		Bind: []interface{}{
			app,
		},
		// Mac 特定配置
		Mac: &mac.Options{
			TitleBar:   mac.TitleBarHiddenInset(),
			Appearance: mac.NSAppearanceNameDarkAqua,
			About: &mac.AboutInfo{
				Title:   "AuroraDB",
				Message: "© 2024 Database Tool",
			},
		},
		HideWindowOnClose: true,
	})

	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}

func setupLogging() string {
	// 获取用户的home目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}

	// 创建日志目录
	logDir := filepath.Join(homeDir, ".auroradb", "logs")
	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatal(err)
	}

	// 设置日志文件
	logFile := filepath.Join(logDir, "auroradb.log")
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}

	// 设置日志输出到文件
	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Application started, logging to %s", logFile)
	return logFile
}
