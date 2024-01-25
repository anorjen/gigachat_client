package app

import (
	"errors"
	"fmt"
	"gigachat_client/internal/app/config"
	"gigachat_client/internal/app/logger"
	"gigachat_client/internal/app/service"
	"os"
)

var (
	log       logger.Logger
	appConfig *config.AppConfig

	authService *service.AuthService
	chatService *service.ChatService
)

func Main() {
	var err error

	input, err := readArgs(os.Args)
	if err != nil {
		usage()
		return
	}

	appConfig = &config.AppConfig{}
	appConfig.GetConfig()

	log.SetLogLevel(appConfig.Log.Level)
	log.SetFile(appConfig.Log.File)
	defer log.Close()

	authService = service.NewAuthService(appConfig.Client)
	chatService = service.NewChatService(authService, appConfig.Chat)

	message, err := chatService.SendMessage(input)
	if err != nil {
		log.LogError("Response error", err)
		sendResponse(err.Error())
		return
	}
	sendResponse(message)
}

func readArgs(args []string) (string, error) {
	if len(args) <= 1 {
		return "", errors.New("empty arguments")
	}

	var inputIndex = 1
	if rune(args[1][0]) == '-' {
		switch args[1] {
		case "-c", "--continue":
			config.Flags.IsContinue = true
			inputIndex++
		case "-h", "--help":
			return "", errors.New("print usage")
		default:
			return "", errors.New("wrong flag")
		}
	}

	return args[inputIndex], nil
}

func sendResponse(message string) {
	fmt.Printf("GigaChat>\t%s", message)
}

func usage() {
	fmt.Printf("\t%s\n\n\tFlags:\n\t%s\n\t%s\n", "gigachat_client [FLAG] [QUESTION]", "-h --help :this info", "-c --continue :continue with message history")
}
