package service

import (
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/mkukarin01/snort/internal/storage"
)

// TestNewURLDeleter проверка структурки + старт/стоп
func TestNewURLDeleter(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := storage.NewMockStorager(mockCtrl)
	deleter := NewURLDeleter(mockStore)

	assert.NotNil(t, deleter)
	assert.Equal(t, mockStore, deleter.store)
	assert.NotNil(t, deleter.inChan)
	assert.NotNil(t, deleter.stopCh)
	assert.Equal(t, 100, deleter.bufSize)

	go deleter.Run()
	time.Sleep(500 * time.Millisecond)
	deleter.Stop()
}

// TestURLDeleter_FlushByBufferThreshold когда буфер заполнен =>
// вызывается flush и MarkUserURLsDeleted из стораджа.
func TestURLDeleter_FlushByBufferThreshold(t *testing.T) {
	// Уменьшим bufSize для быстрого теста
	const testBufSize = 3

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := storage.NewMockStorager(mockCtrl)

	deleter := &URLDeleter{
		store:   mockStore,
		inChan:  make(chan deleteJob, 10),
		stopCh:  make(chan struct{}),
		bufSize: testBufSize,
	}

	// Сценарий: +2 задачи от одного юид, а потом +1
	// flush должен сработать после третьей задачи
	userID := "test-user"
	shortIDsAll := []string{"id1", "id2", "id3"}
	// MarkUserURLsDeleted - вызывается 1 раз
	mockStore.
		EXPECT().
		MarkUserURLsDeleted(userID, shortIDsAll).
		Return(nil).
		Times(1)

	// запуск в отдельной рутине
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deleter.Run()
	}()

	// сабмит задач
	deleter.Submit(userID, []string{"id1", "id2"}) // +2
	deleter.Submit(userID, []string{"id3"})        // +1 === bufSize

	// поспим
	time.Sleep(100 * time.Millisecond)

	// остановка
	close(deleter.stopCh)
	wg.Wait()
}

// TestURLDeleter_FlushByTimeout проверка, истечение таймера => flush, даже если буфер еще ок
func TestURLDeleter_FlushByTimeout(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := storage.NewMockStorager(mockCtrl)

	// оставляю реальную секунду для наглядности, в реальных кейсах кажется так делать не стоит
	deleter := &URLDeleter{
		store:   mockStore,
		inChan:  make(chan deleteJob, 10),
		stopCh:  make(chan struct{}),
		bufSize: 10,
	}

	userID := "test-user-timeout"
	shortIDsAll := []string{"id1", "id2"}

	// MarkUserURLsDeleted по таймеру должен сработать
	mockStore.
		EXPECT().
		MarkUserURLsDeleted(userID, shortIDsAll).
		Return(nil).
		Times(1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deleter.Run()
	}()

	// отправляем задачу
	deleter.Submit(userID, shortIDsAll)

	// 1.2 секунды чтобы точно сработал flush
	time.Sleep(1200 * time.Millisecond)

	close(deleter.stopCh)
	wg.Wait()
}

// TestURLDeleter_Stop закрытие тоже вызывает flush
func TestURLDeleter_Stop(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := storage.NewMockStorager(mockCtrl)

	deleter := &URLDeleter{
		store:   mockStore,
		inChan:  make(chan deleteJob, 10),
		stopCh:  make(chan struct{}),
		bufSize: 100,
	}

	userID := "test-user-stop"
	shortIDsAll := []string{"id1", "id2", "id3"}

	mockStore.
		EXPECT().
		MarkUserURLsDeleted(userID, shortIDsAll).
		Return(nil).
		Times(1)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		deleter.Run()
	}()

	// +3
	deleter.Submit(userID, shortIDsAll)

	// run жует данные, пусть будет 300
	time.Sleep(300 * time.Millisecond)

	// стопаем канал
	close(deleter.stopCh)
	wg.Wait()
}

// TestURLDeleter_flush - дополнительный не красивый тест приватной функции, потому что могу
// так лучше не делать наверно
func TestURLDeleter_flush(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockStore := storage.NewMockStorager(mockCtrl)

	deleter := &URLDeleter{
		store:   mockStore,
		inChan:  make(chan deleteJob),
		stopCh:  make(chan struct{}),
		bufSize: 100,
	}

	userBatch := map[string][]string{
		"userA": {"short1", "short2"},
		"userB": {"short3"},
	}

	// 2 пользака === 2 вызова
	mockStore.EXPECT().MarkUserURLsDeleted("userA", []string{"short1", "short2"}).Return(nil).Times(1)
	mockStore.EXPECT().MarkUserURLsDeleted("userB", []string{"short3"}).Return(nil).Times(1)

	// стопаем канал
	deleter.flush(userBatch)
}
