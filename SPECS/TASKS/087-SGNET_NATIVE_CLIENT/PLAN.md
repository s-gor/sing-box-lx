# PLAN 087 — SGNET_NATIVE_CLIENT

## Архитектура

Нативный `sgnet` outbound реализуется как тонкая client-only фича за `with_sgnet`. Wire/auth/session код переносится по поведению из `s-gor/sg-net`, но адаптируется к интерфейсам sing-box вместо SOCKS/CLI. Сервер остаётся неизменным.

## Файловая карта

Новые lx-файлы/пакеты:
- `protocol/sgnet/` — SG-Net/1 frame/destination codec, auth transcript, client session/stream/packet adapter и тесты; файлы build-tagged `with_sgnet`.
- `protocol/sgnet/outbound.go` — реализация sing-box outbound поверх SG-Net client session.
- `include/sgnet.go` и `include/sgnet_stub.go` — build-tag registration/stub по принятому шаблону форка.
- `test/sgnet/` либо package-local tests — golden/compatibility fixtures без секретов.

Минимальные upstream-швы, точные имена уточняются по текущей структуре 1.14/lx до первой кодовой правки:
- outbound type constant/option union для `sgnet`;
- outbound factory registration;
- lx build tag list (`Makefile.lx`) при необходимости.

Каждый upstream-шов — отдельный атомарный коммит и `// lx:begin sgnet` marker.

## Этапы

1. Зафиксировать wire compatibility: скопировать/перепроверить константы, frame codec, destination codec и auth golden vector из `s-gor/sg-net`; сначала failing tests.
2. Реализовать TLS-exporter handshake и credential validation на клиенте; тестовый TLS server используется только в unit tests, production server код в форк не добавляется.
3. Реализовать logical TCP stream с обязательным OpenResult ACK и корректным Close/FIN/error handling.
4. Реализовать UDP packet path с destination metadata.
5. Адаптировать client session к sing-box outbound interfaces, включая reconnect после физического разрыва для новых операций.
6. Добавить option/type/factory/include wiring за `with_sgnet`; без тега конфиг SG-Net должен быть недоступен.
7. Добавить config validation и `sing-box check` fixture.
8. Прогнать package tests, race для SG-Net пакета, `go test ./...`, `go vet ./...`, no-tag build и lx build.
9. Собрать бинарь для compatibility stand и выполнить live проверки против неизменённого `orange.opik.net:443` SG-Net Server: TCP sequential/parallel, wrong credential, reconnect, UDP.
10. После функциональной совместимости отдельно снять pcap и сравнить внешний TLS/traffic fingerprint с обычным HTTPS. Не добавлять маскировку до конкретной измеренной находки.

## Зона касания upstream

Бюджет: не более трёх upstream-файлов для type/options/factory плюс build wiring. Если для базового SG-Net outbound требуется более широкая правка, реализацию остановить и переработать границу, а не размазывать diff.

## Проверки

Основные команды:

```sh
go test -tags with_sgnet ./protocol/sgnet/...
go test -race -tags with_sgnet ./protocol/sgnet/...
go test ./...
go vet ./...
go build ./...
make -f Makefile.lx lx-print-tags
make -f Makefile.lx lx-build
./sing-box check -c test/sgnet/client.json
```

На Go 1.25 полный набор lx-тегов проверяется без `badlinkname`/`with_naive_outbound` согласно IMPLEMENTATION_PROMPT; релизная полная линковка — на закреплённом Go 1.24.x.

## Коммитная дисциплина

Сначала новые SG-Net package files, затем отдельные `// lx:` upstream seams, затем build-tag wiring. Коммиты формата `lx(sgnet): ...`. Ветка `lx` напрямую не изменяется до приёмки feature branch.
