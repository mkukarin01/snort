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
	url, found := ms.Load("testKey")
	assert.True(t, found)
	assert.Equal(t, "http://ya.ru", url)
}

// TestMemoryStorage_NotFound - проверяем отсутствие ключа
func TestMemoryStorage_NotFound(t *testing.T) {
	ms := NewMemoryStorage()

	// Загружаем несуществующий ключ
	url, found := ms.Load("missing")
	assert.False(t, found)
	assert.Empty(t, url)
}

func TestMemoryStorage_PingClose(t *testing.T) {
	ms := NewMemoryStorage()

	ping := ms.Ping()
	assert.Error(t, ping)

	close := ms.Close()
	assert.Nil(t, close)
}
