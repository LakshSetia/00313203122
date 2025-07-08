package loggingMiddleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
)
const AuthEndPoint = "http://20.244.56.144/evaluation-service/auth"

type AuthRequest struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	RollNo       string `json:"rollNo"`
	AccessCode   string `json:"accessCode"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

type Response struct {
	TokenType 	string `json:"token_type"`
	AccessToken string `json:"access_token"`
	ExpiresIN 	uint32 `json:"expires_in"`
}

type LogRequest struct {
	Stack   string `json:"stack"`
	Level   string `json:"level"`
	Package string `json:"package"`
	Message string `json:"message"`
}

const logEndPoint = "http://20.244.56.144/evaluation-service/logs"

var Stack = map[string]bool{
	"backend":  true,
	"frontend": true,
}
var Level = map[string]bool{
	"debug": true,
	"info":  true,
	"warn":  true,
	"error": true,
	"fatal": true,
}
var Package = map[string]bool{
	"cache":      true,
	"controller": true,
	"db":         true,
	"domain":     true,
	"handler":    true,
	"repository": true,
	"route":      true,
	"service":    true,
}

func GetAuthToken() (string, error) {
	request := AuthRequest{
		Email:        "lakshsetia30@gmail.com",
		Name:         "Laksh Setia",
		RollNo:       "00313203122",
		AccessCode:   "VPpsmT",
		ClientID:     "e3e5c180-46ca-412d-89d7-7c34c6b8b9cf",
		ClientSecret: "ZANXWyjqTjdGbZsf",
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", AuthEndPoint, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("Authentication failed status: " + resp.Status)
	}
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}
	return response.TokenType + " " + response.AccessToken, nil
}

func Log(stack, level, pkg, message string) error {
	stack = strings.ToLower(stack)
	level = strings.ToLower(level)
	pkg = strings.ToLower(pkg)
	if !Stack[stack] {
		return errors.New("invalid stack value")
	}
	if !Level[level] {
		return errors.New("invalid level value")
	}
	if !Package[pkg] {
		return errors.New("invalid package value")
	}
	request := LogRequest{
		Stack: stack,
		Level: level,
		Package: pkg,
		Message: message,
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	req, err := http.NewRequest("POST", logEndPoint, &buf)
	if err != nil {
		return err
	}
	authToken, err := GetAuthToken()
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return errors.New("Logging failed status: " + resp.Status)
	}
	return nil
}

func Middleware(stack, level, pkg string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		message := fmt.Sprintf("Request: %s %s", r.Method, r.URL.Path)
		if err := Log(stack, level, pkg, message); err != nil {
			log.Fatal("Logging Failed: ", err.Error())
		}
		next(w, r)
	}
}