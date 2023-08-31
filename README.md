# Тестовое задание для стажёра Backend : Сервис динамического сегментирования пользователей

## Описание

Сервис, реализованный на языке Golang, предоставляющий HTTP API с форматом JSON как при отправке запроса, так и при получении результата. 

* Для роутинга запросов использовался пакет go-chi/chi;
* Для логирования использовался пакет go.uber.org/zap; 
* Для хранения использовалась СУБД PostgreSQL 15;
* Для кодирования и декодирования данных в формате json использовался встроенный пакет encoding/json;
* Для генерирования документации сервиса сделан Swagger при использовании swaggo;
* Для генерации CSV файла использовался встроенный пакет encoding/csv.

## Для запуска приложения:

```
make build && make run
```

## HTTP API 

### Метод создания сегмента

**Описание:**

В зависимости от параметров либо просто создает сегмент, либо создает сегмент и добавляет в него переданный процент случайно выбранных пользователей

**Метод:** 

`POST`

**Параметры:** 

* `slug` (обязательный) - название сегмента
* `percentage_random` (опциональный) - процент случайных пользователей для добавления в сегмент

**Ограничения на параметры:**  

* `slug` - название сегмента может состоять только из латинских a-z A-Z букв и цифр 0-9 и нижнего подчеркивания
* `percentage_random`- значния процента должно находится в пределах от 0 до 100

####  Пример запроса

```shell
curl -X POST localhost:8080/segments/SEG_1  -H 'Content-Type: application/json' -d '{"percentage_random":5}'
```

#### Пример ответа

Код ответа 200:

```json
{}  
```

Код ответа 400:

```json
{"error":{"code":400,"message":"Wrong body request or url params format"}} 
```

Код ответа 500:

```json
{"error":{"code":500,"message":"Error while processing request. Please, contact support"}} 
```
-----------------------------------------------

### Метод удаления сегмента

**Описание:**

Удаляет сегмент

**Метод:**

`DELETE`

**Параметры:**

* `slug`  (обязательный) - название сегмента. 

**Ограничения на параметры:**  

* `slug` - название сегмента может состоять только из латинских a-z A-Z букв и цифр 0-9 и нижнего подчеркивания

####  Пример запроса

```shell
 curl -X DELETE localhost:8080/segments/SEG_1  -H 'Content-Type: application/json' -d '{}'
```

#### Пример ответа

Код ответа 200:

```json
{}  
```

Код ответа 400:

```json
{"error":{"code":400,"message":"Wrong body request or url params format"}} 
```

Код ответа 500:

```json
{"error":{"code":500,"message":"Error while processing request. Please, contact support"}} 
```


------------------------

### Метод добавления пользователя в сегмент

**Описание:** 

Для пользователя удаляет сегменты из переданного списка, затем добавляет из второго переданного списка сегменты с указанным в днях для каждого сегмента TTL, если TTL не указан, то пользователь добавляется в сегмент без ограничения по времени

**Метод:** 

`PUT`

**Параметры:**

* `user_id`  (обязательный) - идентификатор пользователя

* `list_delete`(опциональный) - список сегментов для удаления 

* `list_add` (опциональный) - список сегментов для добавления
    * `segment_slug` (обязательный) - название сегмента
    * `days_ttl` (опционально) - TTL (в днях)

**Ограничения на параметры:**  

*  `segment_slug` - название сегмента может состоять только из латинских a-z A-Z букв и цифр 0-9 и нижнего подчеркивания
*  `days_ttl`  - максимально 5000

####  Пример запроса

```shell
curl -X PUT localhost:8080/users-segments/8  -H 'Content-Type: application/json' -d '{"list_add":[{"segment_slug":"SEG1","days_ttl":2},{"segment_slug":"SEG2"},{"segment_slug":"SEG3"}]}'
```

#### Пример ответа

Код ответа 200:

```json
{}  
```

Код ответа 400:

```json
{"error":{"code":400,"message":"Wrong body request or url params format"}} 
```

Код ответа 500:

```json
{"error":{"code":500,"message":"Error while processing request. Please, contact support"}} 
```

------------------------

### Метод получения активных сегментов пользователя. 

**Описание:** 

Возвращает список сегментов, в которых состоит пользователь, если таких нет, то возвращает пустой список

**Метод:** 

`GET`

**Параметры:** 

* `user_id` - идентификатор пользователя

####  Пример запроса

```shell
curl -X GET localhost:8080/users-segments/1  
```

#### Пример ответа

Код ответа 200:

```json
[]     
```

```json
[{"segment_slug":"SEG1"},{"segment_slug":"SEG2"},{"segment_slug":"SEG3"}]  
```

Код ответа 400:

```json
{"error":{"code":400,"message":"Wrong body request or url params format"}} 
```

Код ответа 500:

```json
{"error":{"code":500,"message":"Error while processing request. Please, contact support"}} 
```

------------------------

### Метод получения истории пользователей

**Описание:**

Принимает список пользователей и период (в днях) за который надо выгрузить историю добавлений и удалений пользователей в сегменты и возвращает имя файла с полученной историей. Файл генерируется в формте csv

**Метод:** 

`GET`

**Параметры:**
* `user_list` (обязательный) - список идентификаторов пользователей, для которых необходимо выгрузить историю
* `days` (обязательный) - период в днях за который надо выгрузить историю 

**Ограничения на параметры:**  
*  `days` - минимальный период 1 день, максимальный период = 5000 дней

####  Пример запроса

```shell
curl -X GET localhost:8080/history/  -H 'Content-Type: application/json' -d '{"user_list":[1,8],"days":2}'
```

#### Пример ответа

Код ответа 200:

```json
{"filename":"/tmp/file.csv"}
```

Код ответа 400:

```json
{"error":{"code":400,"message":"Wrong body request or url params format"}} 
```

Код ответа 500:

```json
{"error":{"code":500,"message":"Error while processing request. Please, contact support"}} 
```

#### Пример csv файла

идентификатор пользователя 2,сегмент3,операция (добавление/удаление),дата и время:

```csv
8,SEG1,добавление,2023-08-30T14:45:50.086161Z
8,SEG2,добавление,2023-08-30T14:45:50.086161Z
8,SEG3,удаление,2023-08-30T14:50:50.086161Z
```
## Принятые решения реализации

### Для обновления метрик

* При запросе на обновление, мы получаем список сегментов на удаление и список на добавление. Я решила, что если пользователя не было в сегменте из списка на удаление, то сегмент игнорируется и удаление не будет вноситься в историю операций. Если же у пользователя был сегмент из списка на добавление, он обновится на новое TTL и информация будет добавлена в историю.
  
 ### Для поддержания двух категорий сегментов с TTL и перманентных

* Также я решила, что в базе данных NULL в expires_time поле будет означать, что у сегмента нет TTL, соотвественно, когда я получаю актуальные сегменты пользователя я получаю сегменты с TTL и сегменты "перманентные".

### Удаление по TTL

* Есть некоторое допущение при удалении по TTL, хотя я возвращаю только актуальные сегменты, информация об "отложенном" удалении (т.е. косвенное удаление по TTL) вносится с задержкой в 1 минут в историю, хотя этот интервал можно изменить на меньший через конфиг.
