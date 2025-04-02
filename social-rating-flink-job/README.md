# Flink Job: Система Расчета Социального Рейтинга

## 1. Описание и Назначение

Этот Apache Flink Job является центральным компонентом обработки данных в системе социального рейтинга. Его основная задача — в режиме реального времени принимать поток событий, связанных с действиями граждан, применять к ним гибкую систему правил скоринга, вычислять и обновлять индивидуальный социальный рейтинг, а также распределять результаты обработки по специализированным базам данных для различных целей:

*   **Оперативный доступ:** Предоставление актуального значения рейтинга.
*   **Исторический анализ:** Хранение полной хронологии изменений рейтинга для каждого гражданина.
*   **Аналитика и Отчетность:** Расчет и сохранение агрегированных показателей по различным срезам (регион, тип события, время и т.д.) для OLAP-запросов.

Система спроектирована для обработки больших объемов данных с низкой задержкой, обеспечивая своевременное обновление рейтинга и аналитических витрин. Динамическое обновление правил скоринга позволяет системе быстро адаптироваться к изменяющимся требованиям и условиям.

## 2. Архитектура и Поток Данных

Система использует событийно-ориентированный подход и разделение ответственности между компонентами.

### 2.1. Источники Данных

*   **Основной Поток Событий (`citizen_events` в Kafka):**
    *   **Назначение:** Центральная шина для всех входящих событий, обеспечивающая буферизацию, отказоустойчивость и возможность параллельной обработки несколькими потребителями.
    *   **Формат:** События в формате JSON, десериализуемые в POJO `com.github.kxrxh.model.CitizenEvent`. Каждое событие содержит ID гражданина, тип и подтип события, временную метку, геолокацию и полезную нагрузку (`payload`) с деталями события.
*   **Вторичные Обновления (`secondary_score_updates` в Kafka):**
    *   **Назначение:** Технический топик для инициирования пересчета рейтинга у связанных граждан. Сюда попадают триггеры `com.github.kxrxh.model.RelatedEffectTrigger`.
    *   **Формат:** JSON, десериализуемый в `RelatedEffectTrigger`, содержащий ID исходного гражданина, ID правила и параметры поиска связей.
*   **Правила Скоринга (Коллекция `scoring_rules` в MongoDB):**
    *   **Назначение:** MongoDB выбрана для хранения правил из-за гибкости её документной модели, что удобно для описания сложных условий и действий правил.
    *   **Механизм:** `com.github.kxrxh.source.MongoRulePollingSource` периодически (интервал настраивается) опрашивает коллекцию `scoring_rules` на наличие активных правил (`active: true`).
    *   **Формат:** Документы, представляемые POJO `com.github.kxrxh.model.ScoringRule`, содержащие условия срабатывания (тип/подтип события), формулу расчета баллов (включая множители), флаг `effectOnRelated` и параметры поиска связанных.

### 2.2. Обработка в Flink

Пайплайн Flink объединяет чтение данных, применение логики и запись результатов.

*   **Чтение и Десериализация:**
    *   `com.github.kxrxh.source.KafkaSourceFactory` создает `KafkaSource` для чтения и десериализации `CitizenEvent` и `RelatedEffectTrigger` из соответствующих топиков Kafka.
*   **Распределение Правил (`BroadcastState`):**
    *   Поток правил, получаемый от `MongoRulePollingSource`, преобразуется в `BroadcastStream`. Это позволяет эффективно доставить *одинаковый* набор актуальных правил на *каждый* параллельный экземпляр основной функции обработки (`StatefulApplyRulesFunction`).
*   **Основная Логика (`StatefulApplyRulesFunction`):**
    *   **Состояние:** Функция является stateful (`KeyedProcessFunction`), используя `citizenId` в качестве ключа.
        *   `ValueState<Double> currentScore`: Хранит *текущее* известное значение рейтинга для конкретного `citizenId`. Инициализируется нулем при первом событии для гражданина.
        *   `ReadOnlyBroadcastState<String, ScoringRule> ruleState`: Предоставляет доступ *только для чтения* к актуальному набору правил, полученному через `BroadcastStream`. Правила хранятся в виде `Map<String, ScoringRule>`, где ключ - это комбинация типа и подтипа события.
    *   **Обработка `CitizenEvent`:**
        1.  Извлекает `citizenId`, `eventType`, `eventSubtype`.
        2.  Находит соответствующее правило в `ruleState`. Если правило не найдено, событие игнорируется (или логируется).
        3.  Получает текущий рейтинг из `currentScore`.
        4.  Вычисляет изменение рейтинга (`scoreChange`) по формуле из правила (включая применение множителей, если они есть).
        5.  Обновляет `currentScore`, добавляя `scoreChange`.
        6.  Создает объект `com.github.kxrxh.model.ScoreSnapshot`, который содержит:
            *   ID гражданина
            *   Предыдущий рейтинг
            *   Новый рейтинг
            *   Изменение рейтинга (`scoreChange`)
            *   ID примененного правила (`triggeringRuleId`)
            *   Детали исходного события (ID, тип, подтип, временная метка, геолокация).
        7.  Отправляет `ScoreSnapshot` в основной выходной поток (`DataStream<ScoreSnapshot>`).
        8.  Если у сработавшего правила `effectOnRelated` равно `true`, создает `RelatedEffectTrigger` с параметрами из правила и отправляет его в `SideOutput` (который далее направляется в Kafka топик `secondary_score_updates`).
*   **Обработка Влияния на Связанных (`AsyncRelatedEffectFunction`):**
    *   **Назначение:** Асинхронно обрабатывает триггеры пересчета для связанных граждан, чтобы не блокировать основной поток обработки событий.
    *   **Логика:**
        1.  Принимает `RelatedEffectTrigger`.
        2.  Используя `com.github.kxrxh.database.dgraph.DgraphRepository`, асинхронно выполняет запрос к Dgraph (`findRelatedCitizensByUid`) для поиска UIDs граждан, связанных с исходным гражданином (с учетом типов связей и глубины поиска из триггера).
        3.  Для *каждого* найденного UID связанного гражданина:
            *   Создает "техническое" событие `CitizenEvent` с типом, указывающим на вторичное обновление (например, `SECONDARY_UPDATE`), и ID этого связанного гражданина.
            *   Отправляет это техническое событие обратно в основной поток Flink (после источника Kafka), где оно будет обработано `StatefulApplyRulesFunction` для пересчета рейтинга связанного лица (вероятно, по специальному правилу для `SECONDARY_UPDATE`).
*   **Агрегация Событий (`EventAggregator`, `ProcessWindowFunction`):**
    *   **Назначение:** Расчет аналитических метрик по временным окнам для загрузки в OLAP-хранилище (ClickHouse).
    *   **Механизм:**
        1.  Поток `ScoreSnapshot` преобразуется и ключуется (`keyBy`) по составному ключу, включающему измерения: `region`, `city`, `district`, `eventType`, `eventSubtype`, `ageGroup`, `gender`. (Предполагается, что `ageGroup` и `gender` извлекаются из `ScoreSnapshot` или обогащаются дополнительно).
        2.  Применяется скользящее временное окно (`TumblingEventTimeWindows`) размером в 1 минуту (настраивается). Используется время события из `ScoreSnapshot`.
        3.  `com.github.kxrxh.functions.EventAggregator` (`AggregateFunction`) выполняет инкрементальную агрегацию внутри окна:
            *   Подсчитывает количество событий (`eventCount`).
            *   Суммирует изменение рейтинга (`totalScoreImpact`).
            *   Собирает уникальные ID граждан (`distinctCitizenCount`) в `HashSet` (внутри аккумулятора).
        4.  `com.github.kxrxh.functions.EventAggregateWindowFunction` (`ProcessWindowFunction`) применяется после агрегации:
            *   Получает результат от `EventAggregator`.
            *   Извлекает временные метки начала и конца окна (`window.getStart()`, `window.getEnd()`).
            *   Вычисляет среднее изменение рейтинга (`avgScoreImpact`).
            *   Формирует итоговый объект `com.github.kxrxh.model.EventAggregate`, добавляя временные метки окна.

### 2.3. Запись Результатов (Sinks)

Результаты обработки направляются в различные системы хранения:

*   **Dgraph (`com.github.kxrxh.sinks.DgraphSinkFunction`):**
    *   **Назначение:** Хранение *полной, детальной истории* изменений рейтинга как части графа `Citizen`. Идеально для запросов "покажи всю историю для гражданина X".
    *   **Механизм:** Принимает `ScoreSnapshot`. Формирует Dgraph `Upsert` операцию: находит узел `Citizen` по `citizenId`, добавляет новый объект `ScoreChange` (содержащий временную метку, изменение, ID правила) к предикату `scoreHistory`.
*   **MongoDB (`com.github.kxrxh.sinks.MongoDbSinkFunction`):**
    *   **Назначение:** Хранение *текущего* состояния рейтинга гражданина для быстрого оперативного доступа.
    *   **Механизм:** Принимает `ScoreSnapshot`. Обновляет (`updateOne`) документ в коллекции `citizens`, найденный по `citizenId`, устанавливая встроенный документ `scoreSnapshot`, содержащий актуальное значение `currentScore` и, возможно, `ratingLevel` (если рассчитывается).
*   **InfluxDB - История (`com.github.kxrxh.sinks.InfluxDbScoreHistorySink`):**
    *   **Назначение:** Хранение истории рейтинга в формате временного ряда для анализа трендов и визуализации изменений рейтинга во времени.
    *   **Механизм:** Принимает `ScoreSnapshot`. Формирует точку данных (`Point`) для measurement `score_history`. Использует `citizenId`, `eventType`, `region` и т.д. как теги (`tags`), а `currentScore`, `scoreChange` как поля (`fields`). Временная метка берется из `ScoreSnapshot`.
*   **ClickHouse - Агрегаты (`com.github.kxrxh.sinks.ClickHouseEventAggregateSink`):**
    *   **Назначение:** Запись предрасчитанных агрегатов в OLAP-хранилище для быстрой аналитики и построения отчетов.
    *   **Механизм:** Принимает `EventAggregate`. Формирует `INSERT` запрос для таблицы `events_aggregate`. Использует **пакетную запись (batching)** для повышения производительности: накапливает записи и отправляет их пачкой при достижении определенного размера или по таймауту. Параметры батчинга настраиваются.
*   **Kafka - Триггеры (`FlinkKafkaProducer`):**
    *   **Назначение:** Отправка `RelatedEffectTrigger` в топик `secondary_score_updates` для запуска асинхронной обработки связанных граждан.
    *   **Механизм:** Стандартный Flink Kafka Producer, настроенный на сериализацию `RelatedEffectTrigger` в JSON.
*   **(Планируется) InfluxDB - Сырые События (`InfluxDbEventSink`):**
    *   **Назначение:** Запись *каждого* входящего `CitizenEvent` в InfluxDB для возможности анализа исходного потока данных.
    *   **Механизм:** Будет принимать `CitizenEvent` и писать в measurement `events`, используя ID и метаданные события как теги, а `payload` как поле.
*   **(Планируется) ClickHouse - Архив Событий (`ClickHouseEventArchiveSink`):**
    *   **Назначение:** Долговременное хранение всех входящих `CitizenEvent` в ClickHouse для ретроспективного анализа и аудита.
    *   **Механизм:** Будет принимать `CitizenEvent` и писать в таблицу `events_archive`, вероятно, с использованием пакетной записи.

## 3. Конфигурация

Job настраивается через параметры командной строки, которые считываются с помощью `org.apache.flink.api.java.utils.ParameterTool`.

*   **`ParameterTool`:** Стандартный инструмент Flink для работы с параметрами. Job получает его как Global Job Parameter.
*   **`com.github.kxrxh.config.ParameterNames`:** Класс-константа, содержащий строковые ключи для всех ожидаемых параметров (например, `KAFKA_BOOTSTRAP_SERVERS`, `MONGODB_URI`). Это помогает избежать опечаток.
*   **Получение Параметров:** Каждая функция или Sink, требующая внешних подключений (`RichSinkFunction`, `RichAsyncFunction`, `SourceFunction`), получает доступ к `ParameterTool` в своем методе `open()` или конструкторе и извлекает необходимые параметры (`params.getRequired(...)`, `params.get(...)`, `params.getInt(...)`, etc.).

### Пример Запуска

```bash
# Путь к Flink CLI
FLINK_HOME=/path/to/flink

# Путь к JAR файлу вашего Job'а
JOB_JAR=./target/social-rating-flink-job-1.0-SNAPSHOT.jar

# Основной класс Job'а
MAIN_CLASS=com.github.kxrxh.DataStreamJob

$FLINK_HOME/bin/flink run -c $MAIN_CLASS $JOB_JAR \
 # Параметры Kafka
 --kafka.bootstrap.servers kafka1:9092,kafka2:9092 \
 --kafka.consumer.group.id social-rating-main-consumer \
 --kafka.topic.citizen.events citizen_events_prod \
 --kafka.topic.secondary.updates secondary_updates_prod \
 # Параметры MongoDB
 --mongodb.uri mongodb://user:password@mongo1:27017,mongo2:27017/social_rating_db?replicaSet=rs0 \
 --mongodb.database social_rating_db \
 --mongodb.collection.rules scoring_rules_prod \
 --mongodb.collection.citizens citizens_prod \
 --mongodb.polling.interval.ms 60000 \
 # Параметры Dgraph
 --dgraph.grpc.host dgraph-alpha1:9080,dgraph-alpha2:9080 \
 --dgraph.grpc.port 9080 \
 # Параметры InfluxDB
 --influxdb.url http://influxdb-host:8086 \
 --influxdb.token YOUR_SECURE_INFLUXDB_TOKEN \
 --influxdb.org your-influx-organization \
 --influxdb.bucket social-rating-bucket \
 # Параметры ClickHouse
 --clickhouse.jdbc.url jdbc:clickhouse://clickhouse-node1:8123,clickhouse-node2:8123/social_rating_data \
 # Параметры батчинга для ClickHouse Sinks (можно задать общие или специфичные)
 --clickhouse.sink.batch.size 1500 \
 --clickhouse.sink.batch.interval.ms 10000
```

**Примечание:** Рекомендуется использовать файлы конфигурации или системы управления секретами для передачи чувствительных данных (пароли, токены) вместо прямых аргументов командной строки в производственных средах.

## 4. Ключевые Компоненты (Структура Проекта)

*   **`com.github.kxrxh` (Корень):**
    *   `DataStreamJob.java`: Точка входа приложения. Собирает и запускает пайплайн Flink.
*   **`com.github.kxrxh.model`:**
    *   Содержит POJO (Plain Old Java Objects), представляющие основные структуры данных системы: `CitizenEvent`, `ScoringRule`, `ScoreSnapshot`, `RelatedEffectTrigger`, `EventAggregate`, `Location`, `ScoreChange`.
*   **`com.github.kxrxh.functions`:**
    *   Содержит пользовательские функции Flink, реализующие основную логику обработки:
        *   `StatefulApplyRulesFunction`: Ключевая функция применения правил и управления состоянием рейтинга.
        *   `AsyncRelatedEffectFunction`: Асинхронная обработка влияния на связанных граждан.
        *   `EventAggregator`: Функция инкрементальной агрегации для ClickHouse.
        *   `EventAggregateWindowFunction`: Добавляет метаданные окна к агрегатам.
*   **`com.github.kxrxh.sinks`:**
    *   Содержит реализации `SinkFunction` для записи данных в различные внешние системы: `DgraphSinkFunction`, `MongoDbSinkFunction`, `InfluxDbScoreHistorySink`, `ClickHouseEventAggregateSink`.
*   **`com.github.kxrxh.source`:**
    *   Содержит пользовательские источники данных или фабрики для них:
        *   `MongoRulePollingSource`: Источник для периодической загрузки правил из MongoDB.
        *   `KafkaSourceFactory`: Вспомогательный класс для создания источников Kafka.
*   **`com.github.kxrxh.database`:**
    *   Содержит классы-репозитории для инкапсуляции взаимодействия с базами данных: `DgraphRepository`, `MongoRepository`, `ClickHouseRepository`, `InfluxDbRepository`. Они используются Sink'ами и асинхронными функциями.
*   **`com.github.kxrxh.config`:**
    *   Содержит классы, связанные с конфигурацией:
        *   `ParameterNames`: Константы для имен параметров конфигурации.
*   **`com.github.kxrxh.serde`:**
    *   Содержит сериализаторы/десериализаторы (SerDe) для Kafka и других систем, если требуются пользовательские реализации (например, `CitizenEventDeserializationSchema`, `RelatedEffectTriggerSerializationSchema`).

## 5. Зависимости

Основные внешние библиотеки, используемые в проекте (полный список см. в `pom.xml`):

*   **Apache Flink:**
    *   `flink-streaming-java`: Ядро потоковой обработки.
    *   `flink-connector-kafka`: Для чтения/записи из/в Kafka.
    *   `flink-connector-jdbc`: Для взаимодействия с ClickHouse через JDBC.
    *   `flink-statebackend-rocksdb` (Опционально): Для использования RocksDB как бэкенда состояния для больших состояний.
*   **Базы Данных:**
    *   `mongodb-driver-sync`: Драйвер для MongoDB.
    *   `dgraph-java-client`: Клиент для Dgraph.
    *   `influxdb-client-java`: Клиент для InfluxDB v2.
    *   `clickhouse-jdbc`: JDBC драйвер для ClickHouse.
*   **Сериализация/Десериализация:**
    *   `flink-json` / `jackson-databind`: Для работы с JSON.
    *   `gson`: Альтернативная библиотека для JSON (используется в некоторых компонентах).
*   **Логирование:**
    *   `slf4j-api`: API логирования.
    *   `log4j-slf4j-impl` / `logback-classic`: Конкретная реализация логирования (настраивается через `log4j2.properties` или `logback.xml`).

## 6. Сборка и Запуск

1.  **Сборка JAR:**
    ```bash
    mvn clean package -DskipTests
    ```
    Это создаст исполняемый JAR-файл в директории `target/`.

2.  **Запуск Flink Job:**
    Используйте команду `flink run`, как показано в разделе "Пример Запуска", указав правильный путь к JAR, основной класс и все необходимые параметры конфигурации. Убедитесь, что кластер Flink запущен и доступен.

## 7. Мониторинг и Логирование

*   **Flink UI:** Веб-интерфейс Flink предоставляет детальную информацию о запущенном Job'е, включая его статус, топологию, метрики производительности (пропускная способность, задержка), состояние чекпоинтов и бэкпрессинг.
*   **Логирование:** Job использует SLF4j для логирования. Настройте уровень логирования и место вывода логов в конфигурационном файле Log4j2 (`log4j2.properties`) или Logback (`logback.xml`) в директории `src/main/resources`. Логи будут выводиться в стандартные файлы логов TaskManager'ов Flink. Анализ логов критически важен для диагностики проблем.

## 8. Дальнейшее Развитие (Планируемые Улучшения)

*   Реализация Sink'ов для записи сырых событий: `InfluxDbEventSink` и `ClickHouseEventArchiveSink`.
*   Внедрение пакетной записи (batching) в оставшиеся ClickHouse Sinks (кроме `ClickHouseEventAggregateSink`).
*   Реализация механизма "затухания" (decay) старых баллов рейтинга.
*   Поддержка дополнительных конфигураций из MongoDB (например, порогов уровней рейтинга из `rule_configurations`).
*   Разработка общего механизма оповещений (Alerting) при достижении определенных условий рейтинга.
*   Улучшение обработки ошибок и добавление стратегий retry/dead-letter-queue для Sink'ов.
*   Оптимизация запросов к Dgraph и другим базам данных. 