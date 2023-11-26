package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"

	"github.com/sashabaranov/go-openai"
)

type Openai struct {
	APIKey string
}

func main() {
	openaiResource := Openai{
		APIKey: "Enter your API key", // Replace with your API key
	}
	audioFilePath := "hello.mp3" // Replace with your audio file path

	resp, err := TranscribeAudio(openaiResource, audioFilePath)
	if err != nil {
		fmt.Println("Error in transcribing:", err)
	}

	ans, err := getResponse(openaiResource, resp)
	if err != nil {
		fmt.Println("error in getting resp", err)
	}

	err = generateSpeech(openaiResource, ans)
	if err != nil {
		fmt.Println("Error in generation:", err)
	}
}

func TranscribeAudio(openaiResource Openai, audioFilePath string) (string, error) {
	// Read the audio file
	audioFile, err := os.Open(audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open audio file: %v", err)
	}
	defer audioFile.Close()

	// Prepare the multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", audioFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %v", err)
	}
	_, err = io.Copy(part, audioFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy audio file to form file: %v", err)
	}

	writer.WriteField("model", "whisper-1")
	writer.Close()

	// Make the POST request to the OpenAI API
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+openaiResource.APIKey)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API request failed with status %d", response.StatusCode)
	}

	// Parse and return the response data
	responseData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response data: %v", err)
	}
	responseString := string(responseData)

	// Create an instance of the struct
	var data map[string]interface{}

	// Unmarshal the JSON string into the struct
	err = json.Unmarshal([]byte(responseString), &data)
	if err != nil {
		fmt.Println("Error:", err)
		return "", fmt.Errorf("failed to read response data: %v", err)
	}

	// Access the desired field
	resp, ok := data["text"].(string)
	if !ok {
		fmt.Println("Error: 'text' field is not a string")
		return "", fmt.Errorf("failed to convert response data: %v", err)
	}

	// Print the resul

	// fmt.Println()
	return resp, nil
}

func getResponse(openaiResource Openai, input string) (string, error) {
	token := openaiResource.APIKey
	client := openai.NewClient(token)
	response, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: input,
				},
			},
		},
	)
	if err != nil {
		log.Printf("error occurred here:  %v", err)
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}

func generateSpeech(openaiResource Openai, output string) error {
	// Extract the text from the previous output
	text := output

	// Prepare the request body
	body := map[string]interface{}{
		"model": "tts-1",
		"input": text,
		"voice": "alloy",
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Make the POST request to the OpenAI API
	request, err := http.NewRequest("POST", "https://api.openai.com/v1/audio/speech", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+openaiResource.APIKey)
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("OpenAI API request failed with status %d", response.StatusCode)
	}

	// Get the response data as a byte slice
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read response data: %v", err)
	}

	// Write the data to an mp3 file
	err = os.WriteFile("speech.mp3", data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write data to file: %v", err)
	}

	return nil
}
