# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).

## Структура проекта

Приведённая в этом репозитории структура проекта является рекомендуемой, но не обязательной.

Это лишь пример организации кода, который поможет вам в реализации сервиса.

При необходимости можно вносить изменения в структуру проекта, использовать любые библиотеки и предпочитаемые структурные паттерны организации кода приложения, например:
- **DDD** (Domain-Driven Design)
- **Clean Architecture**
- **Hexagonal Architecture**
- **Layered Architecture**

## Полученный вывод
```shell
File: ___go_build_github_com_delyke_urlShortener_cmd_shortener
Type: inuse_space
Time: 2026-01-18 22:03:46 +07
Showing nodes accounting for -1192.86kB, 12.99% of 9182.82kB total
      flat  flat%   sum%        cum   cum%
 -678.95kB  7.39%  7.39%  -678.95kB  7.39%  github.com/delyke/urlShortener/internal/repository.(*LocalRepository).CreateUser
     514kB  5.60%  1.80%      514kB  5.60%  bufio.NewReaderSize (inline)
    -514kB  5.60%  7.39%     -514kB  5.60%  bufio.NewWriterSize (inline)
    -513kB  5.59% 12.98%     -513kB  5.59%  runtime.allocm
 -512.75kB  5.58% 18.56%  -512.75kB  5.58%  sync.(*Pool).pinSlow
 -512.22kB  5.58% 24.14%  -512.22kB  5.58%  runtime.malg
 -512.05kB  5.58% 29.72%  -512.05kB  5.58%  github.com/delyke/urlShortener/internal/service.fanIn[go.shape.struct { UserID int64; ShortenedURL string }].func1
  512.05kB  5.58% 24.14%   512.05kB  5.58%  runtime.acquireSudog
  512.05kB  5.58% 18.57%   512.05kB  5.58%  sync.runtime_SemacquireWaitGroup
  512.02kB  5.58% 12.99%   512.02kB  5.58%  syscall.anyToSockaddr
         0     0% 12.99%      514kB  5.60%  bufio.NewReader (inline)
         0     0% 12.99% -1191.70kB 12.98%  github.com/delyke/urlShortener/internal/app.NewRouter.authCookieMiddleware.func4.1
         0     0% 12.99% -1191.70kB 12.98%  github.com/delyke/urlShortener/internal/app.gzipMiddleware.func1
         0     0% 12.99%  -512.75kB  5.58%  github.com/delyke/urlShortener/internal/handler.(*Handler).HandlePost
         0     0% 12.99%  -512.75kB  5.58%  github.com/delyke/urlShortener/internal/logger.(*Logger).RequestLogger-fm.(*Logger).RequestLogger.func1
         0     0% 12.99%  -678.95kB  7.39%  github.com/delyke/urlShortener/internal/service.(*URLService).CreateUser
         0     0% 12.99%   512.05kB  5.58%  github.com/delyke/urlShortener/internal/service.fanIn[go.shape.struct { UserID int64; ShortenedURL string }].func2
         0     0% 12.99%  -512.75kB  5.58%  github.com/go-chi/chi/v5.(*Mux).Mount.func1
         0     0% 12.99% -1191.70kB 12.98%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 12.99%  -512.75kB  5.58%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap.(*SugaredLogger).Errorf (inline)
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap.(*SugaredLogger).log
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/internal/pool.(*Pool[go.shape.*uint8]).Get (inline)
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.(*CheckedEntry).Write
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.(*ioCore).Write
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.(*jsonEncoder).Clone
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.(*jsonEncoder).clone
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.consoleEncoder.EncodeEntry
         0     0% 12.99%  -512.75kB  5.58%  go.uber.org/zap/zapcore.consoleEncoder.writeContext
         0     0% 12.99%   512.02kB  5.58%  internal/poll.(*FD).Accept
         0     0% 12.99%   512.02kB  5.58%  internal/poll.accept
         0     0% 12.99%   512.02kB  5.58%  main.main
         0     0% 12.99%   512.02kB  5.58%  net.(*TCPListener).Accept
         0     0% 12.99%   512.02kB  5.58%  net.(*TCPListener).accept
         0     0% 12.99%   512.02kB  5.58%  net.(*netFD).accept
         0     0% 12.99%   512.02kB  5.58%  net/http.(*Server).ListenAndServe
         0     0% 12.99%   512.02kB  5.58%  net/http.(*Server).Serve
         0     0% 12.99% -1191.70kB 12.98%  net/http.(*conn).serve
         0     0% 12.99% -1191.70kB 12.98%  net/http.HandlerFunc.ServeHTTP
         0     0% 12.99%   512.02kB  5.58%  net/http.ListenAndServe (inline)
         0     0% 12.99%      514kB  5.60%  net/http.newBufioReader
         0     0% 12.99%     -514kB  5.60%  net/http.newBufioWriterSize
         0     0% 12.99% -1191.70kB 12.98%  net/http.serverHandler.ServeHTTP
         0     0% 12.99%   512.05kB  5.58%  runtime.chanrecv
         0     0% 12.99%   512.05kB  5.58%  runtime.chanrecv1
         0     0% 12.99%   512.02kB  5.58%  runtime.main
         0     0% 12.99%     -513kB  5.59%  runtime.mcall
         0     0% 12.99%     -513kB  5.59%  runtime.newm
         0     0% 12.99%  -512.22kB  5.58%  runtime.newproc.func1
         0     0% 12.99%  -512.22kB  5.58%  runtime.newproc1
         0     0% 12.99%     -513kB  5.59%  runtime.park_m
         0     0% 12.99%     -513kB  5.59%  runtime.resetspinning
         0     0% 12.99%     -513kB  5.59%  runtime.schedule
         0     0% 12.99%     -513kB  5.59%  runtime.startm
         0     0% 12.99%  -512.22kB  5.58%  runtime.systemstack
         0     0% 12.99%   512.05kB  5.58%  runtime.unique_runtime_registerUniqueMapCleanup.func2
         0     0% 12.99%     -513kB  5.59%  runtime.wakep
         0     0% 12.99%  -512.75kB  5.58%  sync.(*Pool).Get
         0     0% 12.99%  -512.75kB  5.58%  sync.(*Pool).pin
         0     0% 12.99%   512.05kB  5.58%  sync.(*WaitGroup).Wait
         0     0% 12.99%   512.02kB  5.58%  syscall.Accept

```