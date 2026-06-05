
Описание
Данная программа поддерживает управление книгами, читателями, выдачу и возврат книг

 Запуск

1. `cp .env.example .env`
2. `go mod tidy`
3. `go run cmd/server/main.go`

Сервер доступен на http://localhost:8080

 Примеры curl

 Авторизация
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

 Книги 
```bash
curl http://localhost:8080/books
curl http://localhost:8080/books?page=1&limit=5&author=Толстой
```

Создать книгу 
```bash
curl -X POST http://localhost:8080/books \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"title":"Война и мир","author":"Лев Толстой","isbn":"123456","year":1869}'
```

Выдать книгу
```bash
curl -X POST http://localhost:8080/issues \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{"book_id":"...", "user_id":"..."}'
```

