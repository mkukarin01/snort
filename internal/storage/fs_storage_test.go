package storage

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestFileStorage_SaveLoad - тестируем сохранение и загрузку
func TestFileStorage_SaveLoad(t *testing.T) {
	tempFile, err := os.CreateTemp("", "storage.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	fs, err := NewFileStorage(tempFile.Name())
	assert.NoError(t, err)

	err = fs.Save("testID", "http://ya.ru")
	assert.NoError(t, err)

	url, found := fs.Load("testID")
	assert.True(t, found)
	assert.Equal(t, "http://ya.ru", url)
}

// TestFileStorage_EmptyFile - проверка работы с пустым файлом
func TestFileStorage_EmptyFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "empty.json")
	assert.NoError(t, err)
	defer os.Remove(tempFile.Name())

	fs, err := NewFileStorage(tempFile.Name())
	assert.NoError(t, err)

	url, found := fs.Load("unknownID")
	assert.False(t, found)
	assert.Empty(t, url)
}

func TestFileStorage_CorruptFile(t *testing.T) {
	testFile := "../../test_corrupt.json"
	defer os.Remove(testFile)

	err := os.WriteFile(testFile, []byte("invalid json"), 0644)
	assert.NoError(t, err, "Ошибка при создании тестового файла")

	_, err = NewFileStorage(testFile)
	assert.Error(t, err, "Ожидалась ошибка на поврежденном файле")
}

func TestFileStorage_Permissions(t *testing.T) {
	testFile := "test_permissions.json"
	defer os.Remove(testFile)

	file, err := os.Create(testFile)
	assert.NoError(t, err)
	file.Close()

	err = os.Chmod(testFile, 0444) // -r-r-r- | RO
	assert.NoError(t, err)

	storage, _ := NewFileStorage(testFile)

	storage.Save("test", "http://ya.ru")

	os.Chmod(testFile, 0644)
}

func TestFileStorage_PingClose(t *testing.T) {
	ms := NewMemoryStorage()

	ping := ms.Ping()
	assert.Error(t, ping)

	close := ms.Close()
	assert.Nil(t, close)
}
