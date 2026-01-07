# Monitoring Dashboard

Bu dizin, projenin canlı monitoring dashboard'u için gerekli konfigürasyonları içerir.

## Servisler

### cAdvisor (Port: 8082)
- Container metriklerini toplar (CPU, Memory, Network, Disk)
- URL: http://localhost:8082

### Prometheus (Port: 9090)
- Metrikleri toplar ve saklar
- URL: http://localhost:9090
- 7 günlük veri saklama

### Grafana (Port: 3001)
- Dashboard görselleştirme
- URL: http://localhost:3001
- Kullanıcı: `admin`
- Şifre: `admin`

## Kullanım

1. Servisleri başlat:
```bash
docker compose up -d cadvisor prometheus grafana
```

2. Grafana'ya giriş yap:
- http://localhost:3001
- Kullanıcı: `admin`
- Şifre: `admin`

3. Dashboard'u görüntüle:
- "Container Metrics - Real-time CPU & Memory" dashboard'u otomatik yüklenecek

## pprof Endpoints

Backend'de pprof endpoint'leri mevcut:
- http://localhost:8080/debug/pprof/
- http://localhost:8080/debug/pprof/heap (memory)
- http://localhost:8080/debug/pprof/profile (CPU)
- http://localhost:8080/debug/pprof/goroutine (goroutines)

## Örnek pprof Kullanımı

```bash
# CPU profili al (30 saniye)
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30

# Memory profili al
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine profili al
go tool pprof http://localhost:8080/debug/pprof/goroutine
```

