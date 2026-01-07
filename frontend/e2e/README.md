# E2E Test Suite

Bu dizin Playwright E2E testlerini içerir.

## Test Dosyaları

### Sayfa Testleri
- `dashboard.spec.ts` - Dashboard sayfası testleri
- `video-list.spec.ts` - Video listesi sayfası testleri
- `video-detail.spec.ts` - Video detay sayfası testleri
- `search.spec.ts` - Arama sayfası testleri
- `settings.spec.ts` - Ayarlar sayfası testleri
- `cost-analysis.spec.ts` - Cost Analysis sayfası testleri

### Özellik Testleri
- `transcript.spec.ts` - Transcript görüntüleme testleri
- `summary.spec.ts` - Summary görüntüleme testleri
- `similar-videos.spec.ts` - Benzer videolar testleri
- `video-actions.spec.ts` - Video aksiyonları (delete, analyze) testleri

### Genel Testler
- `navigation.spec.ts` - Navigasyon ve routing testleri
- `integration.spec.ts` - End-to-end entegrasyon testleri

## Test Çalıştırma

### Tüm Testleri Çalıştır
```bash
npm run test:e2e
```

### UI Modunda Çalıştır
```bash
npm run test:e2e:ui
```

### Headed Modda Çalıştır
```bash
npm run test:e2e:headed
```

### Debug Modu
```bash
npm run test:e2e:debug
```

### Belirli Bir Test Dosyası
```bash
npx playwright test e2e/dashboard.spec.ts
```

### Belirli Bir Test
```bash
npx playwright test -g "should display dashboard page"
```

## Test Yapısı

### Mock API Responses
Testler API çağrılarını mock'lar. `page.route()` kullanarak API yanıtlarını simüle eder.

### Test Helpers
- `fixtures.ts` - Test fixtures ve helper fonksiyonlar
- `helpers.ts` - Ortak helper fonksiyonlar
- `global-setup.ts` - Global test setup

## Test Coverage

### Dashboard
- ✅ Sayfa görüntüleme
- ✅ Video ekleme formu
- ✅ Video ekleme işlemi
- ✅ Hata durumları
- ✅ Recent videos listesi
- ✅ Empty state

### Video List
- ✅ Video listesi görüntüleme
- ✅ Filtreleme
- ✅ Sıralama
- ✅ Pagination
- ✅ Video kartları

### Video Detail
- ✅ Video detay görüntüleme
- ✅ Video player
- ✅ Video istatistikleri
- ✅ Tab navigasyonu (Transcript, Summary, Similar)
- ✅ Video analiz başlatma
- ✅ Cost breakdown card

### Search
- ✅ Arama formu
- ✅ Arama sonuçları
- ✅ Filtreleme
- ✅ Sıralama

### Settings
- ✅ Ayarlar sayfası görüntüleme
- ✅ LLM provider değiştirme
- ✅ Whisper provider değiştirme
- ✅ Model seçimi
- ✅ Ayarları kaydetme

### Cost Analysis
- ✅ Cost analysis sayfası görüntüleme
- ✅ Summary kartları
- ✅ Provider breakdown
- ✅ Operation breakdown
- ✅ Model breakdown
- ✅ Usage tablosu
- ✅ Period filtreleme
- ✅ Empty state
- ✅ Hata durumları

### Navigation
- ✅ Tüm sayfalara navigasyon
- ✅ Active link highlighting
- ✅ Sidebar görünürlüğü
- ✅ Theme toggle

### Transcript
- ✅ Transcript görüntüleme
- ✅ Transcript yükleme
- ✅ Transcript arama

### Summary
- ✅ Summary görüntüleme
- ✅ Summary oluşturma
- ✅ Key points

### Similar Videos
- ✅ Benzer videolar listesi
- ✅ Similarity skorları
- ✅ Video kartları

### Video Actions
- ✅ Video silme
- ✅ Video analiz başlatma
- ✅ Video durumu güncelleme

## CI/CD Integration

GitHub Actions veya diğer CI/CD sistemlerinde:

```yaml
- name: Install Playwright
  run: npx playwright install --with-deps

- name: Run E2E tests
  run: npm run test:e2e
```

## Best Practices

1. **Mock API Responses**: Gerçek API çağrıları yapmayın, mock kullanın
2. **Test Isolation**: Her test bağımsız olmalı
3. **Wait Strategies**: `waitFor` ve `expect` kullanarak async işlemleri bekleyin
4. **Selectors**: Stable selector'lar kullanın (data-testid tercih edilir)
5. **Error Handling**: Hata durumlarını test edin

## Troubleshooting

### Testler çok yavaş
- `fullyParallel: true` ayarını kontrol edin
- Worker sayısını artırın

### Flaky testler
- Timeout değerlerini artırın
- `waitFor` kullanarak elementlerin yüklenmesini bekleyin
- Network idle durumunu bekleyin

### API mock'ları çalışmıyor
- Route pattern'lerini kontrol edin
- Request method'larını kontrol edin
- Route order'ını kontrol edin
