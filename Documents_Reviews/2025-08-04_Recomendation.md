### 1. **Принцип постепенной сложности**
- **Проблема**: сразу использовал горутины/каналы без необходимости
- **Заметки**: 
  - Начинать с синхронной версии (`for` + `filepath.WalkDir`)
  - Добавлять конкурентность только при реальной потребности (например, при работе с сетью в Telegram)
  - Пример эволюции кода:
    ```go
    // Этап 1: Синхронная обработка
    for _, file := range files { Process(file) }
    
    // Этап 2: Worker pool (когда файлов >1000)
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go func(ch <-chan File) {
            defer wg.Done()
            for file := range ch { Process(file) }
        }(ch)
    }
    ```

### 2. **Обработка ошибок как first-class citizen**
- **Критические ошибки**:
  - `panic/recover` для бизнес-логики
  - Выход из программы при первой же ошибке
- **Решение**:
  ```go
  // Вместо паники
  if err := Process(); err != nil {
      log.Printf("Ошибка обработки %s: %v", fileName, err)
      continue // Пропустить файл, но продолжить работу
  }
  ```
  - Разделить ошибки на:
    - Критические (неверная конфигурация) → `log.Fatal`
    - Файловые (отсутствует файл) → лог + пропуск
    - Временные (сеть) → retry

### 3. **Разделение ответственности (для проекта с Telegram)**
Архитектура будущего проекта:
```go
type (
    TelegramFetcher  struct { /* ... */ } // Получение сообщений
    MessageProcessor struct { /* ... */ } // Обработка через LLM
    MarkdownSaver    struct { /* ... */ } // Сохранение в .md
)

func main() {
    fetcher := NewTelegramFetcher(token)
    processor := NewMessageProcessor(llmConfig)
    saver := NewMarkdownSaver("./OUT")
    
    messages, err := fetcher.Fetch(ctx)
    // ...
    processed := processor.Transform(messages)
    // ...
    saver.SaveAll(processed)
}
```

### 4. **Работа с ресурсами**
Что упустил и что критично для проекта:
1. **Закрытие ресурсов**:
   ```go
   // Для будущего работы с сетевыми соединениями
   defer response.Body.Close()
   ```
2. **Ограничение числа открытых файлов** (когда файлов тысячи)
3. **Таймауты** (особенно для Telegram/LLM):
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
   defer cancel()
   ```

### 5. **Конфигурация и шаблоны**
- **Жесткие параметры** (модель, температура) → вынести в конфиг:
  ```go
  type LLMConfig struct {
      Model       string  `yaml:"model"`
      Temperature float64 `yaml:"temperature"`
  }
  ```
- **Шаблоны сообщений** (для LLM) → использовать `text/template`:
  ```go
  const promptTemplate = `Текст: {{.Content}}`
  ```

### 6. **Тестируемость**
писать тесты для компонентов:
```go
func TestSaveJSON(t *testing.T) {
    t.Run("Создание директории при отсутствии", func(t *testing.T) {
        // Тест на os.MkdirAll
    })
}
```

### 7. **Идиоматические практики Go**
- **Интерфейсы вместо структур**:
  ```go
  type ContentSaver interface {
      Save(data FileData) error
  }
  ```
- **Работа с путями**: `filepath.Join` вместо конкатенации
- **Буферизованная обработка** для работы с большими файлами

### Рекомендации для ментора
1. **Code Review Focus**:
   - Первый PR разбирать вместе: "Давай посмотрим, где можно упростить"
   - Разница между "работает" и "поддерживаемо"
2. **Практические упражнения**:
   - "Как обработать 10K файлов без OOM?"
   - "Как добавить retry для Telegram API?"
3. **Инструменты**:
   - `go vet` для проверки закрытия ресурсов
   - `pprof` для анализа памяти

**Ключевое сообщение для новичка**:  
"Твой код решает задачу, но представь, что к нему нужно добавить обработку Telegram завтра. Как сделать так, чтобы изменения требовали 5 минут, а не 5 часов?"