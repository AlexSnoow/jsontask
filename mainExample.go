package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

func main() {
	inputDir := "./IN"
	outputDir := "./OUT"

	// Создаем выходную директорию
	if err := os.MkdirAll(outputDir, 0750); err != nil {
		log.Fatalf("Ошибка создания директории: %v", err)
	}

	// Обрабатываем файлы
	if err := processFiles(inputDir, outputDir); err != nil {
		log.Fatalf("Ошибка обработки: %v", err)
	}

	log.Println("Конвертация завершена успешно")
}

// Структуры для требуемого формата JSON
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

func processFiles(inputDir, outputDir string) error {
	return filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Пропуск %s: %v", path, err)
			return nil
		}

		// Пропускаем директории
		if info.IsDir() {
			return nil
		}

		// Фильтруем только .md файлы
		if filepath.Ext(info.Name()) != ".md" {
			return nil
		}

		// Читаем содержимое файла
		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("Пропуск %s: ошибка чтения - %v", path, err)
			return nil
		}

		// Формируем имя файла без расширения
		filename := info.Name()[:len(info.Name())-len(filepath.Ext(info.Name()))]

		// Создаем запрос в требуемом формате
		request := Request{
			Model: "gemma3:1b",
			Messages: []Message{
				{
					Role:    "system",
					Content: "Ты — помощник для анализа текстов. Отвечай кратко и по делу.",
				},
				{
					Role:    "user",
					Content: "Начало текста: " + string(content) + "\n\nКонец текста.",
				},
			},
			Options: Options{
				Temperature: 0.3,
				NumCtx:      2048,
			},
			Stream: false,
		}

		// Конвертируем в JSON
		jsonData, err := json.MarshalIndent(request, "", "  ")
		if err != nil {
			log.Printf("Пропуск %s: ошибка JSON - %v", path, err)
			return nil
		}

		// Сохраняем результат
		outputPath := filepath.Join(outputDir, filename+".json")
		if err := os.WriteFile(outputPath, jsonData, 0644); err != nil {
			log.Printf("Ошибка сохранения %s: %v", outputPath, err)
		}

		log.Printf("Обработан: %s -> %s", path, outputPath)
		return nil
	})
}
