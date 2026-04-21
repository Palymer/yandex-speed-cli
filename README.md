# yandex-speed-cli

Замер скорости через API [Яндекс.Интернетометра](https://yandex.ru/internet): IP, задержка, download/upload, опционально JSON. Неофициальный клиент; см. [ограничения](https://yandex.ru/legal/internet_termsofuse/) сервисов Яндекса при использовании.

Сборка из исходников: [README.build.md](README.build.md).

## Скачать (последний релиз)

[**Страница latest**](https://github.com/Palymer/yandex-speed-cli/releases/latest) · [все релизы](https://github.com/Palymer/yandex-speed-cli/releases).

| ОС / архитектура | Ссылка |
|------------------|--------|
| Linux x86_64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64) |
| Linux x86 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-386) |
| Linux ARM64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-arm64) |
| Windows x64 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe) |
| Windows x86 | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-386.exe) |
| macOS Intel | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64) |
| macOS Apple Silicon | [скачать](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64) |
| Контрольные суммы | [SHA256SUMS](https://github.com/Palymer/yandex-speed-cli/releases/latest/download/SHA256SUMS) |

Установка через Go: `go install github.com/Palymer/yandex-speed-cli@latest`

## Запуск

Короткий замер без паузы и без геолокации по IP:

```bash
yandex-speed-cli -quick -no-wait -no-geo
```

JSON в stdout:

```bash
yandex-speed-cli -quick -no-wait -no-geo -json
```

## Загрузка с GitHub и проверка

Подставьте нужный URL из таблицы выше (другая архитектура — другое имя файла в ссылке).

**Linux x86_64** — `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64" -o ./yandex-speed-cli && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

**Linux x86_64** — `wget`:

```bash
wget -O ./yandex-speed-cli "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64" && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

**macOS Intel** — `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64" -o ./yandex-speed-cli && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

**macOS Apple Silicon** — `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64" -o ./yandex-speed-cli && chmod +x ./yandex-speed-cli
./yandex-speed-cli -quick -no-wait -no-geo
```

**Windows x64** — PowerShell:

```powershell
Invoke-WebRequest -Uri "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe" -OutFile ".\yandex-speed-cli.exe"
.\yandex-speed-cli.exe -quick -no-wait -no-geo
```

**Windows x64** — `curl.exe`:

```powershell
curl.exe -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe" -o ".\yandex-speed-cli.exe"
.\yandex-speed-cli.exe -quick -no-wait -no-geo
```

Для **Windows x86** в URL используйте имя файла `yandex-speed-cli-windows-386.exe`.

## Флаги

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

Полный список: `yandex-speed-cli -h`.

## Лицензия

[MIT](LICENSE)
