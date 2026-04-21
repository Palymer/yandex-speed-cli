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

Геолокация по IP (ipwho.is) **включена по умолчанию** — флаг `-no-geo` только если регион не нужен.

Быстрый замер (~4 с на канал), без паузы в конце:

```bash
yandex-speed-cli -quick -no-wait
```

JSON:

```bash
yandex-speed-cli -quick -no-wait -json
```

**Длительный замер** (например **30 с** на download/upload) — без `-quick` (он задаёт свою короткую длительность), явно укажите `-duration`:

```bash
yandex-speed-cli -duration 30 -no-wait
```

По умолчанию без `-duration` фазы DL/UL длятся **10** секунд.

## Загрузка с GitHub и проверка

Подставьте URL из таблицы выше. Ниже короткое имя файла: `ysc` / `ysc.exe`; **curl** пишет файл флагом `-o`, **wget** — `-O`.

**Linux x86_64** — `curl -o` *или* `wget -O` (одна строка: скачать → права → запуск):

```bash
curl -fL "$URL" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
# или
wget -O ysc "$URL" && chmod +x ysc && ./ysc -quick -no-wait
```

С полной ссылкой и **30 с** на DL/UL:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64" -o ysc && chmod +x ysc && ./ysc -duration 30 -no-wait
```

```bash
wget -O ysc "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64" && chmod +x ysc && ./ysc -duration 30 -no-wait
```

**macOS Intel** — только `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
```

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64" -o ysc && chmod +x ysc && ./ysc -duration 30 -no-wait
```

**macOS Apple Silicon**:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
```

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64" -o ysc && chmod +x ysc && ./ysc -duration 30 -no-wait
```

**Windows x64** — `Invoke-WebRequest -OutFile` *или* `curl.exe -o`:

```powershell
iwr "URL" -OutFile ysc.exe; .\ysc.exe -quick -no-wait
# или
curl.exe -fL "URL" -o ysc.exe; .\ysc.exe -quick -no-wait
```

С полной ссылкой и **30 с**:

```powershell
curl.exe -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe" -o ysc.exe; .\ysc.exe -duration 30 -no-wait
```

```powershell
iwr "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe" -OutFile ysc.exe; .\ysc.exe -duration 30 -no-wait
```

Для **Windows x86** замените в URL имя на `yandex-speed-cli-windows-386.exe`.

## Флаги

| Флаг | Назначение |
|------|------------|
| `-version` | Версия сборки |
| `-quick` | Короткий замер (перекрывает `-duration`) |
| `-duration` | Длительность DL/UL в секундах (по умолчанию 10) |
| `-workers`, `-ping` | Потоки и число пингов |
| `-no-download` / `-no-upload` | Отключить фазу |
| `-json` | Вывод JSON |
| `-no-color` | Без ANSI |
| `-no-wait` | Без ожидания Enter в конце |
| `-no-geo` | Не запрашивать геолокацию (ipwho.is) |

Полный список: `yandex-speed-cli -h`.

## Лицензия

[MIT](LICENSE)
