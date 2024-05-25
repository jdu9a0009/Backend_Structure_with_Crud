package web

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"project/internal/pkg/config"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Logger struct {
	Folder string
}

func NewLogger(folder string) *Logger {
	if folder == "" {
		folder = "logs"
	}
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Println("cannot create directory, reason:", err)
	}
	return &Logger{Folder: folder}
}

func (l *Logger) WriteLog(ctx *Context, data interface{}) error {
	var (
		file *os.File
		err  error
		id   string
	)

	userId, ok := ctx.Value("user_id").(int64)
	if !ok {
		id = "-"
	} else {
		id = fmt.Sprintf("%d", userId)
	}

	name := fmt.Sprintf("%s/%s.csv", l.Folder, time.Now().Format("02-01-2006"))

	file, err = os.OpenFile(name, os.O_RDWR|os.O_APPEND, 0755)
	if err != nil {
		file, err = os.Create(name)
		if err != nil {
			return err
		}

		headers := []string{"time", "url", "user_id", "request_method", "request_body", "request_useragent", "response"}

		writer := csv.NewWriter(file)
		if err = writer.Write(headers); err != nil {
			return err
		}

		writer.Flush()
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	body := ctx.Value("body")

	recordBuffer := []string{
		fmt.Sprintf("Time: %s", time.Now().Format("02-01-2006 15:04:05")),
		fmt.Sprintf("URL: %v", ctx.Request.URL),
		fmt.Sprintf("UserID: %s", id),
		fmt.Sprintf("Method: %s", ctx.Request.Method),
		fmt.Sprintf("Request.Body: %v", body),
		fmt.Sprintf("Useragent: %s", ctx.Request.UserAgent()),
		fmt.Sprintf("Response.Body: %v", data),
	}
	if err = writer.Write(recordBuffer); err != nil {
		return err
	}

	if err = l.SendBotMsg(recordBuffer); err != nil {
		log.Println(err)
	}

	return nil
}

func (l *Logger) SendBotMsg(recordBuffer []string) error {
	cfg := config.NewConfig()
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.ErrorBotToken)

	for _, v := range cfg.ErrorChatID {
		body, err := json.Marshal(map[string]interface{}{
			"chat_id": v,
			"text":    strings.Join(recordBuffer, "\n"),
		})
		if err != nil {
			return err
		}

		response, err := http.Post(url, "application/json", bytes.NewBuffer(body))
		if err != nil {
			return err
		}

		if response.StatusCode >= 300 && response.StatusCode < 200 {
			log.Println(response.StatusCode)
			return errors.New("status code was not okay")
		}
	}

	return nil
}
