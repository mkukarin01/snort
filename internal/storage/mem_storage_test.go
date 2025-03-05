package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMemoryStorage_SaveLoad - тестируем сохранение и загрузку
func TestMemoryStorage_SaveLoad(t *testing.T) {
	ms := NewMemoryStorage()

	err := ms.Save("testKey", "http://ya.ru")
	assert.NoError(t, err)

	// Загружаем обратно
	url, foundErr := ms.Load("testKey")
	assert.Nil(t, foundErr)
	assert.Equal(t, "http://ya.ru", url)
}

// TestMemoryStorage_NotFound - проверяем отсутствие ключа
func TestMemoryStorage_NotFound(t *testing.T) {
	ms := NewMemoryStorage()

	// Загружаем несуществующий ключ
	url, foundErr := ms.Load("missing")
	assert.ErrorIs(t, foundErr, ErrURLNotFound)
	assert.Empty(t, url)
}

func TestMemoryStorage_PingClose(t *testing.T) {
	ms := NewMemoryStorage()

	ping := ms.Ping()
	assert.Error(t, ping)

	close := ms.Close()
	assert.Nil(t, close)
}
