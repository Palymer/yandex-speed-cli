# yandex-speed-cli

Консольная утилита на Go для замера скорости интернета через публичный API сервиса [Яндекс.Интернетометр](https://yandex.ru/internet): внешний IP, задержка до CDN (min / max / σ / p95), затем download и upload с полосой прогресса, таймером фазы и краткой сводкой. Результат можно вывести в JSON для скриптов и мониторинга.

> **Важно.** Проект не является официальным продуктом Яндекса. Он обращается к тем же HTTPS-эндпоинтам, что и веб-страница Интернетометра. Поведение API может измениться; при публикации своих измерений учитывайте нагрузку на сеть и условия использования сервисов Яндекса и сторонних API (геолокация).

## Возможности

- **IP** — запрос к API Яндекса (`/api/v0/ip`).
- **Задержка** — параллельный скан списка зондов, выбор лучшего узла, серия GET-запросов, статистика в миллисекундах.
- **Download / upload** — многопоточная загрузка по URL из ответа `get-probes`, учёт байт и мгновенной скорости по окнам.
- **Терминал** — цвета ANSI (отключаются в не-TTY и флагом `-no-color`), на Windows при поддержке консоли включается виртуальный терминал (VT).
- **Геолокация по IP** — опционально [ipwho.is](https://ipwho.is) для подписей региона (отключается `-no-geo`).
- **JSON** — машиночитаемый вывод (`-json`), удобно для CI и парсинга.

## Требования

- [Go](https://go.dev/dl/) **1.21+** — для сборки из исходников.
- Доступ в интернет до `yandex.ru` (и при включённой геолокации — до `ipwho.is`).

## Установка

### Готовые бинарники (GitHub Releases)

Бинарники публикуются в [**последнем релизе**](https://github.com/Palymer/yandex-speed-cli/releases/latest) и в [списке всех релизов](https://github.com/Palymer/yandex-speed-cli/releases). Версия для автосборки задаётся файлом **`VERSION`** в корне репозитория (сейчас **1.0.0** → тег **`v1.0.0`**).

| Файл | ОС / архитектура |
|------|-------------------|
| `yandex-speed-cli-linux-amd64` | Linux x86_64 |
| `yandex-speed-cli-linux-386` | Linux x86 |
| `yandex-speed-cli-linux-arm64` | Linux ARM64 (в т.ч. многие SBC) |
| `yandex-speed-cli-windows-amd64.exe` | Windows x64 |
| `yandex-speed-cli-windows-386.exe` | Windows x86 |
| `yandex-speed-cli-darwin-amd64` | macOS Intel |
| `yandex-speed-cli-darwin-arm64` | macOS Apple Silicon |
| `SHA256SUMS` | Контрольные суммы артефактов |

Прямая загрузка из **последнего** релиза (URL с `latest` редиректят на файлы актуального тега):

- [Linux x86_64](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64)
- [Linux x86](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-386)
- [Linux ARM64](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-arm64)
- [Windows x64](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe)
- [Windows x86](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-386.exe)
- [macOS Intel](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64)
- [macOS Apple Silicon](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64)
- [SHA256SUMS](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/SHA256SUMS)

Шаблон для скриптов: `https://github.com/Palymer/yandex-speed-cli/releases/latest/download/<имя_файла>`.

### Быстрый тест бинарника с GitHub (`curl` / `wget`)

Ниже — примеры для **Linux x86_64** (`yandex-speed-cli-linux-amd64`). Для другой платформы замените имя файла из таблицы выше (например `yandex-speed-cli-linux-arm64`, `yandex-speed-cli-darwin-arm64`).

**curl** — скачать во временный каталог, выставить права на запуск и сразу короткий тест без ожидания Enter и без запроса геолокации:

```bash
BASE="https://github.com/Palymer/yandex-speed-cli/releases/latest/download"
curl -fL "$BASE/yandex-speed-cli-linux-amd64" -o ./yandex-speed-cli
chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

**wget**:

```bash
BASE="https://github.com/Palymer/yandex-speed-cli/releases/latest/download"
wget -O ./yandex-speed-cli "$BASE/yandex-speed-cli-linux-amd64"
chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

Флаг `-f` у `curl` и проверка кода выхода помогают заметить ошибку (например, если релиза ещё нет или имя артефакта опечатано). Пока на GitHub не появился ни один релиз, эти URL вернут 404 — тогда соберите утилиту из исходников (раздел **«Самостоятельная сборка из исходников»** ниже) или установите через `go install`.

Скачайте файл под свою систему вручную, при необходимости проверьте суммы (`sha256sum -c SHA256SUMS` на Linux) и положите бинарник в каталог из `PATH`.

### Через Go

```bash
go install github.com/Palymer/yandex-speed-cli@latest
```

Бинарник окажется в `$(go env GOPATH)/bin` — этот каталог должен быть в `PATH`.

### Самостоятельная сборка из исходников

Нужны **Go 1.21+** и **Git**. Сборка без внешних зависимостей (только стандартная библиотека), `CGO_ENABLED=0` для статического бинарника под текущую ОС/архитектуру не обязателен, но удобен для переносимости.

```bash
git clone https://github.com/Palymer/yandex-speed-cli.git
cd yandex-speed-cli
go build -trimpath -ldflags="-s -w" -o yandex-speed-cli .
./yandex-speed-cli -quick -no-wait -no-geo
```

Версия в бинарнике по умолчанию — `dev`. Чтобы подставить номер (как в релизах), при сборке укажите `-ldflags`:

```bash
go build -trimpath -ldflags="-s -w -X main.version=v1.0.0" -o yandex-speed-cli .
./yandex-speed-cli -version
```

**Кросс-компиляция** (все поддерживаемые комбинации `GOOS`/`GOARCH`, артефакты в `dist/`):

```bash
VERSION=v1.0.0 make all
```

Альтернатива без `make`: на Unix `./build.sh`, на Windows PowerShell `.\build.ps1` (необязательная переменная окружения `VERSION` — строка версии для `-X main.version=…`).

После `make all` или скриптов проверьте, например: `dist/yandex-speed-cli-linux-amd64 -version`.

## Быстрый старт

Интерактивный прогон с ожиданием Enter в конце:

```bash
yandex-speed-cli -quick
```

Для скриптов — без паузы и с JSON:

```bash
yandex-speed-cli -quick -no-wait -no-geo -json
```

## Флаги

| Флаг | Описание |
|------|----------|
| `-version` | Напечатать версию сборки и выйти |
| `-quick` | Короткий тест (~4 с на канал, меньше пингов) |
| `-duration` | Длительность download/upload в секундах (по умолчанию 10) |
| `-workers` | Число параллельных потоков на фазу |
| `-ping` | Число замеров задержки после выбора узла |
| `-no-download` / `-no-upload` | Пропуск соответствующей фазы |
| `-json` | Вывод результата в JSON (без анимации прогресса) |
| `-no-color` | Отключить ANSI-цвета |
| `-no-wait` | Не ждать Enter после теста (удобно для CI и пайпов) |
| `-no-geo` | Не обращаться к ipwho.is |

Справка стандартной библиотеки: `yandex-speed-cli -h`.

## JSON и автоматизация

В режиме `-json` в stdout уходит один объект с полями приложения, сессии (локальные дата/время и смещение от UTC, IPv4/IPv6 из логики классификации публичного адреса, регионы при наличии геолокации), задержки и при успешных фазах — download/upload (средняя, пик, p95/p50 мгновенной скорости, байты, длительность).

Дата и время в человекочитаемой панели берутся из **локальных** часов машины. Поля региона по данным ipwho.is **не гарантируют** полное совпадение с виджетом на сайте Яндекса.

## Релизы и CI

- **CI** (`.github/workflows/ci.yml`) — на push и pull request в ветки `main` / `master`: `gofmt`, `go vet`, `go test`, нативная сборка; затем на каждой ОС скачивается бинарник из [**последнего GitHub Release**](https://github.com/Palymer/yandex-speed-cli/releases/latest) (`linux-amd64` / `windows-amd64.exe` / `darwin-arm64`). Если артефакт или API ещё недоступны (первый деплой, задержка CDN), шаг **пропускается** с заметкой в логе; при HTTP 200 версия `./… -version` сравнивается с тегом релиза.
- **Release** (`.github/workflows/release.yml`) — **автоматически при каждом push в `main` или `master`** (и вручную через **Actions → Release → Run workflow**): `VERSION` → тег `v…`, те же проверки `gofmt` / `vet` / `test`, кросс-сборка `make all`, проверка `-version` у `dist/yandex-speed-cli-linux-amd64`, `SHA256SUMS`, публикация или обновление GitHub Release.
- **Dependabot** (`.github/dependabot.yml`) — еженедельная проверка обновлений **GitHub Actions**.

Если тег `v<VERSION>` уже существует на **другом** коммите, workflow завершится с ошибкой: нужно увеличить число в **`VERSION`**, закоммитить и снова запушить в `main`.

Ручной тег для релиза больше не обязателен — достаточно пуша в `main` после изменения кода при актуальной версии в `VERSION`.

Имена файлов и архитектуры: **amd64** — то, что часто называют x64; **386** — x86; **arm64** — AArch64 (в т.ч. Apple Silicon). 32-битный macOS (`darwin/386`) в компиляторе Go не поддерживается.

## Структура репозитория

| Путь | Назначение |
|------|------------|
| `main.go` | HTTP-клиент, замеры, флаги, JSON |
| `display.go` | Баннер, тема, прогресс, сводки |
| `geo.go` | Геолокация по IP |
| `vt_windows.go` / `vt_other.go` | Включение VT на Windows / заглушка на остальных ОС |
| `VERSION` | Версия для автоматического релиза (`1.0.0` → тег `v1.0.0`) |
| `.github/dependabot.yml` | Обновления зависимостей GitHub Actions |
| `Makefile`, `build.sh`, `build.ps1` | Кросс-сборка в `dist/` |

## Лицензия

[MIT](LICENSE)
