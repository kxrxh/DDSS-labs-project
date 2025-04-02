-- Таблица для агрегированных данных по событиям
CREATE TABLE events_aggregate
(
    date Date,                            -- Дата агрегации
    datetime DateTime,                    -- Время агрегации
    region String,                        -- Регион
    city String,                          -- Город
    district String,                      -- Район
    event_type String,                    -- Тип события
    event_subtype String,                 -- Подтип события
    age_group String,                     -- Возрастная группа граждан
    gender String,                        -- Пол
    occupation_category String,           -- Категория профессии
    income_bracket String,                -- Категория дохода
    event_count UInt32,                   -- Количество событий
    total_score_impact Int32,             -- Суммарное влияние на рейтинг
    avg_score_impact Float32,             -- Среднее влияние на рейтинг
    citizen_count UInt32                  -- Количество затронутых граждан
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, region, event_type, event_subtype)
SETTINGS index_granularity=8192;

-- Таблица для агрегированных данных по социальному рейтингу
CREATE TABLE credit_score_aggregate
(
    date Date,                            -- Дата агрегации
    datetime DateTime,                    -- Время агрегации
    region String,                        -- Регион
    city String,                          -- Город
    district String,                      -- Район
    age_group String,                     -- Возрастная группа
    gender String,                        -- Пол
    occupation_category String,           -- Категория профессии
    income_bracket String,                -- Категория дохода
    avg_score Float32,                    -- Средний рейтинг
    median_score Float32,                 -- Медианный рейтинг
    min_score Int32,                      -- Минимальный рейтинг
    max_score Int32,                      -- Максимальный рейтинг
    excellent_tier_count UInt32,          -- Количество граждан в уровне "Образцовый"
    standard_tier_count UInt32,           -- Количество граждан в уровне "Обычный"
    concerning_tier_count UInt32,         -- Количество граждан в уровне "Тревожный"
    untrusted_tier_count UInt32,          -- Количество граждан в уровне "Неблагонадежный"
    score_improvement_count UInt32,       -- Количество улучшений рейтинга
    score_deterioration_count UInt32,     -- Количество ухудшений рейтинга
    total_citizens UInt32                 -- Общее количество граждан
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, region, age_group, gender)
SETTINGS index_granularity=8192;

-- Таблица для агрегированных данных по социальным связям
CREATE TABLE social_graph_aggregate
(
    date Date,                            -- Дата агрегации
    datetime DateTime,                    -- Время агрегации
    region String,                        -- Регион
    relation_type String,                 -- Тип связи (семья, коллеги, друзья)
    score_bucket Int16,                   -- Категория рейтинга (-100 до 100, шаг 10)
    relation_count UInt32,                -- Количество связей
    avg_connection_depth Float32,         -- Средняя глубина связей
    max_connection_depth UInt8,           -- Максимальная глубина связей
    network_density Float32,              -- Плотность сети
    avg_score_difference Float32          -- Среднее различие в рейтинге между связанными гражданами
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, region, relation_type, score_bucket)
SETTINGS index_granularity=8192;

-- Таблица для хранения долговременных данных событий (архив)
CREATE TABLE events_archive
(
    id UUID,                             -- ID события
    date Date,                           -- Дата события
    datetime DateTime,                   -- Время события
    citizen_id String,                   -- ID гражданина
    event_type String,                   -- Тип события
    event_subtype String,                -- Подтип события
    score_impact Int16,                  -- Влияние на рейтинг
    region String,                       -- Регион
    city String,                         -- Город
    district String,                     -- Район
    source_system String,                -- Система-источник
    has_related_documents UInt8,         -- Флаг наличия связанных документов
    metadata String                      -- Метаданные в формате JSON
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, citizen_id, event_type)
TTL date + toIntervalYear(5)
SETTINGS index_granularity=8192;

-- Таблица для анализа темпов изменения рейтинга
CREATE TABLE score_velocity_metrics
(
    date Date,                           -- Дата расчета
    datetime DateTime,                   -- Время расчета
    region String,                       -- Регион
    city String,                         -- Город
    district String,                     -- Район
    score_bracket Int16,                 -- Категория текущего рейтинга
    daily_velocity_avg Float32,          -- Средняя скорость изменения за день
    weekly_velocity_avg Float32,         -- Средняя скорость изменения за неделю
    monthly_velocity_avg Float32,        -- Средняя скорость изменения за месяц
    velocity_std_dev Float32,            -- Стандартное отклонение скорости
    improvement_ratio Float32,           -- Отношение улучшений к ухудшениям
    citizen_count UInt32                 -- Количество граждан
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (date, region, score_bracket)
SETTINGS index_granularity=8192;

-- Представление для анализа аномалий в данных
CREATE VIEW event_anomalies AS
SELECT
    date,
    event_type,
    region,
    event_count,
    avg_score_impact,
    runningDifference(event_count) AS count_change,
    runningDifference(avg_score_impact) AS impact_change
FROM
(
    SELECT
        date,
        event_type,
        region,
        sum(event_count) AS event_count,
        avg(avg_score_impact) AS avg_score_impact
    FROM events_aggregate
    GROUP BY
        date,
        event_type,
        region
    ORDER BY
        date ASC
);

-- Создание словаря для категорий событий
CREATE DICTIONARY event_categories
(
    event_type String,
    event_subtype String,
    category_name String,
    description String,
    severity_level UInt8,
    default_impact Int16
)
PRIMARY KEY event_type, event_subtype
SOURCE(HTTP(URL 'http://example.com/dictionaries/event_categories.json' FORMAT 'JSONEachRow'))
LIFETIME(MIN 300 MAX 3600)
LAYOUT(HASHED());

-- Функция для расчета прогнозируемого рейтинга на основе текущего тренда
CREATE FUNCTION predictedScoreBasedOnVelocity AS (current_score, velocity, days_ahead) -> {
    return current_score + (velocity * days_ahead);
}; 