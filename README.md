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

## Запуск (стандарт — быстрый тест)

Геолокация по IP (ipwho.is) **включена по умолчанию** — `-no-geo` только если регион не нужен.

Рекомендуемый режим: **быстрый** замер (короткие фазы), без паузы «нажмите Enter» в конце:

```bash
yandex-speed-cli -quick -no-wait
```

Вывод JSON:

```bash
yandex-speed-cli -quick -no-wait -json
```

## Замер 30 секунд

Фазы download и upload длятся столько секунд, сколько задано в `-duration`. Значение **30** задаётся так:

1. **Не используйте** `-quick` — он включает короткий тест и перекрывает `-duration`.
2. Укажите **`-duration 30`** (секунды — число с плавающей точкой, можно `30` или `30.0`).

```bash
yandex-speed-cli -duration 30 -no-wait
```

Если не задавать ни `-quick`, ни `-duration`, длительность DL/UL по умолчанию **10** секунд.

## Загрузка с GitHub и быстрый тест

Подставьте URL из таблицы выше. Файл сохраняйте как `ysc` / `ysc.exe`: **curl** — `-o`, **wget** — `-O`.

**Linux x86_64:**

| Инструмент | Одна строка (скачать → права → **быстрый** замер) |
|------------|---------------------------------------------------|
| `curl` | `curl -fL "$URL" -o ysc && chmod +x ysc && ./ysc -quick -no-wait` |
| `wget` | `wget -O ysc "$URL" && chmod +x ysc && ./ysc -quick -no-wait` |

Пример с полной ссылкой:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-linux-amd64" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
```

**macOS Intel** — `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-amd64" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
```

**macOS Apple Silicon** — `curl`:

```bash
curl -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-darwin-arm64" -o ysc && chmod +x ysc && ./ysc -quick -no-wait
```

**Windows x64:**

```powershell
iwr "URL" -OutFile ysc.exe; .\ysc.exe -quick -no-wait
# или
curl.exe -fL "URL" -o ysc.exe; .\ysc.exe -quick -no-wait
```

```powershell
curl.exe -fL "https://github.com/Palymer/yandex-speed-cli/releases/latest/download/yandex-speed-cli-windows-amd64.exe" -o ysc.exe; .\ysc.exe -quick -no-wait
```

Для **Windows x86** в URL используйте `yandex-speed-cli-windows-386.exe`.

### После загрузки: замер 30 секунд

Тот же бинарник (`./ysc` / `.\ysc.exe`), но **без** `-quick` и с **`-duration 30`**:

```bash
./ysc -duration 30 -no-wait
```

```powershell
.\ysc.exe -duration 30 -no-wait
```

## Флаги

| Флаг | Назначение |
|------|------------|
| `-version` | Версия сборки |
| `-quick` | Короткий замер (перекрывает `-duration`) |
| `-duration` | Длительность DL/UL в секундах (по умолчанию **10**, если не задан `-quick`) |
| `-workers`, `-ping` | Потоки и число пингов |
| `-no-download` / `-no-upload` | Отключить фазу |
| `-json` | Вывод JSON |
| `-no-color` | Без ANSI |
| `-no-wait` | Без ожидания Enter в конце |
| `-no-geo` | Не запрашивать геолокацию (ipwho.is) |

Полный список: `yandex-speed-cli -h`.

## Лицензия

[MIT](LICENSE)
