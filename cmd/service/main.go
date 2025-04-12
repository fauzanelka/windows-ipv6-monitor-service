package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sys/windows/svc"

	"windows-ipv6-monitor-service/pkg/ipv6"
	"windows-ipv6-monitor-service/pkg/telegram"
)

var (
	logger        *zap.Logger
	logLevel      string
	checkInterval int
	botToken      string
	chatID        string
	logFile       string
)

type ipv6Service struct {
	logger        *zap.Logger
	checkInterval time.Duration
	botToken      string
	chatID        string
	stopChan      chan bool
}

func initLogger(level string) *zap.Logger {
	// Parse log level
	var zapLevel zapcore.Level
	err := zapLevel.UnmarshalText([]byte(level))
	if err != nil {
		fmt.Printf("Invalid log level %q: %v, defaulting to info\n", level, err)
		zapLevel = zapcore.InfoLevel
	}

	// Create encoder configuration
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create JSON encoder
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// Create console encoder for stdout
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	// Open the log file
	logPath := logFile
	if !filepath.IsAbs(logPath) {
		wd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Failed to get working directory: %v\n", err)
			os.Exit(1)
		}
		logPath = filepath.Join(wd, logPath)
	}

	// Ensure log directory exists
	logDir := filepath.Dir(logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Failed to create log directory: %v\n", err)
		os.Exit(1)
	}

	// Open the log file with append mode
	file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}

	// Create core with both file and stdout output
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), zapLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapLevel),
	)

	// Create logger
	logger := zap.New(core)

	logger.Info("Logger initialized",
		zap.String("level", level),
		zap.String("file", logPath))

	return logger
}

func (s *ipv6Service) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown

	changes <- svc.Status{State: svc.StartPending}
	s.logger.Info("Starting service")

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	s.stopChan = make(chan bool)

	// Start the monitoring routine
	go s.monitorIPv6()

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				s.logger.Info("Received shutdown command")
				close(s.stopChan)
				break loop
			default:
				s.logger.Error("Unexpected control request", zap.Int("command", int(c.Cmd)))
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return false, 0
}

func (s *ipv6Service) monitorIPv6() {
	ticker := time.NewTicker(s.checkInterval)
	defer ticker.Stop()

	notifier := telegram.NewNotifier(s.botToken, s.chatID)
	var lastIPv6 string

	// Initial check
	if currentIPv6, err := ipv6.GetGlobalIPv6(); err == nil {
		lastIPv6 = currentIPv6
		s.logger.Info("Initial IPv6 address detected", zap.String("ipv6", currentIPv6))
		err = notifier.SendMessage(fmt.Sprintf("🚀 Service started\nCurrent IPv6: <code>%s</code>", currentIPv6))
		if err != nil {
			s.logger.Error("Failed to send initial notification", zap.Error(err))
		}
	} else {
		s.logger.Error("Failed to get initial IPv6 address", zap.Error(err))
	}

	for {
		select {
		case <-ticker.C:
			currentIPv6, err := ipv6.GetGlobalIPv6()
			if err != nil {
				s.logger.Error("Failed to get IPv6 address", zap.Error(err))
				continue
			}

			s.logger.Debug("IPv6 check completed", zap.String("ipv6", currentIPv6))

			if lastIPv6 != "" && currentIPv6 != lastIPv6 {
				s.logger.Info("IPv6 address changed",
					zap.String("old", lastIPv6),
					zap.String("new", currentIPv6))

				message := fmt.Sprintf("📢 IPv6 address changed\nOld: <code>%s</code>\nNew: <code>%s</code>",
					lastIPv6, currentIPv6)

				err = notifier.SendMessage(message)
				if err != nil {
					s.logger.Error("Failed to send notification", zap.Error(err))
				}
			}

			lastIPv6 = currentIPv6

		case <-s.stopChan:
			s.logger.Info("Stopping IPv6 monitoring")
			message := "🛑 Service stopped"
			err := notifier.SendMessage(message)
			if err != nil {
				s.logger.Error("Failed to send stop notification", zap.Error(err))
			}
			return
		}
	}
}

func main() {
	rootCmd := &cobra.Command{
		Use:   "ipv6-monitor",
		Short: "IPv6 address monitoring service",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize logger
			logger = initLogger(logLevel)
			defer logger.Sync()

			// Validate required credentials
			if botToken == "" || chatID == "" {
				logger.Fatal("Telegram bot token and chat ID are required")
			}

			// Create service instance
			svc := &ipv6Service{
				logger:        logger,
				checkInterval: time.Duration(checkInterval) * time.Minute,
				botToken:      botToken,
				chatID:        chatID,
			}

			// Determine if we're running as a service
			isService, err := svc.IsWindowsService()
			if err != nil {
				logger.Fatal("Failed to determine if running as service", zap.Error(err))
			}

			if isService {
				err = svc.Run(svc)
				if err != nil {
					logger.Fatal("Service failed", zap.Error(err))
				}
			} else {
				logger.Info("Running in console mode")
				go svc.monitorIPv6()
				// Keep the application running
				select {}
			}
		},
	}

	// Define flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Logging level (debug, info, warn, error)")
	rootCmd.PersistentFlags().IntVar(&checkInterval, "check-interval", 5, "IPv6 check interval in minutes")
	rootCmd.PersistentFlags().StringVar(&botToken, "bot-token", "", "Telegram bot token")
	rootCmd.PersistentFlags().StringVar(&chatID, "chat-id", "", "Telegram chat ID")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "ipv6-monitor.log", "Log file path (relative to working directory or absolute)")

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (s *ipv6Service) IsWindowsService() (bool, error) {
	return svc.IsWindowsService()
}

func (s *ipv6Service) Run(service interface{}) error {
	return svc.Run("IPv6MonitorService", service.(svc.Handler))
}
