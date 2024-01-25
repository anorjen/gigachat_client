package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"gigachat_client/internal/app/logger"
	"github.com/google/uuid"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresAt   int64  `json:"expires_at"`
}

type ClientConfig struct {
	ClientId      string `yaml:"client_id"`
	ClientSecret  string `yaml:"client_secret"`
	SaveToken     bool   `yaml:"save_token" default:"true"`
	TokenFilePath string `yaml:"token_file_path" default:"token.json"`
}

type AuthService struct {
	logger.Logger
	clientConfig ClientConfig
	auth         string
	token        Token
}

func NewAuthService(clientConfig ClientConfig) *AuthService {
	return &AuthService{
		clientConfig: clientConfig,
		auth:         base64.StdEncoding.EncodeToString([]byte(clientConfig.ClientId + ":" + clientConfig.ClientSecret)),
	}
}

func (s *AuthService) requestNewToken() (Token, error) {
	url := "https://ngw.devices.sberbank.ru:9443/api/v2/oauth"
	method := "POST"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, strings.NewReader("scope=GIGACHAT_API_PERS"))
	if err != nil {
		s.LogError("Create request error", err)
		return Token{}, nil
	}

	req.Header.Add("RqUID", uuid.New().String())
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Authorization", "Basic "+s.auth)

	res, err := client.Do(req)
	if err != nil {
		s.LogError("Auth request error: ", err)
		return Token{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		s.LogError("Read body error", err)
		return Token{}, err
	}
	s.LogDebug("Auth BODY: " + string(body))

	if res.StatusCode != 200 {
		s.LogError(fmt.Sprintf("Auth request error: Status %d", res.StatusCode), nil)
		return Token{}, errors.New("auth request status not 200")
	}

	var token = Token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		s.LogError("Unmarshal error: ", err)
		return Token{}, err
	}

	return token, nil
}

func (s *AuthService) GetToken() (string, error) {
	s.readSavedToken()
	if s.token != (Token{}) && time.Now().UnixMilli() < s.token.ExpiresAt {
		s.LogInfo("Token not expired")
		return s.token.AccessToken, nil
	}

	token, err := s.requestNewToken()
	if err != nil {
		return "", err
	}
	s.token = token

	s.saveToken()
	return token.AccessToken, nil
}

func (s *AuthService) saveToken() {
	if !s.clientConfig.SaveToken {
		return
	}

	f, err := os.OpenFile(s.clientConfig.TokenFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		s.LogError("Error open token file: ", err)
		return
	}
	defer f.Close()

	token, err := json.Marshal(s.token)
	if err != nil {
		s.LogError("Token marshal error", err)
		return
	}

	_, err = io.WriteString(f, string(token))
	if err != nil {
		s.LogError("Save token file error", err)
		return
	}

	s.LogInfo("token file saved")
	s.LogDebug(string(token))
}

func (s *AuthService) readSavedToken() {
	if !s.clientConfig.SaveToken {
		return
	}

	f, err := os.OpenFile(s.clientConfig.TokenFilePath, os.O_RDONLY, 0666)
	if err != nil {
		s.LogError("Error open token file: ", err)
		return
	}
	defer f.Close()

	token, err := io.ReadAll(f)
	if err != nil {
		s.LogError("Read saved token error", err)
		return
	}

	res := Token{}
	err = json.Unmarshal(token, &res)
	if err != nil {
		s.LogError("Unmarshal token error", err)
		return
	}

	s.LogDebug(fmt.Sprintf("%v", res))
	s.token = res
}
