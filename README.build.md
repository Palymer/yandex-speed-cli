# Сборка yandex-speed-cli из исходников

Требования: **Go 1.21+**, **Git**. Внешних Go-модулей нет (стандартная библиотека).

## Клонирование и нативная сборка

```bash
git clone https://github.com/Palymer/yandex-speed-cli.git
cd yandex-speed-cli
go build -trimpath -ldflags="-s -w" -o yandex-speed-cli .
./yandex-speed-cli -quick -no-wait -no-geo
```

Версия по умолчанию в бинарнике — `dev`. Чтобы зашить номер релиза:

```bash
go build -trimpath -ldflags="-s -w -X main.version=v1.0.0" -o yandex-speed-cli .
./yandex-speed-cli -version
```

## Кросс-компиляция

Артефакты в каталоге `dist/`:

```bash
VERSION=v1.0.0 make all
```

Без **GNU make**: Unix — `./build.sh`, Windows PowerShell — `.\build.ps1` (опционально задайте переменную окружения `VERSION` — строка для `-X main.version=…`).

Проверка, например: `chmod +x dist/yandex-speed-cli-linux-amd64 && dist/yandex-speed-cli-linux-amd64 -version`.

### Имена файлов и архитектуры

| Суффикс в имени | Обычно |
|-----------------|--------|
| `amd64` | x86_64 |
| `386` | x86 |
| `arm64` | AArch64, Apple Silicon |

Собираются: Linux amd64/386/arm64, Windows amd64/386, macOS amd64 и arm64. Порт `darwin/386` в Go удалён.

## Установка без сборки

- Готовые бинарники: см. таблицу ссылок в [README.md](README.md).
- `go install github.com/Palymer/yandex-speed-cli@latest`

## Версия и автоматический релиз на GitHub

В корне репозитория файл **`VERSION`** (строка вида `1.0.0` без префикса `v`) задаёт тег **`v1.0.0`** при workflow Release. Если тег `v…` уже существует на **другом** коммите, нужно увеличить `VERSION`, закоммитить и запушить в `main`.

Автоматические проверки при push: см. `.github/workflows/ci.yml` и `release.yml`.

## Проверка контрольных сумм (Linux)

После скачивания `SHA256SUMS` рядом с бинарниками:

```bash
sha256sum -c SHA256SUMS
```
