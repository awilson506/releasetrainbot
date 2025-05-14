# 🚂 Slack Rotation Bot

A simple Slack bot built in Go using Gin that manages a rotation of Slack users for a weekly responsibility (e.g., release train engineer). It supports a `/release-train` slash command and posts the current lead to a Slack channel every Monday.

---

## Slack Command Docs

- [Slack Command API](https://api.slack.com/interactivity/slash-commands)

## ⏰ Weekly Cron Schedule

- Every Monday at 9:00 AM UTC
- Posts a message to the channel with the current lead
- Rotates the order afterward (first becomes last)

## Database

Start a MySQL 8 database using Docker Compose:

```bash
docker-compose up -d
```

Migrations using Goose

```bash
$ goose create add_some_column sql
$ Created new file: 20170506082420_add_some_column.sql
```

## 📝 API Documentation (Swagger)

The API documentation is generated using **Swaggo**. You can view and interact with the API documentation using Swagger UI.

## View the Swagger UI

Once the app is running, open the following URL in your browser: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

## Comands

| Command                     | Behavior                      |
| --------------------------- | ----------------------------- |
| `/release-train`            | Get current rotation schedule |
| `/release-train @user1 ...` | Set/update rotation           |
| `/release-train delete`     | Delete rotation               |
