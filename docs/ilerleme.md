# İlerleme

Oturumlar arası not defteri (plan, çalışma kuralı 9). Yeni bir oturuma
başlarken önce bunu oku, sonra `docs/plan.md`'yi.

---

## Faz 0 — Depo ve yönetişim

**Durum:** Adım 1–4, 6, 7 tamam. Adım 5 (ADR'ler) onay bekliyor.

### 2026-09-25

Tamamlananlar:

- `[SEN]` Ad kontrolleri yapıldı, çakışma bildirilmedi.
- `[SEN]` `forgeprint/forgelore` deposu açıldı (public).
- `[SEN]` Go kuruldu: `go1.27.0 windows/amd64`.
- `[CLAUDE]` Depo iskeleti: `go.mod` (`github.com/forgeprint/forgelore`, `go 1.26`),
  bölüm 3'teki klasör ağacı, `LICENSE`, `DCO`, `CONTRIBUTING.md`, `SECURITY.md`,
  `README.md`, `.gitignore`, `.gitleaks.toml`.
- `[CLAUDE]` `cmd/forgelore/main.go`: yalnızca `version` ve `help`. Üç birim testi.
- `[CLAUDE]` Yerel betikler: `scripts/{test,build,crosscheck,gitleaks,ci}.sh`.
- `[CLAUDE]` `.github/workflows/ci.yml`.
- `[CLAUDE]` `docs/plan.md` (planın repodaki kopyası), bu dosya.

Kabul kriteri doğrulandı: `./scripts/ci.sh` yerelde geçiyor — gofmt, vet, test,
altı hedefin tamamı `CGO_ENABLED=0` ile derleniyor, gitleaks temiz.

### Bu fazda alınan kararlar

| Konu | Karar | Gerekçe |
|---|---|---|
| Asgari Go sürümü | `go 1.26` (N-1), toolchain 1.27 | Güncelleme gecikmesine tolerans, `modernc.org/sqlite` ile sorun yok |
| Yayın aracı | Düz betik, GoReleaser yok | K3'ün sıfır-bağımlılık ruhu |
| Dil | Kamuya açık doküman + kod İngilizce, iç notlar (bu dosya, `docs/plan.md`) Türkçe | Repo public, dış katkıya açık kalsın |
| Plan dosyası | `docs/plan.md` olarak repoda | Kilitli kararların tek kaynağı git'te izlensin (plan, bölüm 6) |
| CI zamanlaması | Faz 0'da | Dal korumasında zorunlu status check için çalışan bir kontrol gerekiyor; CGO sızmasını ilk günden yakalar |
| CI mantığı | YAML'da değil betikte; workflow yalnızca `scripts/ci.sh` çağırır | "CI yalnızca aynı komutları çağırır" (plan, Faz 0 / Adım 6) |
| Action sabitleme | Commit SHA ile, etiketle değil | Etiket sonradan taşınabilir, SHA taşınamaz |
| gitleaks | Resmî action değil, sürümü ve SHA-256'sı sabit binary | Resmî action organizasyon hesaplarında lisans anahtarı istiyor; binary yerelde ve CI'da birebir aynı çalışıyor |
| DCO kontrolü | CI'da değil, GitHub DCO uygulamasında | `[SEN]` — uygulama henüz kurulmadı |
| Boş paketler | `internal/*` altında boş Go paketi açılmadı, `.gitkeep` kondu | Boş paket derlemede gürültü yapar; paketler Faz 1'de ihtiyaç doğdukça açılacak |

### Doğrulanan dış sabitler

Hafızadan değil, kaynağından alındı (2026-09-25):

| Şey | Değer | Kaynak |
|---|---|---|
| `actions/checkout` | `3d3c42e5aac5ba805825da76410c181273ba90b1` (v7.0.1) | GitHub API, tag → commit |
| `actions/setup-go` | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0) | GitHub API, tag → commit |
| gitleaks | v8.30.1, 5 platform SHA-256'sı `scripts/gitleaks.sh` içinde | Release'in kendi `gitleaks_8.30.1_checksums.txt` dosyası |
| Apache-2.0 metni | `LICENSE`, 202 satır, ek bölümü şablon hâliyle bırakıldı | GitHub `licenses/apache-2.0` API'si |
| DCO 1.1 metni | `DCO`, 35 satır | developercertificate.org — sayfanın güncel hâlinde adres bloğu yok, metin birebir alındı |

`gitleaks dir` ve `gitleaks version` komut sözdizimi binary çalıştırılarak
doğrulandı, dokümantasyondan tahmin edilmedi.

---

## Açık maddeler

- `[SEN]` **Dal koruması** kuruldu mu? İlk push `main`'e doğrudan mı gidecek,
  yoksa dal + PR akışı mı?
- `[SEN]` **GitHub DCO uygulaması** depoya kurulacak.
- `[SEN]` Git kimliği: global `aliosman.m@hotmail.com`. GitHub hesabına kayıtlı
  değilse sign-off profille eşleşmez.
- `[CLAUDE]` **Faz 0 / Adım 5:** K1–K15 için ADR'ler. Ayrı onay bekliyor.
- `-race` ile test yok: yarış dedektörü CGO ister, Windows'ta ayrıca gcc.
  Faz 1'de eşzamanlılık girdiğinde Linux CI işi olarak yeniden değerlendirilecek.
- gitleaks sürüm yükseltme prosedürü `scripts/gitleaks.sh` başındaki yorumda;
  sürüm ve checksum birlikte değişir.
