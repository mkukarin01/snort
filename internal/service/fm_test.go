package service

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSaveAndLoad тест - данные корректно сохраняются и загружаются
func TestSaveAndLoad(t *testing.T) {
	testFile := "test_storage.json"
	defer os.Remove(testFile) // очистка

	us := NewURLShortener(testFile)
	us.store = map[string]string{
		"short1": "http://example.com",
		"short2": "http://example.org",
	}

	us.saveToFile()

	// новый экземпляр и загрузка
	usNew := NewURLShortener(testFile)
	err := usNew.loadFromFile()
	assert.NoError(t, err, "Ошибка при загрузке данных из файла")

	// чекаем данные в хранилище
	assert.Equal(t, us.store, usNew.store, "Загруженные данные не совпадают с исходными")
}

// TestFileCreation тест - файл создается, если его не существовало
func TestFileCreation(t *testing.T) {
	testFile := "test_empty_storage.json"
	defer os.Remove(testFile)

	us := NewURLShortener(testFile)

	us.saveToFile()

	_, err := os.Stat(testFile)
	assert.NoError(t, err, "Файл не был создан")
}

// TestEmptyLoad тест - загрузка из пустого файла не вызывает ошибок
func TestEmptyLoad(t *testing.T) {
	testFile := "test_empty_file.json"
	defer os.Remove(testFile)

	file, err := os.Create(testFile)
	assert.NoError(t, err, "Ошибка при создании тестового файла")
	file.Close()

	us := NewURLShortener(testFile)
	err = us.loadFromFile()
	assert.NoError(t, err, "Ошибка загрузки из пустого файла")
	assert.Equal(t, 0, len(us.store), "В хранилище не должно быть записей")
}

// TestCorruptFile тест битого json
func TestCorruptFile(t *testing.T) {
	testFile := "test_corrupt.json"
	defer os.Remove(testFile)

	err := os.WriteFile(testFile, []byte("invalid json"), 0644)
	assert.NoError(t, err, "Ошибка при создании тестового файла")

	us := NewURLShortener(testFile)
	err = us.loadFromFile()
	assert.Error(t, err, "Ожидалась ошибка на поврежденном файле")
}

// TestSaveFilePermissions тест пермишеннов при сохранении
func TestSaveFilePermissions(t *testing.T) {
	testFile := "test_permissions.json"
	defer os.Remove(testFile)

	file, err := os.Create(testFile)
	assert.NoError(t, err)
	file.Close()

	err = os.Chmod(testFile, 0444) // -r-r-r- | RO
	assert.NoError(t, err)

	us := NewURLShortener(testFile)
	us.store["test"] = "http://ya.ru"

	us.saveToFile()

	os.Chmod(testFile, 0644)
}

// TestLoadFilePermissions тест пермишеннов при чтении
func TestLoadFilePermissions(t *testing.T) {
	testFile := "test_read_permissions.json"
	defer os.Remove(testFile)

	entry := map[string]string{"short_url": "short1", "original_url": "http://ya.ru"}
	file, err := os.Create(testFile)
	assert.NoError(t, err, "Ошибка при создании тестового файла")

	enc := json.NewEncoder(file)
	err = enc.Encode(entry)
	assert.NoError(t, err, "Ошибка при записи JSON в файл")
	file.Close()

	// Запрещаем чтение
	err = os.Chmod(testFile, 0000) // возможно все таки нужно подсунуть 0222 aka -w-w-w-
	assert.NoError(t, err, "Ошибка при изменении прав доступа к файлу")

	us := NewURLShortener(testFile)
	err = us.loadFromFile()

	if err == nil {
		assert.Equal(t, 0, len(us.store), "Загруженные данные должны быть пустыми при отсутствии доступа")
	} else {
		assert.Error(t, err, "Ожидалась ошибка при запрете чтения файла")
	}

	os.Chmod(testFile, 0644)
}
