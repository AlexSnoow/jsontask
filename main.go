package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	contentChan := make(chan FileData)
	errChan := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("panic %v", r)
				return
			}
		}()
		defer close(errChan)
		err := ExtractContent("./IN", contentChan)
		if err != nil {
			errChan <- err
		}
	}()

	for {
		select {
		case content, ok := <-contentChan:
			if !ok {
				contentChan = nil
			} else {
				jsonData, err := ParseJSON(content)
				if err != nil {
					fmt.Printf("error %v\n", err)
					return
				}
				err = SaveJSON(content, jsonData)
				if err != nil {
					fmt.Printf("error %v\n", err)
					return
				}
			}

		case err, ok := <-errChan:
			if ok {
				fmt.Printf("Критическая ошибка: %v\n", err)
				return
			} else {
				errChan = nil
			}
		}
		if contentChan == nil && errChan == nil {
			break
		}
	}

}

type FileData struct {
	Name    string
	Content string
}

func ExtractContent(inDir string, contentChan chan FileData) error {
	defer close(contentChan)

	err := filepath.WalkDir(inDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Ошибка обхода %s: %v\n", path, err)
			return nil
		}

		if !d.IsDir() && filepath.Ext(d.Name()) == ".md" {
			data, err := os.ReadFile(path)
			if err != nil {
				fmt.Printf("Ошибка чтения %s: %v\n", path, err)
				return nil
			}
			filename := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))

			contentChan <- FileData{
				Name:    filename,
				Content: string(data),
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("не удалось пройти директорию: %w", err)
	}

	return nil
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Options  Options   `json:"options"`
	Stream   bool      `json:"stream"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Options struct {
	Temperature float64 `json:"temperature"`
	NumCtx      int     `json:"num_ctx"`
}

func ParseJSON(data FileData) ([]byte, error) {
	req := Request{
		Model: "gemma3:1b",
		Messages: []Message{
			{Role: "system", Content: "Ты — помощник для анализа текстов. Отвечай кратко и по делу."},
			{Role: "user", Content: fmt.Sprintf("Начало текста: %s\n\nКонец текста.", data.Content)},
		},
		Options: Options{
			Temperature: 0.3,
			NumCtx:      2048,
		},
		Stream: false,
	}

	jsonData, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

func SaveJSON(data FileData, json []byte) error {
	outputFile := filepath.Join("./OUT", data.Name+".json")
	err := os.WriteFile(outputFile, json, 0644)
	if err != nil {
		return err
	}
	return nil
}
