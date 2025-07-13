package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultTemperature    = 0.0
	defaultInputFile      = "messages.json"
	defaultLogsBaseDir    = "logs"
	defaultInputBaseDir   = "input"
	defaultPromptsBaseDir = "prompts"
)

type MsgRole string

const (
	USER      MsgRole = "user"
	ASSISTANT MsgRole = "assistant"
	SYSTEM    MsgRole = "system"
)

type MessageIn struct {
	Role    MsgRole `json:"role"`
	Content string  `json:"content"`
	File    string  `json:"file"`
}

type Message struct {
	Role    MsgRole `json:"role"`
	Content string  `json:"content"`
}

type RequestPayload struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
}

type ResponseChoice struct {
	Message Message `json:"message"`
}

type ResponseBody struct {
	Choices []ResponseChoice `json:"choices"`
	Usage   Usage            `json:"usage"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type Config struct {
	APIKey      string
	Model       string
	URL         string
	Temperature float64
	InputFile   string
	InputDir    string
	PromptsDir  string
	LogsDir     string
}

func saveConversationLog(messages []Message, model string, logsDir string) error {
	logDir := path.Join(logsDir, strings.Replace(model, "/", "_", -1))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	timestamp := time.Now().Format(time.RFC3339)
	fileName := path.Join(logDir, fmt.Sprintf("%s.log.json", timestamp))
	fileContent, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to JSON parse conversation content: %w", err)
	}

	if err := os.WriteFile(fileName, fileContent, 0644); err != nil {
		return fmt.Errorf("failed to save conversation log file: %w", err)
	}

	fmt.Printf("Conversation saved to %s\n", fileName)
	return nil
}

func readUserInput(reader *bufio.Reader) (string, error) {
	fmt.Print(">> ")
	userInput, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read user input: %w", err)
	}

	userInput = strings.TrimSuffix(userInput, "\n")
	userInput = strings.TrimSuffix(userInput, "\r\n")

	return userInput, nil
}

func displayInitScreen(messages []Message, model string, temperature float32) {
	systemMsgsCount := 0
	userMsgsCount := 0
	assistantMsgsCount := 0

	for _, msg := range messages {
		switch msg.Role {
		case USER:
			userMsgsCount++
		case ASSISTANT:
			assistantMsgsCount++
		case SYSTEM:
			systemMsgsCount++
		}
	}

	fmt.Printf(`
+--------------------------------------------------+
|                                                  |
|      You are now chatting with the model:        |
|                                                  |
+--------------------------------------------------+

> %s

+--------------------------------------------------+
| Temperature: %.2f                                |
|--------------------------------------------------|
| Context Messages Count:                          |
|                                                  |
|   System:    %3d                                 |
|   User:      %3d                                 |
|   Assistant: %3d                                 |
|                                                  |
|--------------------------------------------------|
| Commands:                                        |
|                                                  |
|   >> /quit     to save conversation and exit     |
|   >> /quit!    to exit without saving            |
|                                                  |
+--------------------------------------------------+

`, model, temperature, systemMsgsCount, userMsgsCount, assistantMsgsCount)
}

func loadConfig() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: could not load .env file: %v", err)
	}

	apiKey := flag.String("api-key", os.Getenv("LLM_PROVIDER_KEY"), "LLM provider API key")
	model := flag.String("model", os.Getenv("LLM_MODEL"), "LLM model name")
	url := flag.String("url", os.Getenv("CHAT_COMPLETION_URL"), "Chat completion URL")
	temperatureStr := flag.String("temperature", os.Getenv("TEMPERATURE"), "Temperature for the LLM")
	inputFile := flag.String("input", defaultInputFile, "Path to the input messages file")
	inputDir := flag.String("input-dir", defaultInputBaseDir, "Directory for input files")
	promptsDir := flag.String("prompts-dir", defaultPromptsBaseDir, "Directory for prompt files")
	logsDir := flag.String("logs-dir", defaultLogsBaseDir, "Directory for log files")

	flag.Parse()

	if *apiKey == "" {
		return nil, fmt.Errorf("missing LLM provider API key. Use --api-key flag or LLM_PROVIDER_KEY env var")
	}
	if *model == "" {
		return nil, fmt.Errorf("missing LLM model. Use --model flag or LLM_MODEL env var")
	}
	if *url == "" {
		return nil, fmt.Errorf("missing chat completion URL. Use --url flag or CHAT_COMPLETION_URL env var")
	}

	temperature, err := strconv.ParseFloat(*temperatureStr, 64)
	if err != nil {
		log.Printf("Warning: failed to parse temperature value \"%s\". Using default value instead: %v\n", *temperatureStr, defaultTemperature)
		temperature = defaultTemperature
	}

	return &Config{
		APIKey:      *apiKey,
		Model:       *model,
		URL:         *url,
		Temperature: temperature,
		InputFile:   *inputFile,
		InputDir:    *inputDir,
		PromptsDir:  *promptsDir,
		LogsDir:     *logsDir,
	}, nil
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	inputFile, err := os.Open(path.Join(cfg.InputDir, cfg.InputFile))
	if err != nil {
		log.Fatalf("Failed to open input file: %v", err)
	}

	inputData, err := io.ReadAll(inputFile)
	if err != nil {
		inputFile.Close()
		log.Fatalf("Error reading input file: %v", err)
	}
	inputFile.Close()

	var messagesIn []MessageIn
	err = json.Unmarshal(inputData, &messagesIn)
	if err != nil {
		log.Fatalf("Invalid JSON input: %v", err)
	}

	messages := []Message{}

	for i, msg := range messagesIn {
		messages = append(messages, Message{Role: msg.Role, Content: msg.Content})

		if msg.Role == SYSTEM && msg.File != "" {
			systemMsgFile, err := os.Open(path.Join(cfg.PromptsDir, msg.File))
			if err != nil {
				log.Fatalf("Failed to open system message file: %v", err)
			}

			systemMsgData, err := io.ReadAll(systemMsgFile)
			if err != nil {
				systemMsgFile.Close()
				log.Fatalf("Error reading system message file: %v", err)
			}
			systemMsgFile.Close()

			messages[i].Content = string(systemMsgData)
		}
	}

	displayInitScreen(messages, cfg.Model, float32(cfg.Temperature))

	reader := bufio.NewReader(os.Stdin)

	msgsCount := len(messages)
	if msgsCount == 0 || messages[msgsCount-1].Role != USER {
		userInput, err := readUserInput(reader)
		if err != nil {
			log.Fatalf("Failed to read user input: %v", err)
		}

		if userInput == "/quit!" {
			return
		} else if userInput == "/quit" {
			if err := saveConversationLog(messages, cfg.Model, cfg.LogsDir); err != nil {
				log.Printf("Error saving conversation log: %v", err)
			}
			return
		}

		messages = append(messages, Message{Role: USER, Content: userInput})
	}

	client := &http.Client{}
	payload := RequestPayload{
		Model:       cfg.Model,
		Temperature: float32(cfg.Temperature),
	}
	var responseBody ResponseBody

	for {
		payload.Messages = messages
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshalling payload: %v", err)
			continue
		}

		req, err := http.NewRequest("POST", cfg.URL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+cfg.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error sending request: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			log.Printf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
			fmt.Printf("!! API Error: %s\n", string(bodyBytes))
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			continue
		}

		if err := json.Unmarshal(body, &responseBody); err != nil {
			log.Printf("Error unmarshalling response body: %v", err)
			fmt.Printf("Raw response: %s\n", string(body))
			continue
		}

		if len(responseBody.Choices) > 0 {
			assistantMessage := responseBody.Choices[0].Message
			messages = append(messages, assistantMessage)

			fmt.Printf("<< %s\n", assistantMessage.Content)
			fmt.Printf("\n[Input: %d tokens, Output: %d tokens]\n",
				responseBody.Usage.PromptTokens,
				responseBody.Usage.CompletionTokens,
			)
		} else {
			fmt.Printf("!! Error: No response from API\n\n")
			fmt.Println(string(body))
			fmt.Println("\n> /quit to save and exit")
			fmt.Println("> /quit! to exit without saving")
		}

		fmt.Println()
		userInput, err := readUserInput(reader)
		if err != nil {
			log.Printf("Error reading user input: %v", err)
			continue
		}

		if userInput == "/quit!" {
			return
		} else if userInput == "/quit" {
			if err := saveConversationLog(messages, cfg.Model, cfg.LogsDir); err != nil {
				log.Printf("Error saving conversation log: %v", err)
			}
			return
		}

		messages = append(messages, Message{Role: USER, Content: userInput})
	}
}
