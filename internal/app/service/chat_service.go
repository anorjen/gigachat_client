package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"gigachat_client/internal/app/config"
	"gigachat_client/internal/app/dto"
	"gigachat_client/internal/app/logger"
	"io"
	"net/http"
	"os"
	"strings"
)

type ChatConfig struct {
	Model             string  `yaml:"model" default:"GigaChat:latest"`
	Temperature       float64 `yaml:"temperature" default:"0.87"`
	N                 int64   `yaml:"n" default:"1"`
	MaxTokens         int64   `yaml:"max_tokens" default:"512"`
	RepetitionPenalty float64 `yaml:"repetition_penalty" default:"1.07"`
	SaveHistory       bool    `yaml:"save_history" default:"true"`
	HistoryFilePath   string  `yaml:"history_file_path" default:"history.json"`
}

type ChatService struct {
	logger.Logger
	chatConfig     ChatConfig
	authService    *AuthService
	messageHistory []dto.Message
}

func NewChatService(authService *AuthService, chatConfig ChatConfig) *ChatService {
	return &ChatService{
		authService: authService,
		chatConfig:  chatConfig,
	}
}

func (s *ChatService) SendMessage(input string) (string, error) {
	payload, err := s.preparePayload(input)
	if err != nil {
		return "", err
	}

	body, err := s.sendRequest(payload)
	if err != nil {
		s.LogError("Request error", err)
		return "", err
	}

	answer, err := s.processResponse(body)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func (s *ChatService) getMessages(input string) []dto.Message {
	var messages []dto.Message
	if config.Flags.IsContinue && s.chatConfig.SaveHistory {
		messages = s.getMessageHistory()
	}
	if messages == nil {
		messages = make([]dto.Message, 0, 1)
	}

	messages = append(messages, dto.Message{Role: "user", Content: input})

	s.messageHistory = messages
	return messages
}

func (s *ChatService) preparePayload(input string) (string, error) {
	messages := s.getMessages(input)

	payload := dto.Request{
		Model:             s.chatConfig.Model,
		Messages:          messages,
		Temperature:       s.chatConfig.Temperature,
		N:                 s.chatConfig.N,
		MaxTokens:         s.chatConfig.MaxTokens,
		RepetitionPenalty: s.chatConfig.RepetitionPenalty,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		s.LogError("Marshal error:", err)
		return "", err
	}

	return string(jsonPayload), nil
}

func (s *ChatService) processResponse(body []byte) (string, error) {
	var answer = dto.Response{}
	err := json.Unmarshal(body, &answer)
	if err != nil {
		s.LogError("Unmarshal error: ", err)
		return "", err
	}

	if s.chatConfig.SaveHistory {
		s.saveMessages(answer.Choices[len(answer.Choices)-1].Message)
	}

	return answer.Choices[len(answer.Choices)-1].Message.Content, nil
}

func (s *ChatService) saveMessages(outputMessage dto.Message) {
	s.messageHistory = append(s.messageHistory, outputMessage)

	f, err := os.OpenFile(s.chatConfig.HistoryFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		s.LogError("Error open message history file: ", err)
		return
	}
	defer f.Close()

	history, err := json.Marshal(s.messageHistory)
	if err != nil {
		s.LogError("History marshal error", err)
		return
	}

	_, err = io.WriteString(f, string(history))
	if err != nil {
		s.LogError("Save history file error", err)
		return
	}
	s.LogInfo("History file saved")
	s.LogDebug(string(history))
}

func (s *ChatService) getMessageHistory() []dto.Message {
	f, err := os.OpenFile(s.chatConfig.HistoryFilePath, os.O_RDONLY, 0666)
	if err != nil {
		s.LogError("Error open message history file: ", err)
		return nil
	}
	defer f.Close()

	messageHistory, err := io.ReadAll(f)
	if err != nil {
		s.LogError("Read message history file error", err)
		return nil
	}

	res := make([]dto.Message, 0)
	err = json.Unmarshal(messageHistory, &res)
	if err != nil {
		s.LogError("Unmarshal message history error", err)
		return nil
	}

	s.LogDebug(fmt.Sprintf("%v", res))
	return res
}

func (s *ChatService) sendRequest(payload string) ([]byte, error) {
	url := "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader(payload))
	if err != nil {
		s.LogError("Create request error", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	token, err := s.authService.GetToken()
	if err != nil {
		s.LogError("Get token error", err)
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+token)

	res, err := client.Do(req)
	if err != nil {
		s.LogError("Request error", err)
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		s.LogError("Read body error", err)
		return nil, err
	}
	s.LogDebug("BODY: " + string(body))

	if res.StatusCode != 200 {
		s.LogError(fmt.Sprintf("Request error: Status %d", res.StatusCode), nil)
		return nil, errors.New("request status not 200")
	}

	return body, nil
}
