# Growth MVP (тестовое задание)

## Запуск

```
cp .env.example .env
docker-compose up --build
```

После этого UI будет доступен на [http://localhost:9999](http://localhost:9999), а API - на [http://localhost:8080](http://localhost:8080).

## Endpoints

- `POST /shops/:shopId/telegram/connect`  
  Подключить или обновить Telegram-интеграцию для магазина.

  Пример body:
  ```json
  {
    "botToken": "123456:ABCDEF...",
    "chatId": "-1001234567890",
    "enabled": true
  }
  ```

- `GET /shops/:shopId/telegram/status`  
  Получить статус Telegram-интеграции и статистику отправок за 7 дней.

- `POST /shops/:shopId/orders`  
  Создать заказ и запустить отправку уведомления в Telegram (если интеграция активна).

  Пример body:
  ```json
  {
    "number": "A-1001",
    "total": 1990.50,
    "customerName": "Иван Иванов"
  }
  ```

- `GET /shops/:shopId/orders?limit=20&offset=0`  
  Получить список заказов с пагинацией.

## Примечание

Я позволил себе слегка отступить от ТЗ: отправка сообщения в Telegram API осуществляется асинхронно (и с retry), т.к. считаю, что взаимодействиям со сторонним API не место в цикле запроса даже в MVP или прототипе.