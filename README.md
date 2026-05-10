# RPI4 Fan Controller

Контроллер охлаждающего вентилятора для Raspberry Pi 4, написанный на Go. Приложение автоматически регулирует скорость вентилятора на основе температуры процессора и системной нагрузки.

## Описание

RPI4 Fan Controller — это легкая и эффективная утилита для управления PWM-вентилятором на Raspberry Pi 4. Приложение:

- Мониторит температуру CPU в реальном времени
- Анализирует системную нагрузку (использование CPU, сетевой трафик, операции дискового ввода-вывода)
- Плавно регулирует скорость вентилятора, избегая резких скачков
- Использует алгоритм сглаживания для стабильной работы
- Запускается в Docker-контейнере

## Требования

- Raspberry Pi 4
- GPIO pin 14, подключенный к PWM управлению вентилятором
- Go 1.26.1 (для компиляции) или Docker

## Установка

### Компиляция

```bash
go build -o rpi4-fan-controller main.go
```

### Docker

#### Сборка и запуск

```bash
docker build -t rpi4-fan-controller .
docker run -d --privileged \
  --restart=always \
  --name rpi4-fan-controller \
  rpi4-fan-controller
```

#### Автоматический запуск после перезагрузки

Параметр `--restart=always` гарантирует, что контейнер будет автоматически перезапущен после перезагрузки системы:

```bash
docker run -d --privileged \
  --restart=always \
  --name rpi4-fan-controller \
  rpi4-fan-controller
```

#### Docker Compose (альтернатива)

Создайте файл `docker-compose.yml`:

```yaml
version: '3.8'

services:
  fan-controller:
    build: .
    image: rpi4-fan-controller
    container_name: rpi4-fan-controller
    privileged: true
    restart: always
    logging:
      driver: "json-file"
      options:
        max-size: "10m"
        max-file: "3"
```

Затем запустите:

```bash
docker-compose up -d
```

Для остановки:

```bash
docker-compose down
```

#### Проверка статуса контейнера

```bash
# Просмотр запущенных контейнеров
docker ps

# Просмотр логов
docker logs rpi4-fan-controller

# Просмотр логов в реальном времени
docker logs -f rpi4-fan-controller

# Остановка контейнера
docker stop rpi4-fan-controller

# Удаление контейнера
docker rm rpi4-fan-controller
```

## Конфигурация

Параметры приложения можно изменить через константы в файле `main.go`:

| Параметр | Значение | Описание |
|----------|----------|---------|
| `GPIO_PIN` | 14 | Номер GPIO пина для подключения вентилятора |
| `PWM_MAX` | 1000 | Максимальное значение PWM |
| `T_MIN` | 45.0°C | Минимальная температура, при которой включается вентилятор |
| `T_FULL` | 65.0°C | Температура, при которой вентилятор работает на полную мощность |
| `S_MIN` | 20 | Минимальная скорость вращения вентилятора (в % от максимума) |
| `CHECK_INT` | 3 | Интервал проверки системных параметров (в секундах) |
| `SMOOTHING` | 0.2 | Коэффициент сглаживания (0.0–1.0), чем меньше значение, тем медленнее меняется скорость |

## Использование

### Запуск приложения

```bash
./rpi4-fan-controller
```

Приложение начнет выводить информацию о работе в консоль и автоматически управлять вентилятором на основе температуры и нагрузки.

### Systemd служба

Для автоматического запуска при загрузке создайте файл `/etc/systemd/system/rpi4-fan-controller.service`:

```ini
[Unit]
Description=RPi4 Fan Controller
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/rpi4-fan-controller
Restart=on-failure
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Затем включите и запустите службу:

```bash
sudo systemctl daemon-reload
sudo systemctl enable rpi4-fan-controller
sudo systemctl start rpi4-fan-controller
```

## Алгоритм управления

1. Приложение считывает текущую температуру CPU
2. Рассчитывает "эффективную температуру" с учетом системной нагрузки
3. На основе эффективной температуры определяет желаемую скорость вентилятора
4. Плавно переводит скорость вентилятора к целевому значению (с учетом коэффициента сглаживания)
5. Отправляет PWM сигнал на GPIO пин
6. Повторяет процесс каждые `CHECK_INT` секунд

## Системные требования

Утилита использует следующие зависимости:

- `gopsutil` — для получения информации о системной нагрузке (CPU, сеть, диск)
- `go-rpio` — для управления GPIO пинами

## [LICENSE](LICENSE)
