# Cистема мониторинга метрик и телеметрии Prometheus + Grafana порка ТС.

Содержание:
* [Скрины](#скрины)
* [Архитектура инфраструктуры мониторинга](#архитектура-инфраструктуры-мониторинга)
* [docker compose content](#docker-compose-content)
* [Уведомления](#уведомления)
* [Описание проекта](#описание-проекта)
* [Походные записи](#походные-записи)

    * [Формирование пароля для MQTT brokr mosquitto](#формирование-пароля-для-mqtt-brokr-mosquitto)
    * [Создание пароля для закрытых именно под авторизацией nginx роутов](#создание-пароля-для-закрытых-именно-под-авторизацией-nginx-роутов)
    * [Подход Dashboard-as-Code for Grafana (в графане называют Provisioning)](#подход-dashboard-as-code-for-grafana-в-графане-называют-provisioning)
    * [Оцениваю ресурсы для ВМ на компоненты инфраструктуры мониторинга (на 100 активных ТС)](#-оцениваю-ресурсы-для-вм-на-компоненты-инфраструктуры-мониторинга-на-100-активных-тс)
    * [Источники](#источники-не-все)

## Скрины
![](./assets/tractor_details_dash.png)
![](./assets/park_state.png)

## Архитектура инфраструктуры мониторинга


<details open>
<summary> свернуть</summary>

```mermaid
block
columns 1
  user(("👤<br>Диспетчер"))
  nginx[["🔒 Nginx Proxy<br>80 / 443<br>Basic Auth / SSL"]]
  user --> nginx

  block:grafana
  columns 2
    block:gr
    columns 2
      g["Grafana"]
      gright<["Grafana's built-in<br>AlertManager"]>(right)
    end
    block:agents
    columns 1
      tg["Telegram"] email["Email<br>SMTP"]
    end
  end
  nginx --> grafana

  block:arrows_st
  columns 2
    downsql<["SQL<br>requests"]>(down) 
    downpql<["PromQL<br>requests"]>(down) 
  end

  block:storage
  columns 2
    pg[("PostgreSQL")]
    vm[("Victoriametrics")]
  end
 

  block:arrows_gopr
  columns 2
    upsql<["SQL<br>inserts<br>events"]>(up) 
    space 
  end

  block:gopr
  columns 2
    backend
    ps["Park Simulator"]
  end

  block:arrows_gopr2
  columns 2
    subscribed<["MQTT<br>subscribed"]>(up) 
    tomqttbroker<["MQTT<br>deliver"]>(down)
  end

  mqttbroker["MQTT<br>Broker<br>mosquitto"]

  block:mqtt_exp
  columns 3
    space
    frommqttbroker<["MQTT<br>subscribed"]>(down)
    ns(("Node system"))
  end

  block:exporters
  columns 3
  postgres_exporter["Postgres Exporter<br>gets data from system"]
  mqtt_exporter["MQTT Exporter<br>mqtt2prometheus"]
  node_exporter["Node Exporter<br>gets data from system"]
  end
  postgres_exporter -->pg

  block:promql
  columns 3
    promql_pg<["GET /metrics"]>(up)
    promql_mqtt<["GET /metrics"]>(up)
    promql_node<["GET /metrics"]>(up)
  end
  prometheus -->vm

  classDef nginx_style fill:#696,stroke:#333,stroke-width:0px,color:#fff;
  class nginx nginx_style

  classDef orange fill:#dd5522,stroke:#333,stroke-width:0px,color:#fff;
  class g,gright orange

  classDef lorange fill:#ff8855,stroke:#333,stroke-width:0px,color:#fff;
  class gr lorange
  
  classDef blue fill:#5888f5,stroke:#fff,color:#fff;
  class backend,ps blue
  
  classDef tgblue fill:#8899ff,stroke:#fff,color:#fff;
  class tg tgblue

  classDef deepblue fill:#6666ff,stroke:#fff,color:#ff0000,stroke-width:0px;
  class mqttbroker,mqtt_exporter,frommqttbroker,subscribed,tomqttbroker deepblue

  classDef red fill:#aa6666,stroke:#fff,color:#fff,stroke-width:0px;
  class prometheus,vm,downpql,promql_pg,promql_mqtt,promql_node red

  classDef lightblue fill:#58bbff,stroke:#fff,color:#fff;
  class postgres_exporter,pg,upsql,downsql lightblue

  classDef green fill:#55aa88,stroke:#fff,color:#fff;
  class node_exporter,ns green
```

</details>

## docker compose content

<details>
<summary>compose.yml (развернуть)
</summary>

```yaml
networks:
  monitoring:
    driver: bridge
volumes:
  mqtt2prometheus_cache:
  prometheus_data:
  victoria_metrics_data:
  grafana_data:
  postgres_data:

services:
  park-simulator:
    image: ghcr.io/aykuli/armmeh_promgraf_telemetry/simulator:latest
    container_name: park_simulator
    networks:
      - monitoring
    environment:
      - MQTT_USER=${MQTT_USER}
      - MQTT_PASS=${MQTT_PASS}
      - MQTT_BROKER_URL=${MQTT_BROKER_URL}
  backend:
    image: ghcr.io/aykuli/armmeh_promgraf_telemetry/backend:latest
    container_name: backend
    networks:
      - monitoring
    environment:
      - MQTT_USER=${MQTT_USER}
      - MQTT_PASS=${MQTT_PASS}
      - MQTT_BROKER_URL=${MQTT_BROKER_URL}
      - FLEET_BACKEND_URL=${FLEET_BACKEND_URL}
      - POSTGRES_DSN=${POSTGRES_DSN}

  # --- БРОКЕР ОЧЕРЕДИ СООБЩЕНИЙ ---
  mosquitto:
    image: eclipse-mosquitto:2.1.2-alpine
    container_name: mosquitto
    volumes:
      - ./mosquitto/config:/mosquitto/config
      - ./mosquitto/log:/mosquitto/log
    ports:
      - '1883:1883'
    networks:
      - monitoring
    restart: unless-stopped

  # --- MQTT -> PROMETHEUS ---
  mqtt-exporter:
    image: ghcr.io/hikhvar/mqtt2prometheus:v0.1.8-RC2
    container_name: mqtt_exporter
    volumes:
      - ./mqtt2prometheus/config:/config
      - mqtt2prometheus_cache:/var/lib/mqtt2prometheus
    command:
      - '-config=/config/config.yml'
    ports:
      - '9641:9641'
    entrypoint:
      - /mqtt2prometheus
      - -log-level=debug
    networks:
      - monitoring
    environment:
      - MQTT2PROM_MQTT_USER=${MQTT_USER}
      - MQTT2PROM_MQTT_PASSWORD=${MQTT_PASS}
    depends_on:
      - mosquitto

  # --- NODE DATA -> PROMETHEUS ---
  node-exporter:
    image: prom/node-exporter:master
    container_name: node_exporter
    pid: host          # Without pid: host, Node Exporter cannot inspect processes running directly on your host machine
    network_mode: host # Gives Node Exporter access to real host interfaces (eth0, wlan0, etc.)
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.ignored-mount-points=^/(sys|proc|dev|host|etc)($$|/)'
    restart: unless-stopped

  # --- МОНИТОРИНГ И СБОР МЕТРИК ---
  prometheus:
    image: prom/prometheus:v3.13.1
    container_name: prometheus
    volumes:
      - ./prometheus:/etc/prometheus
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--storage.tsdb.retention.time=15d'
      - '--storage.tsdb.retention.size=10GB'
    networks:
      - monitoring
    extra_hosts: ["host.docker.internal:host-gateway"] # for node-exporter service

  # --- ДОЛГОСРОЧНОЕ ХРАНИЛИЩЕ МЕТРИК ---
  victoriametrics:
    image: victoriametrics/victoria-metrics:v1.101.0
    container_name: victoria_metrics
    volumes:
      - victoria_metrics_data:/victoria-metrics-data
    networks:
      - monitoring
    restart: unless-stopped

  postgres-db:
    image: postgres:17-alpine
    container_name: postgres_db
    ports:
      - "5432:5432"
    environment:
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - monitoring
    restart: unless-stopped


  # --- МОНИТОРИНГ БАЗЫ ДАННЫХ ПОТСГРЕС ---
  postgres-exporter:
    image: quay.io/prometheuscommunity/postgres-exporter:latest
    container_name: postgres_exporter
    network_mode: host
    environment:
      - DATA_SOURCE_URI=${POSTGRES_URI}
      - DATA_SOURCE_USER=${POSTGRES_USER}
      - DATA_SOURCE_PASS=${POSTGRES_PASSWORD}
    restart: unless-stopped

  # --- ВИЗУАЛИЗАЦИЯ И КАРТЫ ---
  grafana:
    image: grafana/grafana:11.6.15-ubuntu
    container_name: grafana
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/etc/grafana/dashboards
      - ./grafana/config/grafana.ini:/etc/grafana/config/grafana.ini:ro
    extra_hosts:
      - "host.docker.internal:host-gateway"
    networks:
      - monitoring
    environment:
      - GF_SECURITY_ADMIN_USER=${GRAFANA_ADMIN_NAME}
      - GF_SECURITY_ADMIN_PASSWORD=${GRAFANA_ADMIN_PASSWORD}
      - GF_ALERTS_ALLOW_LOCAL_MODE_NOTIFIER=true
      - GF_SMTP_ENABLED=true
      - GF_SMTP_HOST=smtp.yandex.ru:465
      - GF_SMTP_USER=${GF_SMTP_USER}
      - GF_SMTP_PASSWORD=${GF_SMTP_PASSWORD}
      - GF_SMTP_FROM_ADDRESS=${GF_SMTP_FROM_ADDRESS}
      - GF_SMTP_FROM_NAME=${GF_SMTP_FROM_NAME}
      - GF_SMTP_SKIP_VERIFY=false
      - BOT_TOKEN=${BOT_TOKEN}
      - TELEGRAM_CHAT_ID=${TELEGRAM_CHAT_ID}
      - POSTGRES_USER=${POSTGRES_USER}
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}
      - POSTGRES_DB=${POSTGRES_DB}
    depends_on:
      - prometheus
      - victoriametrics
    restart: unless-stopped
  nginx:
    image: nginx:latest
    container_name: nginx
    ports:
      - 80:80
      - 443:443
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/conf.d/default.conf:ro
      - ./nginx/.htpasswd:/etc/nginx/.htpasswd:ro 
      - ./certbot/conf:/etc/letsencrypt:ro
      - ./certbot/www:/var/www/certbot
      - /var/run/docker.sock:/var/run/docker.sock # Чтобы контейнер certbot мог выполнить команду docker kill внутри себя и дотянуться до соседнего контейнера Nginx
    networks:
      - monitoring
    restart: always


```
</details>

## Уведомления

📱 Уведомления приходят на https://t.me/grafana_alerts_armmeh и на мою почту.
Для почты в YandexID -> Приложения -> Создать пароль. Этот пароль и будет в перменной окржения `GF_SMTP_PASSWORD`.


## Описание проекта
Проект системы сбора, хранения и визуализации метрик для мониторинга парка транспортных средств (ТС) в реальном времени включает в себя:
* [`compose.yml`](./compose.yml) - docker compose,который враз поднимает на сервере (я сделала на одном для экономии собственных средств, так как сервер я подняла на ЯО)
* [terraform проект](./terra/main.tf) для разворачивания виртуальной машины:
    * [terrafrom apply github action](./.github/workflows/terraaply.yml):
        * кроме создания самой ВМ, через секреты github-actions и [`cloud-init.yml`](./terra/cloud-init.yml) в ВМ передаются креды, переменные, необходимые для работы всей инфраструктуры, когда я разворачиваю через [compose.yml](./compose.yml)
    * [terrafrom destroy github action](./.github/workflows/terradestroy.yml) - это я создала для того, чтобы удалить ВМ, даже с телефона. Скоро уеду на недельку, и не хочу брать с собой ноут.
    * *заметьте* - `actions` запускаются по нажатии на кнопку 'run workflow'
* [deploy github actions](./.github/workflows/deploy.yml) - собирает нужные образы моих симуляторов парка и бекенда, публикует в github registry образы, которые потом на сервере будут скачаны/обновлены, когда запуститься [`compose.yml`](./compose.yml).
* [golang проект симуляции парка ТС](./go-projects-group/cmd/vehicle_park_simulator/main.go),который вещает на MQTT брокер в топик `<vehicle_type>/<fuel_type>/<vehicle_id>/telemetry`
* [golang проект, имитирующий бекенд](./go-projects-group/cmd/fleet_backend/main.go),который сохраняет события в базе данных PostgreSQL, приходящие с подписки на MQTT брокер в топик `<vehicle_type>/<fuel_type>/<vehicle_id>/telemetry`. Там, конечно, всё тело меткрики приходит из симуляции, но устала я уж всё делать и решила ограничиться сохранением событий в бд.

* [настройки для MQTT брокера `mosquitto`](./mosquitto/config/mosquitto.conf)
* [настройки для MQTT expoter `mqtt2prometheus`](./mqtt2prometheus/config/config.yml) - здесь я придумала названия для метрик соответствеено путям в json теле приходящего сообщения из mqtt брокера.
* [настройки для `prometheus`](./prometheus/prometheus.yml)
    * здесь прописаны все экспортеры, упомянутые в задаче - node_exporter postgres_exporter, fms_backend, vehicles и дефолтный самого prometheus.
    * обработку метрик, приходящие с mqtt_exporter, (`job: fms_backend`) я описываю тут более детально, чем для дргуих экспортеров. Так как для других экспортеров я использую уже придуманные метрики, к которым уже и есть соотвественно дашборды для графаны, котрые достпуны в свободном скачивании.
    * есть путь, куда записывать данные. Я решила попробовать здесь `victoriametrics`. Потому что - а почему бы и нет.
* [настройки для `grafana`](./prometheus/prometheus.yml):
    * здесь реализован прицип CaaC - `configure as code` в целях поддрежки масштабирования системы мониторинга.
    * Необходимо прописать в конфигурациях:
        1) [./grafana/config/grafana.ini](./grafana/config/grafana.ini) активировать протокол SMTP
        2) [./grafana/provisioning/datasources/datasources.yml](./grafana/provisioning/datasources/datasources.yml) - прописать, откуда брать данные для отображения различных метрик
        3) [./grafana/provisioning/dashboards/dashboards.yml](./grafana/provisioning/dashboards/dashboards.yml) - общая конфигурация дашбордов
        4) [./](./grafana/provisioning/alerting/alerting.yml) - список contact points, куда отправляются алёрты, и правила алёртов с привязкой в панелям в дашбордах. о как я тут много времени просидела же.
        4) [./grafana/dashboards/*.json](./grafana/dashboards) - сами дашборды собственно. названия дашбордов говорящие сами за себя, как говорится:
            * `node_exporter_full.json` - скачала с маркета графаны
            * `postgres_exporter.json` - тоже скачана с маркета графаны в соответствии с именно тем экспортером `postgres_exporter` от графана коммюнити, который я использовала.
            * `park_state.json`, `tractor_details.json`, `events.json` - сама сделала, так как они уникальные в рамках этого проекта. Сначала формировала в UI Grafana, затем эксопртировала json и сохранила.
* [настройки прокси сервера на nginx](./nginx/nginx.conf) на адрес [mymeddata.ru](https://mymeddata.ru/) (это название ничего не значит, просто валялся у меня в сусеках) настроен адрес главной страницы Grafana.


## Походные записи

### Формирование пароля для MQTT brokr mosquitto

```
$ sudo chown 1883:1883 mosquitto/config/password.txt
$ docker exec -it mosquitto mosquitto_passwd -b /mosquitto/config/password.txt <mqtt-user-name> <my-super-power-password>
```

👍 Искреннее уважение к разработчикам этих инструментов за понятные тексты ошибок. А оОшибалась я много.

## Создание пароля для закрытых именно под авторизацией nginx роутов

Роуты см в [./nginx/nginx.conf](./nginx/nginx.conf)

```
docker run --rm -ti alpine sh -c "apk add --no-cache apache2-utils && htpasswd -nb admin secret_password"
```


##  Подход Dashboard-as-Code for Grafana (в графане называют Provisioning).

```mermaid
treeView-beta
"grafana"
    "provisioning"
        "datasources"
            "datasources.yaml"
        "dashboards"
            "dashboards.yaml"
    "dashboards"
        "vehicles.json"
```


## 📊 Оцениваю ресурсы для ВМ на компоненты инфраструктуры мониторинга (на 100 активных ТС)

| Компонент | Операции в секунду (RPS / IOPS) | Нагрузка на CPU / ОЗУ | Требования к диску (За сутки) | Критическое место (Bottleneck) |
| :--- | :--- | :--- | :--- | :--- |
| **Mosquitto** | ~10–20 msg/sec | Минимальная (<1% CPU / ~15MB RAM) | 0 MB (Хранит только в памяти) | Пропускная способность сети при росте числа ТС |
| **Go Backend** | ~10–20 parsing/sec | Низкая (~2% CPU / ~40MB RAM) | Только логи контейнера (~10MB) | Скорость пул-соединений (Connection Pool) к Postgres |
| **Postgres DB** | ~1–2 write IOPS (только аномалии) | Низкая (~2% CPU / ~120MB RAM) | ~5–10 MB / сутки | Отсутствие индексов при росте таблицы логов событий, но исправимо |
| **MQTT Exporter** | ~0.06 RPS (1 запрос от Prometheus в 15с) | Минимальная (<0.5% CPU / ~25MB RAM) | 0 MB | Накопление невалидных топиков в оперативной памяти |
| **Prometheus v3** | ~0.25 RPS (опрос 4 экспортеров в 15с) | Средняя (~5% CPU / ~250MB RAM) | ~15–20 MB / сутки (без VM) | Потребление ОЗУ при удержании метрик в кэше (TSDB Head) |
| **VictoriaMetrics** | ~1 write RPS (сжатый поток от Prom) | Низкая (~1% CPU / ~60MB RAM) | **~2–4 MB / сутки** (архивное сжатие) | Права на чтение/запись (I/O) в Docker Volume на диске SSD |
| **Nginx Proxy** | Нагрузка только при запросах диспетчера | Минимальная (<0.5% CPU / ~10MB RAM) | Только логи доступа `access.log` | Объем ОЗУ при обработке тяжелых JSON-выгрузок |
| **Grafana v11** | 1 запрос к DB каждые 5с (автообновление) | Средняя (~3% CPU / ~150MB RAM) | ~1 MB (состояние сессий) | Рендеринг тяжелых графиков в браузере у диспетчера |

1. Расчет сетевого потока (RPS / IOPS)
Для симуляции 45 активных транспортных средств (ТС):

* **Входной поток (Mosquitto & Go Backend):** Каждый симулятор шлет пакет телеметрии в раз в 15 секунд.

\(\text{45\ ТС}/\text{15\ секунд}\approx \text{3\ сообщений\ в\ секунду\ (RPS)}.\)

* **Запись в Postgres DB**: По коду симулятора аномалии (события high_speed, low_fuel и т.д.) генерируются по условию vehicleInfo.ID % 4 или % 5. Это значит, что критические статусы содержатся примерно в каждом 4–5 пакете, а в остальных идет штатный режим без отправки массива Events.

\(\text{3\ RPS}\times 20\%\text{\ аномалий}\approx \text{1\ операции\ записи\ (IOPS)\ в\ секунду}.\)

* **Запросы к Prometheus & Exporter**: Сборщик метрик работает по Pull-модели с фиксированным интервалом 15 секунд. Он делает 1 HTTP-запрос к каждому экспортеру (их 4: backend, mqtt-exporter, node-exporter, postgres-exporter) раз в 15 секунд.

\(\text{4\ запроса}/\text{15\ секунд}=\text{0.26\ RPS}.\)

2. **Расчет объема диска (Disk Space Allocation)VictoriaMetrics**: Одна сырая метрика (Time Series) в Prometheus занимает около 1–2 байт на хосте. 
При 45 машинах и опросе раз в 15 секунд генерируется небольшой объем данных. 
* **VictoriaMetrics**: использует мощное блочное сжатие (ZSTD/Chimp), упаковывая данные в архивы. На практике 45 ТС создают поток не более 1-2 МБ сжатых логов за 24 часа.

* **Postgres DB**: Каждая строка инцидента в таблице events (состоящая из INT, VARCHAR и TEXT лога на 10 символов) весит в среднем 100–150 байт.

\(\text{1\ запись/сек}\times \text{86400\ сек/сутки}\times \text{150\ байт}\approx \text{13\ МБ\ сырых\ данных}.\)

3. **Оценка CPU и ОЗУ (Профили контейнеров)**

* **Mosquitto & MQTT Exporter**: Написаны на C и Go соответственно. Обладают очень малым потреблением. Потока в 3 RPS недостаточно, чтобы нагрузить процессор даже на 1%. Память расходуется только на удержание дескрипторов сетевых сокетов и буферизацию топиков.
* **Go Backend**: Скомпилированный бинарник Go (который я оптимизировала с помощью multi-stage сборки в Dockerfile) работает на нативных потоках ОС (Goroutines). Парсинг JSON для 3 пакетов в секунду занимает около 1 миллисекунд процессорного времени ядра.
* **Grafana**: Сама Grafana потребляет память на кэширование сессий пользователя и рендеринг дашбордов. Запрос автообновления 5s запускает легковесные SQL/PromQL селекторы. Нагрузка возрастает только в момент, когда диспетчер открывает браузер и Grafana начинает отрисовывать графики (SVG/Canvas) на экране клиента.

## Источники (не все)

* [node-exporter-full/](https://grafana.com/grafana/dashboards/1860-node-exporter-full/)

* https://prometheus.io/docs/instrumenting/exporters/

* https://github.com/hikhvar/mqtt2prometheus
* https://github.com/prometheus-community/postgres_exporter

* https://hub.docker.com/r/prom/prometheus/tags
* https://grafana.com/blog/how-to-integrate-grafana-alerting-and-telegram/ 
