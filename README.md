# gophermart

Репозитория для индивидуального дипломного проекта курса «Go-разработчик»

# Used packages

* [Logrus](https://github.com/sirupsen/logrus)
* [Chi](https://github.com/go-chi/chi)
* [EasyJSON](https://github.com/mailru/easyjson)
* [JWT](https://github.com/golang-jwt/jwt)
* [Pgx](https://github.com/jackc/pgx)
* [goose](https://github.com/pressly/goose) - db migrations
* [rate]( golang.org/x/time/rate)

# Sources

* [Implementing JWT based authentication in Golang](https://www.sohamkamani.com/golang/jwt-authentication/)

# Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m master template https://github.com/yandex-praktikum/go-musthave-diploma-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/master .github
```

Затем добавьте полученные изменения в свой репозиторий.