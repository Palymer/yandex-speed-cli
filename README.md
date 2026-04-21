# yandex-speed-cli

Замер скорости через API [Яндекс.Интернетометра](https://yandex.ru/internet): IP, задержка, download/upload, опционально JSON. Неофициальный клиент; см. [ограничения](https://yandex.ru/legal/internet_termsofuse/) сервисов Яндекса при использовании.

**Сборка из исходников и кросс-компиляция:** [README.build.md](README.build.md).

## Скачать (последний релиз)

Страница релиза: [**latest**](https://github.com/Palymer/yandex-speed-cli/releases/latest) · [все релизы](https://github.com/Palymer/yandex-speed-cli/releases).

| ОС / архитектура | Прямая ссылка (latest) |
|------------------|------------------------|
| Linux x86_64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64) |
| Linux x86 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-386) |
| Linux ARM64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-arm64) |
| Windows x64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe) |
| Windows x86 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-386.exe) |
| macOS Intel | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64) |
| macOS Apple Silicon | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64) |
| Контрольные суммы | [SHA256SUMS](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/SHA256SUMS) |

Шаблон: `https://github.com/Palymer/yandex-speed-cli/releases/latest/download/<имя_файла>`.

Установка через Go: `go install github.com/Palymer/yandex-speed-cli@latest`

## Тесты

### Локально (без сети к Яндексу для компиляции)

```bash
gofmt -l .          # должно быть пусто
go vet ./...
go test ./... -count=1
go build -trimpath -ldflags="-s -w" .
```

### Локальный прогон замера (нужен интернет)

Короткий сценарий без паузы и без геолокации:

```bash
./yandex-speed-cli -quick -no-wait -no-geo
```

Машиночитаемый вывод: `./yandex-speed-cli -quick -no-wait -no-geo -json`.

### GitHub Actions

| Workflow | Когда | Что проверяется |
|----------|--------|-----------------|
| [CI](.github/workflows/ci.yml) | push / PR в `main`, `master` | `gofmt`, `go vet`, `go test`, сборка из исходников; затем скачивание бинарника из [latest release](https://github.com/Palymer/yandex-speed-cli/releases/latest) под ОС раннера и совпадение `./… -version` с тегом релиза (API). Если релиза ещё нет или CDN отдал не 200 — шаг smoke **пропускается** с `::notice::` в логе. |
| [Release](.github/workflows/release.yml) | push в `main` / `master`, вручную | `VERSION` → тег `v…`, `gofmt` / `vet` / `test`, `make all`, проверка `-version` у `dist/yandex-speed-cli-linux-amd64`, `SHA256SUMS`, публикация [релиза](https://github.com/Palymer/yandex-speed-cli/releases/latest). |

### Быстрый smoke бинарника с GitHub (curl / wget)

Пример для **Linux x86_64** (имя файла см. в таблице выше):

```bash
BASE="https://github.com/Palymer/yandex-speed-cli/releases/latest/download"
curl -fL "$BASE/yandex-speed-cli-linux-amd64" -o ./yandex-speed-cli && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

```bash
BASE="https://github.com/Palymer/yandex-speed-cli/releases/latest/download"
wget -O ./yandex-speed-cli "$BASE/yandex-speed-cli-linux-amd64" && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

## Флаги (для сценариев тестов)

| Флаг | Назначение |
|------|------------|
| `-version` | Версия сборки |
| `-quick` | Короткий замер |
| `-duration`, `-workers`, `-ping` | Длительность, потоки, число пингов |
| `-no-download` / `-no-upload` | Отключить фазу |
| `-json` | Вывод JSON |
| `-no-color` | Без ANSI |
| `-no-wait` | Без ожидания Enter |
| `-no-geo` | Без ipwho.is |

`yandex-speed-cli -h` — полный список.

## Лицензия

[MIT](LICENSE)
