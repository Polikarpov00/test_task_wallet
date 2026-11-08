#!/usr/bin/env bash
# Нагрузочный тест: RATE запросов/сек на ОДИН кошелёк (депозит, снятие, баланс по кругу).
set -euo pipefail

URL=${URL:-http://localhost:8081}
RATE=${RATE:-1000}
DURATION=${DURATION:-30s}

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT   # временные файлы удалятся при любом выходе

ID=$(uuidgen | tr A-Z a-z)

# Большой стартовый баланс, чтобы снятия не упирались в 422
curl -sf -XPOST "$URL/api/v1/wallet" \
  -d "{\"walletId\":\"$ID\",\"operationType\":\"DEPOSIT\",\"amount\":1000000}" >/dev/null

echo "{\"walletId\":\"$ID\",\"operationType\":\"DEPOSIT\",\"amount\":10}"  > "$TMP/deposit.json"
echo "{\"walletId\":\"$ID\",\"operationType\":\"WITHDRAW\",\"amount\":10}" > "$TMP/withdraw.json"

# vegeta перебирает цели по кругу: депозит → снятие → GET → депозит ...
cat > "$TMP/targets.txt" <<EOF
POST $URL/api/v1/wallet
Content-Type: application/json
@$TMP/deposit.json

POST $URL/api/v1/wallet
Content-Type: application/json
@$TMP/withdraw.json

GET $URL/api/v1/wallets/$ID
EOF

echo "wallet: $ID, rate: $RATE/s, duration: $DURATION"
vegeta attack -targets="$TMP/targets.txt" -rate="$RATE" -duration="$DURATION" \
  | tee "$TMP/results.bin" | vegeta report
vegeta report -type='hist[0,5ms,10ms,50ms,100ms,500ms,1s]' < "$TMP/results.bin"

echo "final balance (ожидаем ~1000000):"
curl -s "$URL/api/v1/wallets/$ID"; echo