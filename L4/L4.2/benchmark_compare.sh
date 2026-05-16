#!/bin/bash

echo "=== QUICK BENCHMARK ==="

# Создаём файл с 50000 строк
echo "Creating test file with 50000 lines..."
> testdata/big.log
for i in {1..50000}; do
    [ $((i % 10)) -eq 0 ] && echo "ERROR: Error $i" >> testdata/big.log || echo "INFO: Line $i" >> testdata/big.log
done

echo ""
echo "=== ORIGINAL GREP ==="
time grep "ERROR" testdata/big.log > /dev/null

echo ""
echo "=== STARTING WORKERS ==="
go run worker.go -id w1 -distributorAddr localhost:9090 -port 8081 -c 8 > /dev/null 2>&1 &
go run worker.go -id w2 -distributorAddr localhost:9090 -port 8082 -c 8 > /dev/null 2>&1 &
go run worker.go -id w3 -distributorAddr localhost:9090 -port 8083 -c 8 > /dev/null 2>&1 &
sleep 3

echo ""
echo "=== DISTRIBUTED MYGREP ==="
time go run main.go -p "ERROR" -f testdata/big.log -workers "localhost:8081,localhost:8082,localhost:8083" > /dev/null 2>&1

pkill -f "worker.go" 2>/dev/null

echo ""
echo "=== VERIFY RESULTS ==="
echo "Original grep lines: $(grep -c "ERROR" testdata/big.log)"
echo "Mygrep lines: $(go run main.go -p "ERROR" -f testdata/big.log -workers "localhost:8081,localhost:8082,localhost:8083" 2>/dev/null | grep -c "^ERROR:")"
