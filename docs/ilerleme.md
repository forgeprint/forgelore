# İlerleme

Oturumlar arası not defteri (plan, çalışma kuralı 9). Yeni bir oturuma
başlarken önce bunu oku, sonra `docs/plan.md`'yi.

---

## Faz 0 — Depo ve yönetişim

**Durum:** Tamam. Adım 1–7 bitti, kabul kriterleri karşılandı.

### 2026-09-25

Tamamlananlar:

- `[SEN]` Ad kontrolleri yapıldı, çakışma bildirilmedi.
- `[SEN]` `forgeprint/forgelore` deposu açıldı (public).
- `[SEN]` Go kuruldu: `go1.27.0 windows/amd64`.
- `[SEN]` İlk commit `main`'e push edildi.
- `[CLAUDE]` Depo iskeleti: `go.mod` (`github.com/forgeprint/forgelore`, `go 1.26`),
  bölüm 3'teki klasör ağacı, `LICENSE`, `DCO`, `CONTRIBUTING.md`, `SECURITY.md`,
  `README.md`, `.gitignore`, `.gitattributes`, `.gitleaks.toml`.
- `[CLAUDE]` `cmd/forgelore/main.go`: yalnızca `version` ve `help`. Üç birim testi.
- `[CLAUDE]` Yerel betikler: `scripts/{test,build,crosscheck,gitleaks,ci}.sh`.
- `[CLAUDE]` `.github/workflows/ci.yml`.
- `[CLAUDE]` K1–K15 için 15 ADR + `docs/adr/README.md` dizini.
- `[CLAUDE]` `docs/plan.md` (planın repodaki kopyası), bu dosya.

Kabul kriterleri:

| Kriter | Durum |
|---|---|
| Boş main altı hedefte `CGO_ENABLED=0` ile derleniyor | ✅ linux/darwin/windows × amd64/arm64 |
| ADR'ler yazılı | ✅ 15/15 |
| Yerel test komutu çalışıyor | ✅ `./scripts/ci.sh` |

CI ilk koşusunda yeşil: Ubuntu runner, 53 saniye. Exec bit, LF normalizasyonu ve
gitleaks indirmesi CI'da da çalıştı.

### Bu fazda alınan kararlar

| Konu | Karar | Gerekçe |
|---|---|---|
| Asgari Go sürümü | `go 1.26` (N-1), toolchain 1.27 | Güncelleme gecikmesine tolerans, `modernc.org/sqlite` ile sorun yok |
| Yayın aracı | Düz betik, GoReleaser yok | K3'ün sıfır-bağımlılık ruhu |
| Dil | Kamuya açık doküman + kod İngilizce, iç notlar (bu dosya, `docs/plan.md`) Türkçe | Repo public, dış katkıya açık kalsın |
| Kamuya açık doküman adları | Plandaki Türkçe adlar İngilizceye çevrildi: `surumleme.md` → `versioning.md`, `uyumluluk.md` → `compatibility.md`. `ilerleme.md` iç not olduğu için Türkçe kaldı | Dil kararının doğrudan sonucu |
| Plan dosyası | `docs/plan.md` olarak repoda | Kilitli kararların tek kaynağı git'te izlensin (plan, bölüm 6) |
| CI zamanlaması | Faz 0'da | Dal korumasında zorunlu status check için çalışan bir kontrol gerekiyor; CGO sızmasını ilk günden yakalar |
| CI mantığı | YAML'da değil betikte; workflow yalnızca `scripts/ci.sh` çağırır | "CI yalnızca aynı komutları çağırır" (plan, Faz 0 / Adım 6) |
| Action sabitleme | Commit SHA ile, etiketle değil | Etiket sonradan taşınabilir, SHA taşınamaz |
| gitleaks | Resmî action değil, sürümü ve SHA-256'sı sabit binary | Resmî action organizasyon hesaplarında lisans anahtarı istiyor; binary yerelde ve CI'da birebir aynı çalışıyor |
| DCO kontrolü | CI'da değil, GitHub DCO uygulamasında | Yasal kontrol derleme hattına girmesin |
| Satır sonu | `.gitattributes` ile her şey LF | `core.autocrlf` açık; CRLF'e dönen betik shebang satırında kırılıyor |
| Boş paketler | `internal/*` altında boş Go paketi açılmadı, `.gitkeep` kondu | Boş paket derlemede gürültü yapar; paketler Faz 1'de ihtiyaç doğdukça açılacak |

### Doğrulanan dış sabitler

Hafızadan değil, kaynağından alındı (2026-09-25):

| Şey | Değer | Kaynak |
|---|---|---|
| `actions/checkout` | `3d3c42e5aac5ba805825da76410c181273ba90b1` (v7.0.1) | GitHub API, tag → commit |
| `actions/setup-go` | `b7ad1dad31e06c5925ef5d2fc7ad053ef454303e` (v7.0.0) | GitHub API, tag → commit |
| gitleaks | v8.30.1, beş platformun SHA-256'sı `scripts/gitleaks.sh` içinde | Release'in kendi `gitleaks_8.30.1_checksums.txt` dosyası |
| Apache-2.0 metni | `LICENSE`, 202 satır, ek bölümü şablon hâliyle bırakıldı | GitHub `licenses/apache-2.0` API'si |
| DCO 1.1 metni | `DCO`, 35 satır | developercertificate.org — sayfanın güncel hâlinde adres bloğu yok, metin birebir alındı |

`gitleaks dir` ve `gitleaks version` komut sözdizimi binary çalıştırılarak
doğrulandı, dokümantasyondan tahmin edilmedi.

---

## Açık maddeler

- `[SEN]` **GitHub DCO uygulaması** depoya kurulacak. Sign-off CI'da denetlenmiyor.
- `[SEN]` **Dal koruması** ve zorunlu status check: CI artık bir kez koştuğu için
  `ci` kontrolü kural olarak seçilebilir.
- `ubuntu-latest` 2026-10-19'da Ubuntu 26'ya geçiyor (CI uyarısı). Her şeyi
  sabitleme ilkesine uyup runner imajını da sabitlemek isteyip istemediğimiz
  karara bağlı.
- `-race` ile test yok: yarış dedektörü CGO ister, Windows'ta ayrıca gcc.
  Faz 1'de eşzamanlılık girdiğinde Linux CI işi olarak yeniden değerlendirilecek.
- gitleaks sürüm yükseltme prosedürü `scripts/gitleaks.sh` başındaki yorumda;
  sürüm ve checksum birlikte değişir.
- `vendor/` henüz yok: bağımlılık girmeden `go mod vendor` bir şey üretmiyor.
  Faz 1'de `modernc.org/sqlite` eklenince oluşacak.

---

## Sıradaki

**Faz 1 — Kayıt formatı ve depolama çekirdeği.** İlk iş, kayıt şemasını önerip
onaylatmak (plan, Faz 1 / Adım 1). Sorulacaklar: kayıt tipleri yeterli mi,
frontmatter kısıtlı YAML mı, config dosyası formatı ne olacak.
