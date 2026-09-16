# SPEC 087 — SGNET_NATIVE_CLIENT

**Фича:** [017-SGNET](../../FEATURES/017-SGNET/FEATURE.md)

| Поле | Значение |
|---|---|
| Тип | F |
| Статус | O |

## Цель

Добавить в sing-box-lx нативный client-only outbound `sgnet`, совместимый по проводу с существующим SG-Net/1 Server из `s-gor/sg-net`, и убрать необходимость в отдельном `sgnet-client` + локальном SOCKS hop для конечных клиентов.

## Контракт

Конечный путь данных: `TUN -> sing-box-lx router -> sgnet outbound -> TLS 1.3 -> SG-Net/1 -> existing sgnet-server`.

Первый релиз поддерживает TCP и UDP. TLS обязателен. Аутентификация использует существующую SG-Net/1 схему: per-device secret, HMAC-SHA256, client/server nonces, capability negotiation и TLS exporter label `EXPORTER-SG-NET-1`. Wire constants и encoding должны совпадать с эталонной реализацией `s-gor/sg-net`; при расхождении эталоном для совместимости является уже развернутый сервер.

## Требования

1. Фича полностью client-only; inbound/server в sing-box-lx не добавляется.
2. Весь новый функциональный код изолирован build-tag `with_sgnet`; сборка без тега не получает SG-Net.
3. Новые пакеты/файлы предпочтительнее изменения upstream. Неизбежные upstream-швы минимальны и окружены `// lx:begin sgnet` / `// lx:end sgnet`.
4. Секрет не логируется и не передаётся непосредственно в handshake.
5. TLS certificate verification и SNI обязательны; insecure skip verify отвергается конфигурацией SG-Net.
6. TCP Open подтверждается серверным OpenResult до успешного возврата Dial вызывающему коду.
7. UDP поддерживает destination metadata и ответы сервера в соответствии с SG-Net/1.
8. После потери сети новые dial/packet операции способны создать новое физическое соединение после восстановления. Уже оборванный TCP не обещает миграцию.
9. Malformed/oversized frames, auth failure, unknown required version и revoked/invalid credential дают ошибку, не downgrade.
10. Существующий SG-Net Server не меняется для прохождения compatibility test.

## Тест оправданности §3.1

(a1) Фича нужна SG Client/SG Mobile: текущий рабочий стенд использует отдельный sgnet-client и SOCKS, а целевая архитектура требует SG-Net непосредственно в поставляемом sing-box-lx.

(a2) В целевом канале sing-box-lx нативного SG-Net outbound нет.

(a3) Тонкий client-only outbound дешевле сопровождать, чем отдельный процесс, SOCKS adapter и дополнительную lifecycle/configuration интеграцию в каждом клиенте. Плановая зона upstream-касания ограничивается option/type registration, outbound factory и include/build-tag wiring; wire/client implementation размещается в новых lx-файлах/пакете.

## Критерии приёмки

- Unit tests покрывают frame codec, destination codec, auth transcript/proof, handshake, TCP OpenResult, UDP и reconnect/error paths.
- Golden vectors совместимы с `s-gor/sg-net`.
- `go test` затронутых пакетов проходит с `with_sgnet`.
- `go test ./...` и `go vet ./...` проходят в поддерживаемом toolchain согласно IMPLEMENTATION_PROMPT.
- Сборка без `with_sgnet` проходит и SG-Net отсутствует.
- Сборка lx с `with_sgnet` проходит.
- `sing-box check` принимает валидный `type: sgnet` config и отвергает небезопасный/неполный config.
- Compatibility test против неизменённого существующего SG-Net Server: TCP HTTPS возвращает серверный egress IP; 20 последовательных и 20 параллельных запросов без ошибок; неверный credential отвергается; новый запрос после реального сетевого разрыва работает без перезапуска ядра; UDP подтверждается отдельным тестом.
- После реализации заполнены TASKS.md и IMPLEMENTATION_REPORT.md; статус переводится в C только после критериев, требующих доступной среды. Live/device подтверждение переводит в D отдельно.

## Вне скоупа

SG-Net server/inbound в sing-box-lx, SG-Net H3/QUIC, SG-Net Auto, произвольный padding/timing obfuscation и обещание неразличимости DPI.
