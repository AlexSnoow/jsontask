/*
Package main содержит реализацию программы для извлечения, парсинга и сохранения JSON-данных из файлов в директории "IN".

Работа программы:
1. Создаются два канала: contentChan для передачи данных типа FileData и errChan для ошибок.
2. Запускается горутина, которая:
  - Вызывает функцию ExtractContent для обработки директории "./IN".
  - Ловит паники и отправляет ошибки в errChan.

3. Основной поток обрабатывает данные из contentChan:
  - Парсит JSON с помощью ParseJSON.
  - Сохраняет результат в файл через SaveJSON.

4. При возникновении ошибки или закрытии обоих каналов программа завершается.

Примечание:
- Все ошибки выводятся в stdout.
- Используется механизм recover для обработки паник.
- Каналы закрываются после завершения работы.
*/
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

	// Горутина для безопасного извлечения данных
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

// FileData представляет данные, извлекаемые из файла.
type FileData struct {
	// Поля структуры зависят от реализации ExtractContent
	Name    string
	Content string
}

// ExtractContent извлекает данные из файлов в указанной директории.
// Отправляет результаты в contentChan и ошибки в errChan.
func ExtractContent(inDir string, contentChan chan FileData) error {
	// Реализация зависит от контекста
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

// ParseJSON преобразует сырые данные в структуру JSON.
func ParseJSON(data FileData) ([]byte, error) {
	// Реализация зависит от контекста
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

// SaveJSON сохраняет распарсенные данные в файл.
func SaveJSON(data FileData, json []byte) error {
	// Реализация зависит от контекста
	outputFile := filepath.Join("./OUT", data.Name+".json")
	err := os.WriteFile(outputFile, json, 0644)
	if err != nil {
		return err
	}
	return nil
}
