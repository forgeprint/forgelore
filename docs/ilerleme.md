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

## Faz 0 — açık maddeler

- ~~`[SEN]` GitHub DCO uygulaması~~ — kuruldu (2026-09-25).
- `[SEN]` **Dal koruması** ve zorunlu status check: CI artık bir kez koştuğu için
  `ci` kontrolü kural olarak seçilebilir.
- ~~`ubuntu-latest` imaj geçişi~~ — runner `ubuntu-24.04` olarak sabitlendi.
- `-race` ile test yok: yarış dedektörü CGO ister, Windows'ta ayrıca gcc.
  Faz 1'de eşzamanlılık girdiğinde Linux CI işi olarak yeniden değerlendirilecek.
- gitleaks sürüm yükseltme prosedürü `scripts/gitleaks.sh` başındaki yorumda;
  sürüm ve checksum birlikte değişir.
- `vendor/` henüz yok: bağımlılık girmeden `go mod vendor` bir şey üretmiyor.
  Faz 1'de `modernc.org/sqlite` eklenince oluşacak.

---


## Faz 1 — Kayıt formatı ve depolama çekirdeği

**Durum:** Adım 1 tamam (şema onaylandı, `docs/record-format.md`). Adım 2–7 sırada.

### 2026-09-25 — alınan kararlar

Kullanıcı tarafından verildi, ADR-0016/0017/0018 olarak yazıldı.

**Kayıt tipleri (ADR-0016).** Beş tip kalıyor. Enjeksiyon davranışı tipe bağlı
ve sabit:

| Tip | Oturum başı dizini | Eşleşen hatada | Aramada |
|---|---|---|---|
| `fix` | hayır | evet, tam ipucu | evet |
| `dead_end` | hayır | evet, fix'in yanında | evet |
| `command` | sadece başlık | hayır | evet |
| `decision` | sadece başlık | hayır | evet |
| `note` | hayır | hayır | evet |

- `related: [id]` tek yönlü yazılır, indeks yönsüz kabul eder.
- `superseded_by: id` olan kayıt hiçbir yoldan enjekte edilmez; aramada
  "supersede edilmiş" işaretiyle görünmeye devam eder.
- Tanınmayan tip korunur ve `note` gibi davranır (ADR-0011).

**Frontmatter ve config dili (ADR-0017).** Tek bir kısıtlı YAML lehçesi.
Destek: düz anahtar/değer, dize, tamsayı, bool, dize listesi (`[a, b]` ve `- a`),
`#` yorumu, ISO-8601 UTC tarih dizesi. Destek yok: iç içe harita, anchor/alias,
çok satırlı dize, sekme. Desteklenmeyen dosya atlanır (fail-open), `doctor`
dosya ve satırıyla raporlar. Yazıcı kanonik: sabit anahtar sırası, sabit
tırnaklama, LF. Gidiş-dönüş testi zorunlu ve bayt düzeyinde.

**Config (ADR-0018).** Aynı lehçe, noktalı anahtarla gruplama
(`inject.budget_tokens`). Ekip: `.forgelore/config.yaml` (commit edilir).
Kişisel: `.forgelore/local/config.yaml` (gitignore). Öncelik sırası:
bayrak > ortam değişkeni > yerel > ekip > varsayılan. Bilinmeyen anahtar uyarı,
hata değil.

### Adım 1 — onaylanan kararlar

- `fingerprint` **tekil** kalıyor. Aynı çözüm birden fazla varyantı kapatıyorsa
  ayrı kayıtlar olur; Faz 6'daki tekrar tespiti birleştirme önerir.
- `scope` çelişkisinde **dizin yetkili**, frontmatter alanı doğrulanır, `doctor`
  raporlar. Promote = dosyayı taşımak.
- `updated` alanı **yok**. Ekip kayıtlarının geçmişi git'te; her yazmada değişen
  bir alan gidiş-dönüş testini kirletirdi.
- ULID **büyük harf** (Crockford kanonik), arama büyük/küçük harf duyarsız.
- `fix`/`dead_end` oturum başı dizinine **girmiyor** — bu benim çıkarımımdı,
  kullanıcı itiraz etmedi. Dizin sabit bir maliyet olmasın diye.

### Sırada

Adım 2–7: frontmatter ayrıştırıcı, ULID, dosya deposu, SQLite FTS5 indeksi,
maskeleme, testler. Kabul kriteri 10.000 kayıtlık sentetik depoda indeks ve
arama süresinin ölçülüp rapora yazılması.
