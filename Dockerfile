FROM postgres:15

# Указываем переменные окружения для автоматического создания БД
ENV POSTGRES_USER=myuser
ENV POSTGRES_PASSWORD=mypassword
ENV POSTGRES_DB=mydatabase

# Копируем SQL-скрипты для инициализации данных
COPY init.sql /docker-entrypoint-initdb.d/

EXPOSE 5432
