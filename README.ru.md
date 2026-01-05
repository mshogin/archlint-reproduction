# Эксперимент: Воспроизведение проекта по спецификациям с помощью Claude Code

**Дата эксперимента:** 2025-12-30

[English version](README.md)

---

## Описание эксперимента

### Цель

Проверить, насколько точно Claude Code может воссоздать существующий проект, имея только спецификации (без доступа к исходному коду).

### Методология

1. **Исходный проект:** `archlint` - инструмент для анализа архитектуры Go-проектов
2. **Спецификации:** 10 детальных spec-файлов в формате Markdown + PlantUML (73 KB)
3. **Процесс:** Claude Code получил пустую директорию и спецификации, затем реализовал проект с нуля
4. **Результат:** `archlint-reproduction` - воссозданный проект

### Входные данные

| Спецификация | Описание | Размер |
|--------------|----------|--------|
| [0001-init-project.ru.md](specs/todo/0001-init-project.ru.md) | Инициализация Go модуля | 3.6 KB |
| [0002-makefile.ru.md](specs/todo/0002-makefile.ru.md) | Система сборки | 3.3 KB |
| [0003-data-model.ru.md](specs/todo/0003-data-model.ru.md) | Модель данных Graph/Node/Edge | 5.3 KB |
| [0004-go-analyzer.ru.md](specs/todo/0004-go-analyzer.ru.md) | Анализатор Go кода | 10.5 KB |
| [0005-cli-framework.ru.md](specs/todo/0005-cli-framework.ru.md) | CLI фреймворк на Cobra | 4.9 KB |
| [0006-collect-command.ru.md](specs/todo/0006-collect-command.ru.md) | Команда сбора архитектуры | 7.3 KB |
| [0007-tracer-library.ru.md](specs/todo/0007-tracer-library.ru.md) | Библиотека трассировки | 11.2 KB |
| [0008-trace-command.ru.md](specs/todo/0008-trace-command.ru.md) | Команда генерации контекстов | 11.4 KB |
| [0009-tracerlint.ru.md](specs/todo/0009-tracerlint.ru.md) | Линтер для tracer вызовов | 7.5 KB |
| [0010-integration-tests.ru.md](specs/todo/0010-integration-tests.ru.md) | Интеграционные тесты | 8.3 KB |

### Время выполнения

- Реализация всех 10 спецификаций: ~20 минут
- Исправление ошибок компиляции: ~5 минут
- Прохождение тестов: ~5 минут
- **Итого:** ~30 минут

---

## Executive Summary

| Метрика                       | Значение  |
|-------------------------------|-----------|
| Успешность клонирования       | **85.5%** |
| Структурная идентичность      | **100%**  |
| Семантическая эквивалентность | **~75%**  |
| Количество мутаций            | **23**    |
| Критических мутаций           | **3**     |
| Средних мутаций               | **8**     |
| Минорных мутаций              | **12**    |

**Вердикт:** Клонирование УСПЕШНО - проект воспроизведен с сохранением основной функциональности, но с заметными мутациями в деталях реализации.

---

## 1. Статистика кода

### 1.1 Размер кодовой базы

| Метрика                  | Оригинал | Reproduction | Разница       |
|--------------------------|----------|--------------|---------------|
| Всего строк Go кода      | 2,159    | 1,845        | -314 (-14.5%) |
| Количество .go файлов    | 13       | 13           | 0             |
| Спецификаций реализовано | 10       | 10           | 0             |

### 1.2 Размер по файлам

| Файл                            | Оригинал | Reproduction | Разница       |
|---------------------------------|----------|--------------|---------------|
| cmd/archlint/main.go            | 20       | 19           | -1 (-5%)      |
| cmd/tracelint/main.go           | 14       | 11           | -3 (-21%)     |
| internal/analyzer/go.go         | 862      | 694          | -168 (-19.5%) |
| internal/cli/collect.go         | 159      | 144          | -15 (-9.4%)   |
| internal/cli/root.go            | 38       | 33           | -5 (-13.2%)   |
| internal/cli/trace.go           | 123      | 117          | -6 (-4.9%)    |
| internal/linter/tracerlint.go   | 362      | 292          | -70 (-19.3%)  |
| internal/model/model.go         | 23       | 25           | +2 (+8.7%)    |
| pkg/tracer/trace.go             | 166      | 180          | +14 (+8.4%)   |
| pkg/tracer/context_generator.go | 392      | 330          | -62 (-15.8%)  |

---

## 2. Структурное соответствие

### 2.1 Директории - ИДЕНТИЧНЫ (100%)

```
cmd/
  archlint/      [OK]
  tracelint/     [OK]
internal/
  analyzer/      [OK]
  cli/           [OK]
  linter/        [OK]
  model/         [OK]
pkg/
  tracer/        [OK]
specs/
  done/          [OK]
  inprogress/    [OK]
  todo/          [OK]
tests/
  testdata/      [OK]
```

### 2.2 Файлы по спецификациям

| Спецификация           | Файлы                         | Статус           |
|------------------------|-------------------------------|------------------|
| 0001-init-project      | go.mod, go.sum                | OK               |
| 0002-makefile          | Makefile                      | OK (упрощен)     |
| 0003-data-model        | internal/model/model.go       | OK               |
| 0004-go-analyzer       | internal/analyzer/go.go       | OK (рефакторинг) |
| 0005-cli-framework     | internal/cli/root.go          | OK               |
| 0006-collect-command   | internal/cli/collect.go       | OK               |
| 0007-tracer-library    | pkg/tracer/*.go               | OK (расширен)    |
| 0008-trace-command     | internal/cli/trace.go         | OK               |
| 0009-tracerlint        | internal/linter/tracerlint.go | OK (упрощен)     |
| 0010-integration-tests | tests/*.go                    | OK               |

---

## 3. Каталог мутаций

### 3.1 Критические мутации (влияют на поведение)

#### M-CRIT-01: Изменена логика buildSequenceDiagram()

**Файл:** `pkg/tracer/context_generator.go`

**Оригинал:**
```go
if len(*callStack) > 0 {
    from := (*callStack)[len(*callStack)-1]
    diagram.Calls = append(diagram.Calls, SequenceCall{
        From:    from,
        To:      call.Function,
        Success: true,
    })
}
```

**Reproduction:**
```go
if len(callStack) > 1 {
    from := callStack[len(callStack)-2]
    to := callStack[len(callStack)-1]
    diagram.Calls = append(diagram.Calls, SequenceCall{
        From:    from,
        To:      to,
        Success: call.Event == "exit_success",
        Error:   call.Error,
    })
}
```

**Влияние:** Изменен алгоритм построения диаграммы последовательности. Reproduction записывает вызовы только при глубине стека > 1 и использует индекс-2 как источник.

---

#### M-CRIT-02: isTracerExitCall() принимает Exit()

**Файл:** `internal/linter/tracerlint.go`

**Оригинал:**
```go
return isTracerCall(exprStmt.X, "ExitSuccess") || isTracerCall(exprStmt.X, "ExitError")
```

**Reproduction:**
```go
return isTracerCall(stmt, "ExitSuccess") || isTracerCall(stmt, "ExitError") || isTracerCall(stmt, "Exit")
```

**Влияние:** Линтер в reproduction принимает deprecated `Exit()` как валидный вызов, оригинал - нет.

---

#### M-CRIT-03: GoAnalyzer с дополнительными полями

**Файл:** `internal/analyzer/go.go`

**Оригинал:**
```go
type GoAnalyzer struct {
    packages  map[string]*PackageInfo
    types     map[string]*TypeInfo
    functions map[string]*FunctionInfo
    methods   map[string]*MethodInfo
    nodes     []model.Node
    edges     []model.Edge
}
```

**Reproduction:**
```go
type GoAnalyzer struct {
    packages   map[string]*PackageInfo
    types      map[string]*TypeInfo
    functions  map[string]*FunctionInfo
    methods    map[string]*MethodInfo
    nodes      []model.Node
    edges      []model.Edge
    baseDir    string     // NEW
    modulePath string     // NEW
}
```

**Влияние:** Reproduction определяет путь модуля из go.mod, оригинал использует directory-based подход.

---

### 3.2 Средние мутации (влияют на вывод/логи)

| ID       | Файл                            | Описание                                             | Тип         |
|----------|---------------------------------|------------------------------------------------------|-------------|
| M-MED-01 | pkg/tracer/trace.go             | Добавлена package-level функция `Save()`             | Расширение  |
| M-MED-02 | internal/cli/*.go               | Удалены tracer вызовы из `init()` функций            | Упрощение   |
| M-MED-03 | cmd/tracelint/main.go           | Удалены tracer вызовы полностью                      | Упрощение   |
| M-MED-04 | internal/cli/root.go            | Удален вывод ошибок в stderr                         | Изменение   |
| M-MED-05 | internal/cli/trace.go           | Упрощен вывод `printContextsInfo()`                  | Изменение   |
| M-MED-06 | pkg/tracer/context_generator.go | Добавлена сортировка компонентов                     | Расширение  |
| M-MED-07 | internal/analyzer/go.go         | Методы переименованы (parseTypeDecl -> parseGenDecl) | Рефакторинг |
| M-MED-08 | internal/linter/tracerlint.go   | Упрощена логика `isExcluded()`                       | Упрощение   |

### 3.3 Минорные мутации (косметические)

| ID       | Описание                                                     |
|----------|--------------------------------------------------------------|
| M-MIN-01 | Язык комментариев: русский -> английский                     |
| M-MIN-02 | Tracer paths с префиксом пакета (`cli.Execute` vs `Execute`) |
| M-MIN-03 | Имена переменных (`excludedPackages` -> `excludePackages`)   |
| M-MIN-04 | Порядок функций в файлах изменен                             |
| M-MIN-05 | Порядок объявления переменных изменен                        |
| M-MIN-06 | Error messages на английском                                 |
| M-MIN-07 | Убраны emoji из вывода                                       |
| M-MIN-08 | Makefile упрощен                                             |
| M-MIN-09 | README упрощен                                               |
| M-MIN-10 | .archlint.yaml пустой (вместо 1091 строки правил)            |
| M-MIN-11 | Версии зависимостей обновлены                                |
| M-MIN-12 | Go версия: 1.25.1 -> 1.25.4                                  |

---

## 4. Анализ по спецификациям

### 4.1 Соответствие спецификациям

| Спецификация           | Соответствие | Мутации                  |
|------------------------|--------------|--------------------------|
| 0001-init-project      | 100%         | Обновлены версии         |
| 0002-makefile          | 80%          | Упрощен, убраны emoji    |
| 0003-data-model        | 100%         | Только язык комментариев |
| 0004-go-analyzer       | 90%          | M-CRIT-03, M-MED-07      |
| 0005-cli-framework     | 95%          | M-MED-04                 |
| 0006-collect-command   | 90%          | M-MED-02                 |
| 0007-tracer-library    | 95%          | M-MED-01 (расширение)    |
| 0008-trace-command     | 85%          | M-MED-02, M-MED-05       |
| 0009-tracerlint        | 85%          | M-CRIT-02, M-MED-08      |
| 0010-integration-tests | 95%          | Минорные изменения       |

### 4.2 Acceptance Criteria

Проверка ключевых acceptance criteria из спецификаций:

| Спецификация | Критерий                                     | Статус                   |
|--------------|----------------------------------------------|--------------------------|
| 0003         | Graph, Node, Edge типы определены            | PASS                     |
| 0003         | YAML сериализация работает                   | PASS                     |
| 0004         | Анализ пакетов Go работает                   | PASS                     |
| 0004         | Граф зависимостей строится                   | PASS                     |
| 0006         | Команда collect генерирует architecture.yaml | PASS                     |
| 0007         | Tracer записывает Enter/Exit                 | PASS                     |
| 0007         | Генерация PlantUML работает                  | PASS (измененный формат) |
| 0008         | Команда trace генерирует contexts.yaml       | PASS                     |
| 0009         | Линтер проверяет tracer вызовы               | PASS (с изменениями)     |
| 0010         | Интеграционные тесты проходят                | PASS                     |

---

## 5. Причины мутаций

### 5.1 Интерпретация спецификаций

Claude Code интерпретировал спецификации следуя их духу, но не букве:
- Упростил код где это казалось разумным
- Добавил функциональность где посчитал полезным (Save())
- Изменил алгоритмы сохраняя общую цель

### 5.2 Языковые различия

- Оригинал написан с русскими комментариями и сообщениями
- Claude Code естественно использовал английский

### 5.3 Стилистические предпочтения

- Разный стиль организации кода (порядок функций, переменных)
- Разный подход к именованию (tracer paths)
- Разный уровень детализации tracer инструментирования

---

## 6. Функциональное тестирование

### 6.1 Тесты компиляции

```
Оригинал:      go build ./... - PASS
Reproduction:  go build ./... - PASS
```

### 6.2 Совместимость выхода

| Артефакт          | Совместим | Примечание               |
|-------------------|-----------|--------------------------|
| architecture.yaml | Да        | Формат идентичен         |
| contexts.yaml     | Частично  | Разный формат путей      |
| *.puml файлы      | Частично  | Разный формат участников |
| Tracer JSON       | Да        | Формат идентичен         |

---

## 7. Выводы

### 7.1 Успешность эксперимента

**Spec-driven development с Claude Code работает** - по 10 спецификациям был успешно создан функционирующий проект, который:
- Имеет идентичную структуру директорий
- Реализует все требуемые компоненты
- Проходит компиляцию и тесты
- Выполняет основные функции

### 7.2 Ограничения подхода

1. **Детали реализации варьируются** - одна и та же спецификация может быть реализована по-разному
2. **Алгоритмы интерпретируются** - сложная логика (buildSequenceDiagram) была реализована иначе
3. **Стиль кода отличается** - организация, именование, комментарии
4. **Контекст теряется** - неявные решения оригинала не переносятся

### 7.3 Рекомендации для улучшения спецификаций

Для более точного клонирования спецификации должны содержать:

1. **Псевдокод критических алгоритмов** - не только "что делает", но "как делает"
2. **Примеры входов/выходов** - конкретные тест-кейсы
3. **Требования к tracer инструментированию** - какие функции должны иметь tracer
4. **Требования к стилю кода** - именование, организация
5. **Конкретные acceptance tests** - код тестов, не только описания

---

## 8. Метрики качества клонирования

| Аспект            | Оценка | Комментарий             |
|-------------------|--------|-------------------------|
| Структура         | 10/10  | Полное соответствие     |
| Типы данных       | 9/10   | Минорные расширения     |
| API/Интерфейсы    | 8/10   | Сохранены с изменениями |
| Алгоритмы         | 7/10   | Критические мутации     |
| Tracer интеграция | 7/10   | Упрощена                |
| Вывод программы   | 8/10   | Совместим частично      |
| Тесты             | 9/10   | Проходят                |

**Общая оценка: 8.3/10**

---

## Воспроизведение эксперимента

### Необходимые условия

- Claude Code (использовался Claude Opus 4.5)

### Команды

```bash
# Клонировать репозиторий
git clone <repo-url> archlint-reproduction
cd archlint-reproduction

# Удалить существующую реализацию (оставить только спеки)
rm -rf cmd internal pkg tests Makefile go.mod go.sum

# Запустить Claude Code
claude

# Дать инструкцию:
# "Реализуй проект по спецификациям из specs/todo/ в порядке номеров"
```

### Ожидаемый результат

После выполнения должен получиться функционирующий Go-проект с:
- CLI командами `archlint collect` и `archlint trace`
- Линтером `tracelint`
- Библиотекой трассировки `pkg/tracer`
- Проходящими интеграционными тестами

---

## Заключение

Эксперимент показал, что **spec-driven development с Claude Code работает**:
- Проект успешно воспроизведен с оценкой 8.3/10
- Структура и основная функциональность сохранены на 100%
- Мутации возникают в деталях реализации, особенно в алгоритмах

**Главный вывод:** Для точного воспроизведения спецификации должны содержать не только "что делать", но и "как делать" - псевдокод алгоритмов, примеры входов/выходов, executable acceptance tests

---

## Приложение A: Список файлов

### Оригинал
```
archlint/
  cmd/archlint/main.go
  cmd/tracelint/main.go
  internal/analyzer/go.go
  internal/cli/collect.go
  internal/cli/root.go
  internal/cli/trace.go
  internal/linter/tracerlint.go
  internal/model/model.go
  pkg/tracer/trace.go
  pkg/tracer/context_generator.go
  tests/fullcycle_test.go
  tests/testdata/sample/calculator.go
  tests/testdata/sample/calculator_traced_test.go
```

### Reproduction
```
archlint-reproduction/
  cmd/archlint/main.go
  cmd/tracelint/main.go
  internal/analyzer/go.go
  internal/cli/collect.go
  internal/cli/root.go
  internal/cli/trace.go
  internal/linter/tracerlint.go
  internal/model/model.go
  pkg/tracer/trace.go
  pkg/tracer/context_generator.go
  tests/fullcycle_test.go
  tests/testdata/sample/calculator.go
  tests/testdata/sample/calculator_traced_test.go
```

---

## Приложение B: Спецификации

Спецификации в обоих проектах идентичны (скопированы):

| Файл | Размер |
|------|--------|
| [0001-init-project.ru.md](specs/todo/0001-init-project.ru.md) | 3,675 bytes |
| [0002-makefile.ru.md](specs/todo/0002-makefile.ru.md) | 3,312 bytes |
| [0003-data-model.ru.md](specs/todo/0003-data-model.ru.md) | 5,270 bytes |
| [0004-go-analyzer.ru.md](specs/todo/0004-go-analyzer.ru.md) | 10,539 bytes |
| [0005-cli-framework.ru.md](specs/todo/0005-cli-framework.ru.md) | 4,878 bytes |
| [0006-collect-command.ru.md](specs/todo/0006-collect-command.ru.md) | 7,296 bytes |
| [0007-tracer-library.ru.md](specs/todo/0007-tracer-library.ru.md) | 11,179 bytes |
| [0008-trace-command.ru.md](specs/todo/0008-trace-command.ru.md) | 11,399 bytes |
| [0009-tracerlint.ru.md](specs/todo/0009-tracerlint.ru.md) | 7,505 bytes |
| [0010-integration-tests.ru.md](specs/todo/0010-integration-tests.ru.md) | 8,276 bytes |

**Всего:** 73,329 bytes спецификаций
