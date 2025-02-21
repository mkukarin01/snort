package logger

// следуя Clean Architecture каждая внешняя библиотека и ее синглтончики лучше
// реализовавать отдельно от имплементации, таким образом можно написать функцию
// которая будет реализовывать ретрофит апи и новый апи для новой библиотеки
// (банальный ремап аргументов, в другие аргументы, обогащие и тд и тп).
// https://alexkondov.com/full-stack-tao-clean-architecture-react/
// https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html

import (
	"github.com/sirupsen/logrus"
)

// InitLogger инициализирует и сигнлтонит
func InitLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	return log
}
