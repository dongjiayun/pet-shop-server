package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
	LogDir      string
)

// InitLogger 初始化日志模块
func InitLogger() {
	// 设置日志目录
	LogDir = "./log"
	
	// 确保日志目录存在
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}
	
	// 创建不同级别的日志记录器
	InfoLogger = createLogger("info")
	ErrorLogger = createLogger("error")
	DebugLogger = createLogger("debug")
	
	// 启动日志清理定时任务
	go startLogCleaner()
}

// createLogger 创建特定级别的日志记录器
func createLogger(level string) *log.Logger {
	// 获取当前日期和小时
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	hourStr := now.Format("15")
	
	// 创建日期和小时的目录结构
	dirPath := filepath.Join(LogDir, dateStr, hourStr)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}
	
	// 创建日志文件
	fileName := fmt.Sprintf("%s_%s.log", level, now.Format("20060102_150405"))
	filePath := filepath.Join(dirPath, fileName)
	
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("创建日志文件失败: %v", err)
	}
	
	// 创建日志记录器，同时输出到控制台和文件
	return log.New(
		io.MultiWriter(os.Stdout, file),
		fmt.Sprintf("%s %s: ", time.Now().Format("2006-01-02 15:04:05"), level),
		log.LstdFlags|log.Lshortfile,
	)
}

// Info 记录信息级别的日志
func Info(format string, v ...interface{}) {
	if InfoLogger != nil {
		InfoLogger.Printf(format, v...)
	}
}

// Error 记录错误级别的日志
func Error(format string, v ...interface{}) {
	if ErrorLogger != nil {
		ErrorLogger.Printf(format, v...)
	}
}

// Debug 记录调试级别的日志
func Debug(format string, v ...interface{}) {
	if DebugLogger != nil {
		DebugLogger.Printf(format, v...)
	}
}

// startLogCleaner 启动日志清理定时任务
func startLogCleaner() {
	// 每天凌晨执行一次日志清理
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	// 立即执行一次清理
	cleanOldLogs()
	
	for range ticker.C {
		cleanOldLogs()
	}
}

// cleanOldLogs 清理3个月以上的日志文件
func cleanOldLogs() {
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	
	// 遍历日志目录
	err := filepath.Walk(LogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 只处理文件，不处理目录
		if !info.IsDir() {
			// 检查文件修改时间是否超过3个月
			if info.ModTime().Before(threeMonthsAgo) {
				// 删除旧日志文件
				if err := os.Remove(path); err != nil {
					Error("删除旧日志文件失败: %v, 文件路径: %s", err, path)
				} else {
					Info("删除旧日志文件: %s", path)
				}
			}
		}
		
		return nil
	})
	
	if err != nil {
		Error("清理旧日志文件失败: %v", err)
	}
}
