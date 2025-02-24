package storage

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

// TestNewDatabase_EmptyDSN - ошибка при пустой строке подключения
func TestNewDatabase_EmptyDSN(t *testing.T) {
	db, err := NewDatabase("")
	assert.Error(t, err)
	assert.Nil(t, db)
}

// TestDatabase_Close - вызывается ли Close()
func TestDatabase_Close(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockStorager(ctrl)
	mockDB.EXPECT().Close().Return(nil)

	err := mockDB.Close()

	assert.NoError(t, err)
}

// TestDatabase_Ping_NilDB - проверяем пинга вернет nil
func TestDatabase_Ping_NilDB(t *testing.T) {
	var db *Database
	err := db.Ping()
	assert.Error(t, err)
	assert.Equal(t, "database connection is nil", err.Error())
}

// TestDatabase_Close_NilDB - правильное закрытие при nil
func TestDatabase_Close_NilDB(t *testing.T) {
	var db *Database
	err := db.Close()
	assert.Error(t, err)
	assert.Equal(t, "database connection is already closed or uninitialized", err.Error())
}

// TestDatabase_Close_Mock - вызовем ли Close()
func TestDatabase_Close_Mock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := NewMockStorager(ctrl)
	mockDB.EXPECT().Close().Return(nil)

	err := mockDB.Close()
	assert.NoError(t, err)
}
