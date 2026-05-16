# Примеры локального запуска

./distributor -p "ERROR" -f testdata/app.log

### Поиск с регулярным выражением
./distributor -p "WARNING|ERROR" -f testdata/app.log

### Поиск с учётом регистра
./distributor -p "[A-Z]ERROR" -f testdata/app.log

# Пример запуска на 3х воркерах

### Запуск воркеров (в отдельных терминалах)
./worker -id worker1 -coordinator localhost:9090 -port 8081 -c 4
./worker -id worker2 -coordinator localhost:9090 -port 8082 -c 4
./worker -id worker3 -coordinator localhost:9090 -port 8083 -c 4

### Запуск распределённого поиска
./distributor -p "ERROR" -f testdata/large.log -workers "localhost:8081,localhost:8082,localhost:8083"

# Запуск в докере

### Запуск кластера
docker-compose up -d

### Выполнение поиска
docker-compose exec distributor ./distributor -p "ERROR" -f /data/test.log \
-workers "worker1:8081,worker2:8082,worker3:8083"

### Просмотр логов воркеров
docker-compose logs worker1

## Сравнение с grep
    
```
./benchmark_compare.sh
=== QUICK BENCHMARK ===
Creating test file with 50000 lines...

=== ORIGINAL GREP ===

real    0m0.007s
user    0m0.005s
sys     0m0.001s

=== STARTING WORKERS ===

=== DISTRIBUTED MYGREP ===

real    0m0.072s
user    0m0.062s
sys     0m0.189s

=== VERIFY RESULTS ===
Original grep lines: 5000
Mygrep lines: 5000
```
