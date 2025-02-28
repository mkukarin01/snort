package service

import (
	"log"
	"sync"
	"time"

	"github.com/mkukarin01/snort/internal/storage"
)

// deleteJob - структурка тасков на удаления
type deleteJob struct {
	userID   string
	shortIDs []string
}

// URLDeleter - структурка для фанин, по ней мы пачкой "удалим" записи
type URLDeleter struct {
	store   storage.Storager
	inChan  chan deleteJob
	stopCh  chan struct{}
	wg      sync.WaitGroup
	bufSize int
}

// NewURLDeleter - создаём новый агрегатор
func NewURLDeleter(store storage.Storager) *URLDeleter {
	return &URLDeleter{
		store:   store,
		inChan:  make(chan deleteJob, 100), // буфер
		stopCh:  make(chan struct{}),
		bufSize: 100, // сброс/флеш
	}
}

// Run запускатор, цикл на чтение и агрегация тасков по фанин
func (d *URLDeleter) Run() {
	// какой-то странный костыль, когда запускаешь горутину, если она ничего не сделала, а ты выполнил Stop
	// произойдет `panic: sync: negative WaitGroup counter` с ненулевым wg этой проблемы нет
	d.wg.Add(1)

	flushInterval := time.Second

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	// локальная мапа для агрегации
	userBatch := make(map[string][]string)

	flushFunc := func() {
		if len(userBatch) == 0 {
			return
		}
		d.flush(userBatch)
		// чистим
		userBatch = make(map[string][]string)
	}

	for {
		select {
		case <-d.stopCh:
			// финалочка
			flushFunc()
			d.wg.Done()
			return
		case job := <-d.inChan:
			// копим
			userBatch[job.userID] = append(userBatch[job.userID], job.shortIDs...)

			// при пороге - надо провести сброс
			var count int
			for _, ids := range userBatch {
				count += len(ids)
			}
			if count >= d.bufSize {
				flushFunc()
			}
		case <-ticker.C:
			// очистка таймера
			flushFunc()
		}
	}
}

// Submit - для хендлеров отправка задач
func (d *URLDeleter) Submit(userID string, shortIDs []string) {
	d.inChan <- deleteJob{
		userID:   userID,
		shortIDs: shortIDs,
	}
}

// Stop - остановить
func (d *URLDeleter) Stop() {
	close(d.stopCh)
	d.wg.Wait()
}

// flush - собственно сброс - выполняем задачи
func (d *URLDeleter) flush(userBatch map[string][]string) {
	for uid, sids := range userBatch {
		err := d.store.MarkUserURLsDeleted(uid, sids)
		if err != nil {
			log.Printf("ERROR: MarkUserURLsDeleted user=%s, shortIDs=%v, err=%v", uid, sids, err)
		} else {
			log.Printf("Deleted user URLs for user=%s, shortIDs=%v", uid, sids)
		}
	}
}
