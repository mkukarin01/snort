package service

import (
	"encoding/json"
	"fmt"
	"os"
)

// saveToFile сохраняем данные в хранилище json
func (us *URLShortener) saveToFile() {
	us.RLock()
	defer us.RUnlock()

	if us.fileStoragePath == "" {
		return
	}

	file, err := os.Create(us.fileStoragePath)
	if err != nil {
		fmt.Printf("Failed to save storage file: %v\n", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	for id, url := range us.store {
		entry := map[string]string{"short_url": id, "original_url": url}
		if err := enc.Encode(entry); err != nil {
			fmt.Printf("Failed to encode entry: %v\n", err)
		}
	}
}

// loadFromFile выгружаем данные из файла
func (us *URLShortener) loadFromFile() error {
	file, err := os.Open(us.fileStoragePath)
	if err != nil {
		// если файл не существует, создаем новый
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to open storage file: %w", err)
	}
	defer file.Close()

	dec := json.NewDecoder(file)
	for {
		entry := make(map[string]string)
		if err := dec.Decode(&entry); err != nil {
			if err.Error() == "EOF" { // Читаем до конца, но не выдаём ошибку если это просто конец файла
				break
			}
			return fmt.Errorf("failed to decode JSON: %w", err)
		}
		us.store[entry["short_url"]] = entry["original_url"]
	}
	return nil
}
