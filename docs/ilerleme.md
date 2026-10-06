# İlerleme

Oturumlar arası not defteri (plan, çalışma kuralı 9). Yeni bir oturuma
başlarken önce bunu oku, sonra `docs/plan.md`'yi.

Dosya kronolojik ve uzun. **Nerede kaldığımız aşağıdaki özette**; gerekçeler
ve nasıl bulunduğu tarihli bölümlerde, en yenisi en sonda.

## Nerede kaldık — 2026-10-06

| | |
|---|---|
| Son sürüm | **v0.1.9**, GitHub release attested, npm'de yedi paket, `latest` 0.1.9 |
| Kayıt şeması | 1 |
| Eşleme formatı | 2 (`claude-code` 2 kullanıyor; diğerleri 1) |
| Ajanlar | Claude Code 2.1.290 **A** · Copilot CLI 1.0.91 **B** · Gemini CLI 0.62.0 **B** · Codex CLI doğrulanmamış · Cursor başlanmadı |
| Hata korpusu | 31 aile |
| Ajan payload'ları | 16, dört ajan sürümünden |
| Bu deponun kendi hafızası | `.forgelore/records/`'ta 9 kayıt (4 fix, 5 decision) |

Yayım zinciri elle müdahale istemiyor: etiket → `release.yml` derler, attest
eder, **taslak** çıkarır; taslağı bir insan yayımlayınca `npm.yml` paketleri
release'ten indirip trusted publishing ile yayımlar.

**Açık işler — hepsi `[SEN]`:** iki kişilik bir haftalık ekip denemesi
(`docs/team-trial.md`), Codex CLI doğrulaması, Cursor.

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

### 2026-09-25 — Adım 2, 3, 4 tamam

- `internal/record/yaml.go` — kısıtlı YAML okuyucu. Desteklenmeyen her yapı
  satır numarasıyla ve **adıyla** hata veriyor ("anchors and aliases are not
  supported"), sadece "parse error" değil.
- `internal/record/encode.go` — kanonik yazıcı.
- `internal/record/record.go` — `Record`, `Decode`, `Encode`, doğrulama,
  `InSessionIndex()` / `InjectableOnMatch()`.
- `internal/record/ulid.go` — ULID üretimi, aynı milisaniyede ve saat geri
  giderse bile monotonik.
- `internal/store/store.go` — iki kapsam, atomik yazma (geçici dosya + rename +
  fsync), `Get`/`List`/`All`/`Promote`/`Init`, `Problem` raporlaması.

Testler: 4 dosya, `go test ./...` geçiyor. Kapsam: `record` %90.3,
`store` %75.4, `cmd/forgelore` %66.7.

Kanıtlanan davranışlar: gidiş-dönüş bayt eşitliği (tanınmayan alanlar dahil),
elle yazılmış dosyanın normalize edilip sonra sabit kalması, desteklenmeyen 16
YAML yapısının reddi, tanınmayan tipin `note` gibi davranması, dizin-kapsam
çelişkisinin düzeltilip raporlanması, bozuk dosyanın atlanması, promote'un
dosyayı taşıması.

Uygulama sırasında netleşen iki kural `docs/record-format.md`'ye yazıldı:
`#` yorumu tırnaksız değerde ancak boşluktan sonra başlar, ve tek tırnak bir
dize biçemi değil (hata veriyor).

### 2026-09-27 — Adım 5, 6, 7 tamam

- `internal/redact/` — maskeleme. 12 desen + `<private>` etiketi. Yazmadan önce
  çalışıyor: `store.Put` artık `([]redact.Finding, error)` döndürüyor ve
  kayıttaki metni de yerinde temizliyor, böylece çağıran elinde maskelenmemiş
  sürüm kalmıyor.
- `internal/store/index.go` — SQLite FTS5 indeksi. Artımlı güncelleme
  (mtime + boyut eşleşiyorsa dosya hiç okunmuyor; okunduysa hash eşleşmesi
  yeniden ayrıştırmayı atlıyor), `Rebuild`, `Search`, `ByFingerprint`, `Count`.
- Bozuk indeks sessizce siliniyor ve yeniden üretiliyor — türetilmiş veri.
- Arama FTS5 sözdizimi kabul etmiyor: her terim tırnaklı ifadeye çevrilip AND'
  leniyor, böylece kullanıcının yazdığı hiçbir şey sözdizimi hatası veremiyor.

#### Kabul kriteri ölçümü (10.000 kayıt, Windows 11 / go1.27)

| Ölçüm | Süre |
|---|---|
| Tam indeks üretimi | **1.911 s** |
| Değişiklik yokken yeniden tarama | **28 ms** |
| Arama (bm25 sıralı, 20 sonuç) | **6.5 ms** |
| Parmak izi araması | **33 µs** |
| İndeks dosyası | 8.6 MB |

Enjeksiyon yolundaki sayı 33 µs; hook bütçesi açısından sorun yok.

#### Ölçüm testinin yakaladığı hata

İlk koşuda 10.000 kaydın **4.020'si okunamadı**. Sebep: tamamı rakamdan oluşan
bir değer (ör. `fingerprint: 0000000000000000`) tırnaksız yazılınca tamsayı
olarak geri okunuyordu — baştaki sıfırlar gidiyor, kayıt bozuluyordu. Aynı şey
tamamı rakam olan bir ULID'nin de başına gelebilirdi.

İki taraflı düzeltildi: **yazıcı** geri okunduğunda sayı ya da bool olacak bir
dizeyi tırnaklıyor, **okuyucu** her skalerin metnini olduğu gibi saklıyor, yani
elle yazılmış tırnaksız dosya da yazarının kastettiği gibi okunuyor. İki
regresyon testi eklendi, `docs/record-format.md` ve ADR-0017 güncellendi.

#### İkinci bulgu: çapraz derleme aslında SQLite'ı sınamıyormuş

`scripts/crosscheck.sh` yalnızca `cmd/forgelore`'u derliyordu ve o paket henüz
`internal/store`'u import etmiyor — yani altı hedefteki "başarılı" çıktı
SQLite'a hiç dokunmamıştı. Betik artık her hedef için önce `go build ./...`
çalıştırıyor. Düzeltilmiş hâliyle **altı hedefin hepsi** `CGO_ENABLED=0` ile
derleniyor; ADR-0002'nin asıl vaadi ilk kez gerçekten doğrulandı.

#### gitleaks

Vendor ağacı taramaya 9 yanlış pozitif sokuyordu (hepsi SQLite'ın Go'ya
çevrilmiş tek bir üretilmiş dosyasında). `vendor/` allowlist'e alındı;
gerekçe `.gitleaks.toml` içinde yazılı.

### Karar — vendor depoya girdi

`go mod vendor` 10 modül, 2007 dosya, **135 MB** üretiyor (69 MB libc, 57 MB
sqlite; ikisi de her platform için üretilmiş C→Go çevirileri). ADR-0003 bunların
depoda olmasını söylüyor ama ADR yazıldığında bu sayı bilinmiyordu.

Karar: **girsin, K3 ne diyorsa o.** Kilitli bir kararı rahatsız ettiği için
değiştirmek disiplini aşındırır. Bedeli bilinerek kabul edildi.

İyi haber: çalışma dizini 135 MB ama git paketi **22.3 MiB**. Klonun ödediği
rakam bu — Go kaynağı iyi sıkışıyor. `.gitattributes` vendor ağacını satır sonu
dönüşümünden muaf tutuyor, böylece ağaç upstream'in yayımladığı baytlarla birebir
kalıyor; yerelde yeniden yazılmış bir vendor ağacı kimsenin incelediği şey
değildir.

Binary boyutu: şu an 1.6 MB, ama `cmd/forgelore` henüz `internal/store`'u
import etmiyor. SQLite bağlanınca gerçek boyut ~10 MB olacak (gösterge:
`internal/store` test binary'si 12 MB). Faz 2'de CLI indeksi kullanmaya
başlayınca kesin rakam ölçülecek.

---

## Faz 2 — Hata parmak izi ve CLI

**Durum:** Tamam. Adım 1–6 bitti, kabul kriteri elle denendi ve karşılandı.

### 2026-09-30 — Adım 2: hata örnekleri

Plan bu adımı `[SEN]` işaretlemişti; kullanıcı bana devretti. Web'den örnek
toplamak yerine hepsini **bu makinede gerçekten çalıştırarak** ürettim, çünkü
bir sitedeki çıktı "bu araç şunu basar" iddiasıdır, külliyatın amacı ise ne
bastığını gözlemlemek. Ayrıca Stack Overflow içeriği CC BY-SA lisanslı.

`scripts/capture-errors.sh` — yeniden üretilebilir toplama betiği. **56 dosya,
28 aile**, dört dil:

| Araç | Aile | Örnek |
|---|---|---|
| Go | 8 | undefined, tip uyuşmazlığı, kullanılmayan değişken, sözdizimi, bilinmeyen import, nil map panic, index panic, test hatası |
| Python | 8 | ModuleNotFound, NameError, TypeError, SyntaxError, AttributeError, ZeroDivision, FileNotFound, unittest |
| TypeScript | 4 | argüman tipi, bulunamayan ad, atanamayan tip, sözdizimi |
| Node | 3 | MODULE_NOT_FOUND, not-a-function, JSON parse |
| .NET | 5 | CS0103, CS0029, sözdizimi, tüm yollar dönmüyor, NullReference |

Her aile **iki varyant** halinde yakalandı (`a`, `b`): farklı isimli çalışma
dizini, farklı satır numarası. Yani aynı hatanın parmak izinin *aynı*, farklı
ailelerin parmak izinin *farklı* olması gerektiğini test edebilecek malzeme var.

Sürümler: go1.27.0, dotnet 10.0.103, Node v22.12.0, Python 3.14.7,
TypeScript 7.0.2.

#### Temizleme hatası ve düzeltmesi

İlk koşuda iki dosyaya gerçek kullanıcı adı sızdı. Sebep: Node yolu JavaScript
string'i olarak basıyor, yani ayraçlar çift (`C:\Users\<ad>\app.js`); tek
ayraç için yazılmış sed kuralı bunu atlıyordu. Kendi kontrolüm de hatalıydı,
sızıntıyı ilk taramada göremedim.

Düzeltme: önce çift ayraçlı biçim, sonra tek ayraçlı biçim, en sonda da hesap
adının kendisi — hangi biçimde geldiğinden bağımsız olarak. Külliyat sıfırdan
yeniden üretildi, üç yol biçiminin hepsi `dev` gösteriyor.

#### Ham malzemeden şimdiden görünen kalibrasyon soruları

- Go: `# alpha` / `# beta` paket satırı parmak izine girmeli mi? Aynı hata başka
  pakette aynı hafızayı hak ediyor mu?
- .NET: `Time Elapsed 00:00:04.57` her koşuda değişiyor — silinmeli. Ama aynı
  çıktıdaki `CS0103` kod numarası anahtar, silinmemeli.
- .NET aynı hata satırını iki kere basıyor (özet bölümünde tekrar); tekrarlar
  normalleştirilmeli mi?
- Node: `MODULE_NOT_FOUND` kodu var ama `requireStack` makineye özgü.
- Python traceback'inde `File "..."` satırlarının kaçı ayırt edici?

### 2026-10-02 — ikinci geliştirme makinesi: macOS

Proje ilk kez Windows dışında kuruldu ve doğrulandı. Makine: macOS 27.0.1,
Apple Silicon (arm64).

- Go 1.27.1, `~/.local/go` altına kuruldu: resmî `go.dev` tarball'ı, SHA-256'sı
  go.dev'in yayımladığı değerle karşılaştırıldı. Sudo kullanılmadı, sistem
  dizinlerine dokunulmadı; PATH satırı `~/.zshrc`'ye eklendi.
- gitleaks 8.30.1 betiğin kendi indirmesiyle geldi, checksum tuttu.
- `./scripts/ci.sh` tamamı yeşil, 1 dk 17 sn.

**ADR-0002'nin asıl vaadi ilk kez ikinci bir işletim sisteminde gözlendi.**
Şimdiye kadarki kanıt çapraz derlemeydi — derlenen ikililer çalıştırılmamıştı.
Artık saf Go SQLite ikinci bir platformda gerçekten koştu.

#### Aynı ölçüm, iki makinede (10.000 kayıt)

| Ölçüm | Windows 11 / go1.27.0 | macOS arm64 / go1.27.1 |
|---|---|---|
| Tam indeks üretimi | 1.911 s | 689 ms |
| Değişiklik yokken yeniden tarama | 28 ms | 56 ms |
| Arama (bm25, 20 sonuç) | 6.5 ms | 4.99 ms |
| Parmak izi araması | 33 µs | 34 µs |
| İndeks dosyası | 8.6 MB | 8.6 MB |

Enjeksiyon yolundaki rakam iki makinede de ~33 µs. Donanıma değil erişim
yoluna bağlı görünüyor (tek satır, indeksli arama) — hook bütçesi için
taşınabilir bir sayı. Yeniden tarama macOS'ta iki kat yavaş; 10.000 `stat`
çağrısının maliyeti, indeksin kendisi değil.

#### Bu makinede olmayanlar

`scripts/capture-errors.sh`'in istediği dotnet, node ve tsc yok; python var
ama 3.9.6 (külliyat 3.14.7 ile üretildi). Külliyat commit'li olduğu için Faz
2'nin testleri etkilenmiyor. Ama **külliyatı yeniden üretmek ya da ona yeni
aile eklemek bu makinede yapılamaz**: Windows makinesinde yapılmalı, ya da
araçlar buraya kurulmalı.

### 2026-10-02 — külliyat gözden geçirildi, bir boşluk bulundu

`go/unknown-import` ailesinin `a` ve `b` varyantları **bayt bayt aynı**. Sebep:
hata metni ne paket adını ne de değişen bir satır numarasını içeriyor
(`main.go:3:8` iki koşuda da aynı). Bu aile README'nin "varyantlar, parmak
izinin yok sayması gereken yönlerden farklıdır" iddiasını karşılamıyor —
varyant testine girerse hiçbir şey kanıtlamaz, sadece testi yeşil gösterir.

Kalan 27 ailenin hepsinde gerçek fark var. Açık karar: aile yeniden yakalansın
mı (farklı import yolu / farklı satır gerekir, yani Windows makinesi), yoksa
varyant testinden muaf tutulup gerekçesi mi yazılsın.

---

### 2026-10-02 — Adım 1 ve 3 tamam: `internal/fingerprint`

Kalibrasyon soruları kullanıcıya külliyattan gerçek örneklerle soruldu, sekiz
kararın hepsi onaylandı. Alınan kararlar:

| # | Konu | Karar | Gerekçe |
|---|---|---|---|
| 1 | Kapsam | Çıktıdan **hata olayları ayıklanır**, her birine ayrı parmak izi | Tüm çıktıyı tek hash'lemek, ilgisiz ikinci bir hata eklenince parmak izini değiştirir — hafıza susar |
| 2 | Tanımlayıcılar | **Girer** (`greet`, `'appendx'`, `'no-such-package-here'`) | En ayırt edici sinyal ve aynı hata tekrarladığında aynı kalıyor |
| 3 | Sayılar | Planın "sayıları sil" kuralı **konuma bağlı** daraltıldı | `CS0103`, `TS2345`, `41 != 42` silinirse hata kimliğini kaybeder; silinen yalnızca satır:sütun, hex ofset ve süreler |
| 4 | Go paket satırı | `# alpha` parmak izine **girmez** | "undefined: greet"in çözümü hangi pakette olduğuna bağlı değil |
| 5 | Yığın izleri | Kare satırları **atılır**, mesaj + hata kodu kalır | Node'un 20 kare'si kendi iç satır numaraları; Node sürümü değişince hepsi kayar |
| 6 | Python kaynak yankısı | Kaynak satırı, karet ve fonksiyon adı **atılır**, yalnızca istisna satırı kalır | Yoksa parmak izi o projenin o satırına bağlanır |
| 7 | Komut | `araç + fiil` jetonuna indirgenir | Ortam öneki ve bayraklar aynı hatayı bölmesin |
| 8 | `go/unknown-import` | Varyant testinden **muaf**, gerekçesi yazıldı | İki yakalaması bayt bayt aynı; test geçse bile hiçbir şey kanıtlamaz |

Yazılanlar: `internal/fingerprint/{fingerprint,command,extract}.go`,
`{fingerprint,corpus}_test.go`.

Pozitif eşleştirici yaklaşımı seçildi: **tanınmayan her satır gürültüdür.**
Böylece gürültü deseni listesi tutmak gerekmiyor — ilerleme çubukları,
ayraçlar, sayaçlar, süreler, bannerlar ve yığın kareleri zaten tanı değil.
Her desen satır başına çapalı; bu göründüğünden fazla iş yapıyor, çünkü
`at Object.<anonymous> (C:\work\app.js:1:13)` dosya, satır ve sütun taşıyor
ama bunlar satırın ilk jetonu değil, o yüzden konum eşleştiricisi onlara
erişemiyor.

#### Ölçüm (28 aile, 56 dosya)

| Ölçüm | Sonuç |
|---|---|
| Tanı çıkarılamayan dosya | **0 / 56** |
| Varyant eşleşmesi (kaçırma) | **27 / 27, 0 kaçırma** (1 aile muaf) |
| Aileler arası çakışma | **31 parmak izinde 0** |
| Test kapsamı | %100 |

#### Bu adımda netleşenler

- **Parmak izi uzunluğu normatif hâle geldi:** SHA-256'nın ilk 8 baytı, 16
  küçük harf hex. `docs/record-format.md` zaten örneklerinde bu biçimi
  kullanıyordu, artık yazılı.
- **Aynı metin, farklı sütun tek parmak izine düşüyor.** `tsc` bozuk dosyada
  üç tanı basıyor; ikisi farklı sütunda `':' expected.` ve sütun silinince
  birleşiyorlar. Doğru sonuç — eksik iki noktanın çözümü hangi sütunda
  olduğuna bağlı değil — ama test 3 değil 2 bekliyor, sebebi yazılı.
- **Mesaj içindeki konum bilgisi silinmedi.** Node'un
  `... at position 2 (line 1 column 3)` ifadesi duruyor. İki varyantta da
  aynı olduğu için külliyat bir şey söylemiyor; başka konumdaki bozuk JSON'ın
  *farklı* hata olması savunulabilir olduğu için olduğu gibi bırakıldı.
  Gerçek projelerden örnek gelince yeniden bakılacak.

#### Açık risk — komut jetonundaki fiil (karar 7'nin bedeli)

Aynı derleme hatası `go build`, `go test`, `go run` ve `go vet` altında
çıkabilir. Fiil parmak izine girdiği için bunlar **farklı** parmak izi
üretiyor: `go build` ile çözülmüş bir hata, aynı kod `go test` altında
patladığında bulunamaz. Külliyat bunu ölçemiyor, çünkü her aile tek komutla
yakalandı — ama `go/panic-*` ailesi `go run .`, derleme hataları
`go build ./...` ile yakalandı, yani senaryo uydurma değil.

Karar 7 onaylandığı gibi uygulandı. Öneri: **fiil çıkarılsın, yalnızca araç
(`go`, `dotnet`, `python`, `node`, `tsc`) kalsın** — mesajlar zaten
`go build` ile `go test` çıktısını birbirinden ayırmaya yetiyor. Kullanıcıya
bırakıldı, kendi başına değiştirilmedi.

---

### 2026-10-02 — karar 7 geri alındı: fiil parmak izinden çıktı

Önceki adımın "Açık risk" başlığı kullanıcıya iletildi, değişiklik onaylandı.
Komut jetonu artık `araç + fiil` değil **yalnızca araç**: `go build ./...`,
`go test ./...`, `go run .` ve `go vet ./...` hepsi `go`.

Gerekçe: aynı derleme hatası dört fiilin hepsinin altında çıkabiliyor ve
`go build` ile çözülmüş bir hata `go test` altında bulunamıyordu. Fiilin
ayırt edeceği şeyi mesaj zaten ayırıyor. Külliyatta kayıp yok: değişiklikten
sonra da 27/27 varyant eşleşiyor, 31 parmak izinde 0 çakışma.

İki test eklendi: dört fiilin aynı parmak izini ürettiği, ve araç değişince
(`python` ile `node` aynı `SyntaxError` metnini basınca) parmak izinin hâlâ
ayrıldığı.

### 2026-10-02 — Adım 4, 5, 6 tamam: CLI

Plan bu adım için iki soru soruyordu. Kullanıcı "önerilerinle ilerle" dedi,
alınan cevaplar:

**Komut adları ve bayraklar.** Plandaki adlar olduğu gibi: `init`, `record`,
`recall`, `search`, `show`, `index rebuild`, `doctor`, `stats`. Ortak
bayraklar `--json` ve `--dir`. Proje kökü git gibi yukarı yürüyerek bulunuyor
— ajan komutlarını kökte değil, derlemenin patladığı alt dizinde çalıştırır.

**`init` ne oluştursun.** `.forgelore/records/`, `.forgelore/local/` ve
`.forgelore/.gitignore`. Üçüncüsü asıl sebep: indeksini commit eden kullanıcı
her pull'da çakışan türetilmiş bir dosya commit etmiş olur, local kapsamını
commit eden de kendine ait olması gereken kayıtları paylaşmış olur.
**`config.yaml` oluşturulmadı** — henüz onu okuyan kod yok, ve hiçbir şeyin
okumadığı bir yapılandırma dosyası yazmak yükümlülüktür. ADR-0018 uygulanınca
eklenir.

`init` **AGENTS.md yazmıyor**, yazılacak tek satırı ekrana basıyor (plan
Adım 6). AGENTS.md projenin kendi dosyası; onu izinsiz düzenleyen araç
kurulmaz olur.

#### Uygulama sırasında alınan kararlar

- **Her okuma komutu önce indeksi senkronlar.** Değişiklik yokken maliyeti
  ölçülmüştü (28–56 ms / 10.000 kayıt); alternatifi bayat indeksten cevap
  veren ve bu yüzden daha az güvenilen bir araç.
- **`recall` hata döndürmüyor.** Eşleşme yoksa, hatta `.forgelore` hiç yoksa
  bile 0 ile çıkıyor ve hatayı yine de parmak izliyor. Bu enjeksiyon yolu;
  burada hata döndüren araç ajana kendisini çağırmamayı öğretir (ADR-0007).
- **`record` iki hata arasında seçim yapmıyor.** Çıktıda birden fazla hata
  varsa reddediyor ve parmak izlerini listeleyip `--fingerprint` istiyor.
  İlkini seçmek, aracın hangi tanıyı önce bastığına göre karar vermek olurdu.
- **Konumsal argümandan sonraki bayraklar da okunuyor.** Go'nun `flag`
  paketi ilk konumsal kelimede duruyor, yani `forgelore search sqlite --json`
  "sqlite --json" arıyordu. Elle denerken çıktı; `parseInterspersed` ile
  düzeltildi, testi var.

#### Kabul kriteri — elle denendi

Plan: "Hiçbir ajan entegrasyonu olmadan, bir ajan sadece AGENTS.md
yönlendirmesi ve shell ile hatayı sorgulayıp geçmiş çözümü bulabiliyor."

Derlenmiş binary ile, gerçekten patlayan iki Go projesi üzerinde koşuldu:

1. `demo` projesinde `go build ./...` → `undefined: greet`. `recall`:
   "1 error(s), 0 with something recorded".
2. Bir `fix` ve bir `dead_end` kaydedildi, parmak izi `cd023fb609411574`.
3. **Ayrı bir projede**, farklı modül adı, farklı paket, farklı dosya,
   farklı satır (`internal/svc/svc.go:7:2`) aynı hata üretildi.
4. O çıktı `demo` projesinde `--command "go test ./..."` ile — yani **kayıttan
   farklı bir fiille** — sorgulandı: ikisi de bulundu.

Yani parmak izi dosyadan, satırdan, paketten, modül adından ve fiilden
bağımsız çalışıyor. Karar 7'nin geri alınması 4. adımda doğrudan karşılığını
verdi.

#### Ölçümler

| Ölçüm | Değer |
|---|---|
| `recall` çağrı süresi | **~9.6 ms** (10 koşu / 96 ms, süreç başlatma dahil) |
| Host binary | **7.1 MB** |
| Test kapsamı | `fingerprint` %100, `cmd/forgelore` %84.4 |

Binary boyutu Faz 1'de "~10 MB olacak" diye tahmin edilmişti; `cmd/forgelore`
artık `internal/store`'u gerçekten import ettiği için ölçülen rakam bu.
`CLAUDE.md`'deki eski not düzeltildi.

#### gitleaks kendi testimi yakaladı

`TestRecordRedacts`'in fixture'ı `AWS_SECRET_ACCESS_KEY=` yanında yüksek
entropili bir dizeydi ve `scripts/gitleaks.sh` CI'ı kırdı. Değer sahteydi ama
tarayıcı haklıydı. Allowlist'e almak yerine fixture düşük entropili bir
yer tutucuyla değiştirildi — bir testi allowlist'e almak, aynı dosyadaki
gerçek bir sızıntıya karşı tarayıcıyı kör eder.

---

### Faz 2 — açık maddeler

- `[SEN/CLAUDE]` `go/unknown-import` yeniden yakalanmalı: farklı import yolu,
  kayan satır. Windows makinesi gerekiyor.
- Külliyatta gerçek projelerden hata yok — hepsi küçük ve kasıtlı örnekler.
  `testdata/errors/README.md` bunu zaten söylüyor; parmak izi kurallarının
  asıl sınavı o malzeme.
- `config.yaml` okunmuyor: ADR-0018 yazılı ama uygulanmadı, `init` de bu
  yüzden dosyayı oluşturmuyor. Bayrak > ortam > yerel > ekip > varsayılan
  önceliği hâlâ kodda yok.

---

## Faz 3 — Ölçüm altyapısı

**Durum:** Tamam. Adım 0–6 bitti. Kabul kriterinin sentetik yarısı karşılandı,
gerçek Claude Code oturumu yarısı `[SEN]` olarak açık.

### 2026-10-05 — dış gerçekler doğrulandı, plan buna göre değişti

Kural 7 gereği hafızadan değil kaynağından (resmî dokümantasyon, 2026-10-05).
Çıkan tablo planın Adım 3 varsayımını bozdu:

| Kanal | Kümülatif oturum kullanımı | Not |
|---|---|---|
| Hook stdin | **yok** | Ortak alanlar yalnızca `session_id`, `prompt_id`, `transcript_path`, `cwd`, `permission_mode`, `effort`, `hook_event_name` |
| Status line stdin | **`cost.total_cost_usd`** | "the estimated cost of all API calls in the current session" |
| Status line token alanları | **hayır** | `context_window.total_input_tokens` = "tokens currently in the context window, from the most recent API response" |
| OpenTelemetry | evet, `claude_code.token.usage` | Dosya exporter'ı yok; collector = daemon = K8 ihlali |

Üç sonuç:

1. **Enjeksiyonu yapan kod yolu kullanımı göremiyor.** Hook ile status line
   ayrı kanallar; ikisini birleştiren şey oturum kimliği.
2. **Token kümülatif değil, dolar kümülatif.** Karşılaştırma birimi dolar
   oldu. ADR-0019 olarak yazıldı.
3. **Döngü sayımı hiçbir şeye muhtaç değil.** "Aynı oturumda aynı parmak izi
   tekrar" tamamen kendi defterimizden ölçülüyor — kullanım okuyucusu olmayan
   ajanda bile bir karşılaştırma kalıyor. Bu planda yoktu, eklendi.

Bir de Faz 4'ü ilgilendiren bulgu: Claude Code'da **`PostToolUseFailure`**
diye ayrı bir olay var, planın "PostTool + başarısız çıkış" davranışı için
birebir. Faz 4 Adım 1'deki kanonik olay listesi gerçek adlarla güncellenmeli
(`UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `PostCompact` de var).

### Adım 0 (yeni) — ADR-0018 uygulandı

Plan Faz 3'te config öngörmüyordu ama Adım 2 ve 4 ayar istiyor.
`internal/config`: beş kaynak (bayrak > ortam > yerel > ekip > varsayılan),
her değer nereden geldiğini taşıyor ve `doctor` bunu basıyor — ADR-0018'in
açıkça istediği şey, çünkü beş kaynakla hiçbir dosya etkin değeri göstermiyor.

Tanımlı anahtarlar yalnızca **tüketicisi olanlar**: `measure.ledger`,
`measure.ab.control_percent`, `measure.ab.salt`, `report.period_days`.
`inject.budget_tokens` gibi Faz 4 anahtarları eklenmedi — hiçbir şeyin
okumadığı anahtar, aracın vermediği bir söz.

### Adım 1 — defter

`.forgelore/ledger/{injections,usage}-YYYY-AA.jsonl`, append-only.
**İndekse konmadı:** indeks türetilmiş ve okunamayınca siliniyor, defter ise
olanın tek kopyası. Kayıt da değil — kayıt birinin yazdığı hafıza, bu ölçüm.
`init` artık `ledger/` satırını da `.gitignore`'a koyuyor.

Satır tek `write` çağrısıyla yazılıyor: hook ile status line aynı anda
tetiklenebiliyor ve yarım satırların iç içe girmemesi gerekiyor.

### Adım 2 — A/B ataması

Durumsuz: `sha256(salt + "\x00" + session) % 100 < control_percent`.
Saklanan duruma güvenilemez çünkü her komut ayrı bir süreç (K8). Hash sayesinde
oturumun yüzüncü sorgusu birincisiyle aynı kola düşüyor; saklanacak dosya,
bozulacak durum, bayatlayacak şey yok. Varsayılan 0, yani deneme kapalı.

**Kontrol kolu bir miss'ten ayırt edilemiyor** — eşleşmeler yazdırılmadan önce
siliniyor, `--json` dahil. Defter eşleşmenin var olduğunu yine de kaydediyor;
farkı insan `report`'ta görüyor. Ajan hangi kolda olduğunu anlarsa deneme
bozulur.

### Adım 4, 5 — rapor ve anlamlılık

Bootstrap, t-testi değil: oturum maliyetleri çarpık, sıfırdan küçük olamıyor,
kollar küçük. Yeniden örnekleme yalnızca "bu oturumlar kendilerinin örneği"
varsayıyor. 2000 tur, medyan farkı üzerinde %95 aralık; aralık sıfırı
kapsıyorsa "anlamlı değil" yazılıyor, kol başına 3 oturumun altında hiç
hesaplanmıyor.

Kullanım yoksa maliyet karşılaştırması **nil** — rapor ölçmediği tasarrufu
iddia etmiyor, onun yerine nasıl bağlanacağını söylüyor.

#### İki test, iki bulgu

- **Aralık tekrarlanabilir değildi.** Seed sabitti ama kollar map üzerinde
  dönülerek kuruluyordu ve Go map sırasını rastgeleleştiriyor; aynı veriden
  farklı sıralı dilim, farklı yeniden örnekleme. Dilimler sıralanarak
  düzeltildi. Aynı defterden iki kez farklı sayı veren ölçüme kimse inanmaz.
- **Oturum kimliği olmayan satırlar döngü uyduruyordu.** Hepsi boş oturumda
  birleşince alakasız çağrılar tek bir döngüye benziyordu. Artık harcama
  tarafına sayılıyor, döngü ve maliyet tarafına sayılmıyor.

### Kabul kriteri

Plan: "Sentetik verilerle rapor doğru hesaplıyor. Gerçek bir Claude Code
oturumundan kullanım verisi okunabiliyor."

**Birinci yarı karşılandı.** Derlenmiş binary ile 24 oturumluk deneme: %50
kontrol, kontrol kolu aynı hataya üç kez toslayacak şekilde, tedavi kolu bir
kez. Rapor doğru çıkardı — döngü farkı 2.00 [2.00, 2.00], maliyet farkı
$0.5500 [$0.4900, $0.6100], ikisi de anlamlı. Ayrıca iki kolun aynı dağılımdan
geldiği durumda "anlamlı değil" dediği birim testiyle sabitlendi.

**İkinci yarı açık.** Gerçek bir oturumdan okumak için status line'ın
bağlanması gerekiyor (`forgelore usage --help-wiring` komutu tam ayarı
basıyor). Bu makinedeki mevcut oturum verisine erişmem PII gerekçesiyle
reddedildi; doğru karar. Faz 4 Adım 4 zaten gerçek örnek toplamayı `[SEN]`
işaretlemiş, aynısı burada da geçerli.

#### Ölçümler

| Ölçüm | Değer |
|---|---|
| Test kapsamı | `fingerprint` %100, `measure` %93.1, `config` %89.3, `cmd` %87.4 |
| Defter satırı | ~170 bayt |
| Bootstrap | 2000 tur, raporda gözle görülür gecikme yok |

### Sırada

Faz 4 — Claude Code adaptörü ve hook çalıştırıcı. Planın sorduğu üç soru
cevapsız: varsayılan enjeksiyon bütçesi, adayların onaya sunulma biçimi, hook
zaman aşımı. Yukarıdaki `PostToolUseFailure` bulgusu Adım 1'i etkiliyor.

### Faz 3 — açık maddeler

- `[SEN]` Status line bağlanıp gerçek bir oturumdan kullanım okunduğunun
  gösterilmesi. Kabul kriterinin ikinci yarısı.
- Token tahmini dört bayt/token; düzyazı için kaba, kod için yanlış. Gerçek
  cevap tokenizer ister (K3 yasaklıyor). Rapor bunu "estimated" diye
  etiketliyor, ama karşılaştırma zaten dolar üzerinden.
- `claude_code.token.usage` (OTel) kümülatif token veriyor. Prometheus
  exporter'ı tek seferlik bir süreçle `localhost:9464/metrics` üzerinden
  kazınabilir — daemon gerekmez. Dolar yerine token karşılaştırmak istenirse
  yol bu; ölçülmedi, denenmedi.

---

## Faz 4 — Claude Code adaptörü ve hook çalıştırıcı

**Durum:** Adım 1, 2, 3, 5, 6, 7, 8 tamam. **Adım 4 (gerçek olay örnekleri) açık**
ve kabul kriterinin bir yarısını kilitliyor.

### 2026-10-05 — doğrulananlar (resmî dokümantasyon)

| Konu | Bulgu |
|---|---|
| Olay adları | Planın listesi yanlış: gerçekte `UserPromptSubmit`, `PreToolUse`, `PostToolUse`, `PostCompact` ve ayrıca **`PostToolUseFailure`**; 30'dan fazla olay var |
| Çıktı | `hookSpecificOutput.additionalContext` ile bağlam verilir. Çıkış **2 = engelleme**, diğer sıfırdan farklı = kullanıcıya hata bildirimi |
| Zaman aşımı | `settings.json`'da hook başına saniye; varsayılan 600 s. Ama **`SessionEnd` için tüm hook'ların paylaştığı 1.5 s bütçe** var |
| Oturum durumu | Hook girdisi `scratchpad_dir` taşıyor (v2.1.257+) — oturuma özel geçici dizin |
| Eklenti | `.claude-plugin/plugin.json`, `hooks/hooks.json` `"hooks"` sarmalayıcısıyla. **`bin/` dizini olan eklentiyi claude.ai kurmuyor** |

**Doğrulanamayan, ve kritik olan:** başarısız bir Bash komutunda `tool_response`'un
şekli. Dokümantasyon örneği düz dize gösteriyor, şema vermiyor; sıfırdan farklı
çıkışta `PostToolUse` mi `PostToolUseFailure` mi tetiklediği de yazmıyor.

### Alınan kararlar

Yedi öneri kullanıcı tarafından onaylandı. ADR-0020 ikisini kayda geçiriyor.

| # | Karar | Gerekçe |
|---|---|---|
| 1 | Bütçe 500 token (`inject.budget_tokens`) | Dizin her oturumun ödediği sabit maliyet |
| 2 | Aday asla otomatik kayıt olmaz; `forgelore review` ile onaylanır | Komutun düzelmesi, birinin nedenini anladığı anlamına gelmiyor |
| 3 | Kendi son teslimimiz 500 ms (`hook.deadline_ms`); dış guard 5 s, SessionEnd 1 s | 1.5 s paylaşımlı bütçe |
| 4 | **Dört anlamsal olay**, ajanın yaşam döngüsü değil | Mirror edersek kanonik model her ajanla büyür |
| 5 | Eşlemeler **JSON** | ADR-0017 lehçesi iç içe harita kaldırmıyor; eşleme doğası gereği iç içe |
| 6 | Oturum durumu `scratchpad_dir`, yoksa `.forgelore/cache/sessions/` | Her hook ayrı süreç (K8) |
| 7 | Örnekler tek kullanımlık projede üretilecek | Senin oturum geçmişini okumamak için |

### Belirsizliği koda değil veriye hapsettik

`tool_response`'un şekli bilinmediği için ikisini birden yaptık:

- **Alan yolu bir liste.** `["tool_response.stderr", "tool_response.stdout", "tool_response"]` —
  ilk çözülen kazanıyor. Alan bugün dize, yarın nesne olsa da metin yerinde kalıyor.
- **Yalnızca `PostToolUseFailure` başarısızlık sayılıyor.** `PostToolUse` her
  durumda "başarılı" olarak eşleniyor. Çıkarım deseni uydurmadık: o olay
  tetiklenmiyorsa bedel *kaçırılmış bir enjeksiyon*, uydurursak bedel *çalışan
  bir komuttan sonra verilen yanlış ipucu*. Sessizlik ucuz olan hata.

Her ikisi de eşleme JSON'ında, yani gerçek örnekler gelince **yeniden derleme
olmadan** düzeltilir. `verified_against` şu an boş ve testi bunu sabitliyor.

### Yazılanlar

- `internal/agent/` — kanonik olaylar, eşleme formatı, çevirici, oturum durumu.
  Gömülü `mappings/claude-code.json`.
- `cmd/forgelore/hook.go` — çalıştırıcı, davranışlar, adaylar, `review`.
- `plugin/` — manifest + `hooks/hooks.json` + README.
- Oturum kimliği dosya adına **hash'lenerek** giriyor: dışarıdan gelen bir
  değer yol bileşeni olacaksa traversal'i kendimiz kesiyoruz.

### Kabul kriteri

| Kriter | Durum |
|---|---|
| Bozuk indeks, silinmiş klasör, zaman aşımında ajan etkilenmiyor | ✅ A–E senaryosu elle koşuldu, hepsi çıkış 0 |
| Gecikme ölçülmüş | ✅ defterde `hook_ms`, raporda p50/p95; ölçülen **p50 1 ms** |
| Gerçek Claude Code oturumunda ipucu enjekte ediliyor | ❌ **açık** |

Elle koşulan tam döngü (gerçek binary): SessionStart dizini → bilinmeyen hatada
sessizlik → komut düzelince aday → `review --accept` → **başka oturumda** aynı
hatada ipucu → kontrol kolunda sessizlik.

### 2026-10-05 — Adım 4 tamam: gerçek yükler eşlemeyi yalanladı

`claude` CLI kuruldu: **2.1.289**, resmî yerel yükleyiciyle
(`https://claude.ai/install.sh` → `downloads.claude.ai`). Betik çalıştırılmadan
önce okundu: `$HOME` altına kuruyor, SHA-256'yı manifest'ten doğruluyor, sudo
istemiyor. Node/npm gerekmedi.

**Oturum açma tuzağı:** CLI masaüstü uygulamasından **ayrı** kimlik doğruluyor.
Keychain'de `Claude Code-credentials` vardı ama o uygulamanın; CLI
`loggedIn: false` diyordu. Doğru komut `claude auth login`. İlk koşu login
olmadan yapıldığı için model turu hiç olmadı ve yalnızca oturum olayları
yakalandı — o koşudan çıkarılan "`scratchpad_dir` yok" sonucu **yanlıştı**,
artefakttı.

`scripts/capture-agent-events.sh` tek kullanımlık dizinde çalışıyor, kendi
`settings.json`'ıyla, kullanıcının yapılandırmasına dokunmadan; biri patlayan
biri çalışan iki komut koşturup dört olayı da yakalıyor, hesap adını `dev` ile
değiştiriyor ve kendi işini denetliyor.

#### Asıl bulgu: eşleme yanlıştı ve sessizce yanlıştı

Başarısız Bash komutunda **`tool_response` alanı hiç yok.** Çıktı üst düzey
`error` alanında, çıkış koduyla birlikte:

```
"error": "Exit code 1\n# example.com/broken/cmd/app\ncmd/app/main.go:4:2: undefined: greet"
```

Dokümantasyondan türetilen eşleme `tool_response.stderr`, `.stdout` ve
`tool_response`'a bakıyordu; üçü de yok. Olay çıktısız üretiliyor, parmak izi
boş dönüyor, **enjeksiyon yolu sonsuza kadar sessiz kalıyordu** — ve hiçbir
yerde hata görünmüyordu. Kural 7'nin tam olarak engellemek istediği şey.

*Başarılı* komutta ise `tool_response` **var**, nesne olarak (`stdout`,
`stderr`, `interrupted`). Yani iki olay gerçekten farklı şekilde. Yol-listesi
biçimi tam bunun içindi: düzeltme `error`'ı listenin başına eklemekten ibaret,
**JSON'da, yeniden derleme yok**.

Yüklerin söylediği iki şey daha:

- `is_interrupt` / `tool_response.interrupted` — kullanıcının durdurduğu komut
  hata değil. Eşlemeye `skip_when` eklendi.
- `scratchpad_dir` **dört olayda da var**. Oturum durumu ajanın zaten
  temizlediği yere yazılıyor; `.forgelore/cache/sessions/` yedeği duruyor
  çünkü dokümantasyon alanın bulunmayabileceğini söylüyor.

#### Durum

- `testdata/agents/claude-code/2.1.289/` — dört gerçek yük, temizlenmiş.
- `verified_against: "2.1.289"`, ve bir test bu iddianın arkasında külliyat
  olmasını şart koşuyor.
- `withoutSamples` **boş**, ve boş kalması zorunlu: eşlemeye örneksiz olay
  eklenirse sözleşme testi kırılıyor.
- Gerçek yük binary'den geçirildi: hafıza boşken sessiz, doluyken ipucu,
  `is_interrupt` ile sessiz.
- `internal/agent` kapsamı %95.6.

ADR-0020'ye tarihli bir güncelleme eklendi. Doğrulanan şey eşlemenin içeriği
değil **biçimi**: yanılmanın bedeli bir JSON düzenlemesi oldu.

### Sırada

Faz 5 — kendi MCP sunucumuz (K4). Faz 4'ten devreden açık karar kalmadı.

### Faz 4 — açık maddeler

- `PostToolUse` başarı yükünde `noOutputExpected` ve `isImage` gibi alanlar
  var; hiçbiri kullanılmıyor. İhtiyaç doğarsa eşlemeye eklenir.
- `UserPromptSubmit` ve `PreCompact` hiç kullanılmıyor. Planın kanonik
  listesinde vardılar; dört olaylı modelde karşılıkları yok.
- Eklenti elle kurulup gerçek bir oturumda denenmedi. `plugin/` hazır ve
  hook'lar gerçek yüklerle test edildi, ama uçtan uca "eklentiyi kur, hatayı
  tekrarla, ipucunu gör" adımı yapılmadı.

---

## Faz 5 — MCP sunucusu

**Durum:** Tamam. Adım 1–6 bitti, kabul kriteri iki istemciyle karşılandı.

### 2026-10-05 — spesifikasyon planı çürüttü, ölçüm kararı verdi

Güncel revizyon **2026-07-28** ve `initialize` el sıkışmasını **kaldırmış**.
Protokol durumsuz: her istek kendi sürümünü ve istemci yeteneklerini
`_meta.io.modelcontextprotocol/*` içinde taşıyor, sunucu `server/discover`
uygulamak zorunda, bilinmeyen sürüm `-32022` ile reddediliyor. ADR-0004'ün
"initialize ile sürüm müzakeresi" kapsamı artık önceki dönemi tarif ediyor.

Çift dönem gerekip gerekmediğini **ölçtüm**, akıl yürütmedim. Kaydedici bir
sunucuya karşı Claude Code 2.1.289:

| Çalışma zamanı | Ne gönderiyor |
|---|---|
| varsayılan (v2) | `server/discover`, `_meta` ile `2026-07-28`; `initialize` hiç yok |
| `MCP_SDK_GENERATION=v1` | `initialize` → `2025-11-25`, `notifications/initialized`, sonra `_meta`'sız `tools/list` |

Hangisinin seçildiği feature flag'e bağlı, bizim elimizde değil. Spesifikasyonun
uyumluluk matrisi net: legacy istemci + modern-only sunucu = başarısız. Yani
**oturumların yarısında Forgelore görünmez olurdu.** ADR-0021 yazıldı,
ADR-0004'e tarihli güncelleme düşüldü.

### Yazılanlar

- `internal/mcp/` — satır ayraçlı JSON-RPC, çift dönemli yönlendirme, dört araç.
- `internal/candidate/` — adaylar `cmd`'den çıkarıldı; hem hook hem `propose`
  yazıyor. Aday artık `title` taşıyabiliyor, `review --accept` onu kullanıyor.
- `forgelore mcp` komutu.
- `testdata/mcp/claude-code/2.1.289/{modern,legacy}.jsonl` — gerçek istemci
  istekleri, testte yeniden oynatılıyor.

Sunucu tek bir bağlantı durumu tutuyor: "bu süreç `initialize` ile açıldı"
bayrağı. Modern yol onu hiç okumuyor — durumsuzluk şartını dürüst tutan şey bu.

### Gerçek istemcinin yakaladığı uyum hatası

Elle sürdüğüm testlerin hepsi geçiyordu; `claude mcp list` ise bağlanmayı
reddetti:

```
Invalid result for tools/list: expected number, received undefined, path: ttlMs
```

**`cacheScope` gönderip `ttlMs` göndermemiştim.** İstemci kendi şemasına göre
doğruluyor ve araç listesini tümden atıyor. Yakalanan trafiği yeniden oynatmak
bunu göremezdi, çünkü külliyat isteği tutuyor, istemcinin cevap şemasını değil.
Kabul kriterinin "gerçek istemci bağlanabiliyor" demesinin sebebi tam olarak bu.

### İkinci bulgu: `command` atlanırsa hafıza kaçıyor

`recall_error`'ı `command` olmadan çağırınca parmak izi farklı çıkıyor, çünkü
araç adı parmak izinin parçası (Faz 2, karar 7). Aynı çıktı, `command` ile
`cd023fb609411574` ve eşleşiyor; onsuz `8bd855593ebb1628` ve eşleşmiyor. Araç
açıklaması artık bunu açıkça söylüyor ve bir test doğru kalmasını sağlıyor.

### Kabul kriteri — iki istemci

| İstemci | Sonuç |
|---|---|
| **Claude Code 2.1.289** | `claude mcp add` ile bağlandı (`✔ Connected`), model `recall_error`'ı çağırıp kaydedilmiş düzeltmeyi döndürdü |
| **@modelcontextprotocol/inspector** (resmî, CLI) | `tools/list`, `search`, `recall_error` çalıştı |

İkinci istemci için Node LTS v24.21.0 `~/.local/node` altına kuruldu
(nodejs.org tarball'ı, `SHASUMS256.txt` ile doğrulandı). Projenin kendi
bağımlılık kuralını etkilemiyor — Forgelore'un derlenmesi ya da test edilmesi
için gerekmiyor, yalnızca ikinci istemciyi koşturmak için.

### Sırada

Faz 6 — ekip akışı. **Adım 1 (`review`) zaten yazıldı**, Faz 4'te adaylar için
gerekmişti. Kalanı: `promote`, commit öncesi maskeleme kontrolü, tekrar
tespiti, ve `[SEN]` iki kişilik bir hafta denemesi.

### Faz 5 — açık maddeler

- `tools/list` sayfalama desteklemiyor (`cursor` yok sayılıyor). Dört araçla
  gereksiz; araç sayısı artarsa gerekir.
- MRTR (`InputRequiredResult`), abonelikler ve `notifications/tools/list_changed`
  uygulanmadı. Araç seti derleme zamanında sabit, hiçbirine ihtiyaç yok.
- Yalnızca stdio. Streamable HTTP ve yetkilendirme kapsam dışı (ADR-0004).

---

## Faz 6 — Ekip akışı

**Durum:** Adım 1–5 ve 7 tamam. Adım 6 (bir haftalık deneme) `[SEN]`.

### 2026-10-05 — koda bakınca plan kısaldı

| Adım | Durum |
|---|---|
| 1 `review` | **Zaten yazılmıştı**, Faz 4'te adaylar için gerekmişti |
| 2 `promote` | `store.Promote` Faz 1'den beri vardı; eksik olan CLI ve `tainted` kapısıydı |
| 3 `check` | Redaksiyon `Put`'ta zaten çalışıyordu; açık olan elle düzenlenmiş/merge'den gelmiş dosyaydı |
| 4 tekrar tespiti | **Bir boşluk açığa çıktı** (aşağıda) |

#### `superseded_by` hiç yazılmıyormuş

Kayıt formatı schema 1'den beri tanımlıyor, ADR-0016 davranışını belirliyor,
indeks saklıyor, `stats` sayıyor, `recall` superseded kayıtları enjekte
etmiyor — ama **hiçbir kod onu set etmiyordu.** Formatın yazılmamış yarısıydı.
`forgelore dedupe --supersede <eski> --by <yeni>` bunu kapatıyor.

### Alınan kararlar

| # | Karar | Gerekçe |
|---|---|---|
| 1 | `check` komutu + `init`'in bastığı hook satırı; `init --with-git-hook` isteğe bağlı kurar, mevcut hook'u **ezmez** | AGENTS.md'de aldığımız tavırla aynı: projenin dosyası projenindir |
| 2 | gitleaks zorunlu değil — kendi desenlerimiz her zaman, gitleaks varsa ek olarak; `doctor` hangisinin koştuğunu söyler | Kullanıcının duymadığı bir araç olmadan koşmayan kontrol, atlanan kontroldür |
| 3 | Tekrar tespiti yeni `dedupe` komutunda | |
| 4 | `tainted` için `--force-tainted` | Promote etmek, kimsenin yazmadığı bir şeyi paylaşmak demek (ADR-0013) |
| 5 | `check` yalnızca ekip kapsamını tarar | Local zaten makineden çıkmıyor; birinin kendi notlarını ona rapor etmek bu komutun işi değil |
| 6 | Deneme protokolü `docs/team-trial.md` | |

### Kabul kriteri — ikisi de ölçüldü

**Eşzamanlı kayıtta merge çakışması yok.** İddia edilmedi, koşuldu: bare bir
origin, iki klon, her birinde bir kayıt, sonra merge. Çakışma yok, iki hafıza
da sağ, ve `git ls-files` `ledger/`, `cache/`, `local/` içermiyor. Sebebi
tasarımda: her kayıt ULID adlı ayrı dosya, paylaşılan değişken dosya yok.

**Sır içeren kayıt commit edilemiyor.** Gerçek binary, gerçek git hook:
temiz kayıt commit oldu; dosyaya elle `DEPLOY_TOKEN=...` eklenince hook
commit'i engelledi ve `git log` 1'de kaldı.

Burada iki tarayıcının **birbirini tamamladığı** görüldü: gitleaks bu değere
"no leaks found" dedi (düşük entropi), yakalayan bizim `assigned-secret`
desenimiz oldu. Tersi de doğru — Faz 2'de gitleaks benim yüksek entropili
test fixture'ımı yakalamıştı, bizim desenlerimiz yakalamazdı.

### Düzeltme: Node zaten kuruluymuş

Faz 5'te ikinci MCP istemcisi için Node LTS v24.21.0'ı `~/.local/node` altına
kurmuştum. Makinede **zaten Node v22.23.3 vardı**, `~/.local/opt/node`
altında, sadece PATH'te değil — `~/.local/bin`'deki `pnpm`/`yarn` sembolik
bağları oraya işaret ediyormuş. `command -v node` boş dönünce kurulu değil
diye sonuçlandırmıştım, yeterince bakmamışım. Inspector mevcut Node ile de
çalışıyor; kurduğum kopya kaldırıldı.

Aynı şekilde gitleaks de `~/.local/bin/gitleaks`'te zaten varmış — `check`'in
gitleaks yarısı bu makinede gerçekten koşuyor.

### Sırada

Faz 7 — diğer ajanlar. Faz 6'dan devreden tek şey `[SEN]` bir haftalık deneme.

### Faz 6 — açık maddeler

- `[SEN]` İki kişi, bir hafta. Protokol `docs/team-trial.md`'de.
- `dedupe` yalnızca parmak izine bakıyor. Aynı şeyi söyleyen ama farklı
  parmak izine sahip iki kayıt görünmüyor; metin benzerliği aramıyoruz.
- `check` ekip kayıtlarını tarıyor, commit edilecek **diff'i** değil. Bir sır
  kayda girip aynı commit'te silinirse yakalanmaz; bu hâliyle depoda duran
  hâli kontrol ediyor, ki asıl mesele de o.

---

## Faz 7 — Diğer ajanlar

**Durum:** Copilot CLI doğrulandı (1.0.91, üç gerçek yük). Codex CLI'nin
eşlemesi yazıldı ama **doğrulanmadı** — oturum açılmasını bekliyor.
`docs/compatibility.md` yayımlandı.

### 2026-10-05 — dördü de A adayı çıktı

| Ajan | Hook yapılandırması | Olay adları | Hata olayı |
|---|---|---|---|
| Codex CLI | `~/.codex/hooks.json` veya `config.toml`; **varsayılan kapalı** (`features.hooks`), proje kapsamı güven istiyor | `PreToolUse`, `PostToolUse`, `SessionStart/End`… (12) | yok |
| Copilot CLI | `.github/hooks/*.json`, `~/.copilot/hooks/` | camelCase: `postToolUse`, **`postToolUseFailure`** | var |
| Gemini CLI | `.gemini/settings.json` | **`BeforeTool`/`AfterTool`** | doğrulanmadı |
| Cursor | `.cursor/hooks.json` | `beforeShellExecution`, `afterFileEdit`, `afterMCPExecution` | yok |

Planın A/B/C seviyelendirmesi beklenenden az ayırt edici: 2026 sonunda büyük
dördün hepsinde hook **ve** MCP var. Bu yüzden tabloda seviye, **reklam
edileni değil doğrulananı** gösteriyor.

### Eşleme formatında gerçek bir boşluk çıktı

Copilot CLI bağlamı **üst düzey `additionalContext`** ile istiyor; Claude Code
`hookSpecificOutput` sarmalayıcısıyla. Eşleme formatında cevap şekli yoktu —
`cmd/forgelore/hook.go` Claude Code'un şeklini kodda sabitlemişti. K5'in
engellemesi gereken şey tam buydu ve ikinci ajan gelene kadar görünmedi.

Format `response` bölümü kazandı:

```json
"response": { "context_path": "hookSpecificOutput.additionalContext",
              "event_name_path": "hookSpecificOutput.hookEventName" }
```

Copilot için yalnızca `"context_path": "additionalContext"`. Noktalı yoldan iç
içe nesne kuruluyor. Artık yeni bir ajan için Go değişikliği gerekmiyor —
gerekirse bu, formatın bir eksiği demektir ve öyle kaydedilmeli.

### Yazılanlar

- `internal/agent/mappings/{codex-cli,copilot-cli}.json` — ikisi de
  `verified_against: ""`, ve bir test boş olmayan bir iddianın arkasında
  külliyat olmasını şart koşuyor.
- `scripts/capture-agent-events.sh` artık **profil alıyor**: ajan başına
  binary adı, sürüm komutu, hook config yolu ve biçimi, headless çağrı.
  Codex için `features.hooks=true` ve hook güvenini atlama da profilde.
- `docs/compatibility.md` — seviye, doğrulanan sürüm, ve her ajan için neyin
  test edilmediği.
- Belgelenen yükleri sabitleyen testler. **Geçmeleri doğrulama değil**;
  eşleme yazıldığı gün neye inanıldığının kaydı, ki ilk gerçek yük bir fark
  olarak görünsün, gizem olarak değil.

### Kurulanlar

Codex CLI 0.160.0 ve GitHub Copilot CLI 1.0.91, mevcut Node ile
`~/.local` altına. İkisi de **oturum açmamış** — `codex login`,
`copilot login` senin hesaplarını istiyor.

### 2026-10-05 — Copilot CLI doğrulandı, iki tahmin de yanlış çıktı

Oturum açıldı, `./scripts/capture-agent-events.sh copilot-cli` üç gerçek yük
topladı (1.0.91). Dokümantasyondan türettiğim eşlemenin **iki varsayımı da**
yanlıştı, ve ikisi de Forgelore'u sessiz bırakırdı.

**1. Yük kendi olayını adlandırmıyor.** Hiçbir yazımıyla `hookEventName` yok:
yük `sessionId`, `timestamp`, `cwd` ve olaya özgü alanlardan ibaret. Olay
yalnızca hangi hook girdisinin tetiklendiğinden belli. Eşleme formatı boş
`event_name` kazandı — "adı çağıran verir" — ve hook komutu taşıyor:
`forgelore hook --adapter copilot-cli --event postToolUse`.

**2. Patlayan komut, patlayan araç değil.** `go build`'in 1 ile çıkması
`postToolUse` olarak geliyor, `toolResult.resultType: "success"` ile, ve
kabuk çıkış kodu sonuç metninin sonunda:

```
<shellId: 0 completed with exit code 1>
```

Eşleme artık `output_matches` ile `completed with exit code [1-9]` arıyor.
Başarıda aynı yerde `exit code 0` var; ikisi de test edildi.
`postToolUseFailure` var ama *aracın* başarısızlığını anlatıyor, ki bozuk bir
derleme onu tetiklemiyor — örneği yok, `withoutSamples`'da gerekçesiyle
kayıtlı.

Üçüncü bir şey: Copilot proje içindeki `.github/hooks`'u geçici dizinde
okumadı. Kullanıcının gerçek `~/.copilot`'una yazmak yerine **`COPILOT_HOME`**
ile ajana kendi evi verildi; profil bunu yapıyor.

Yakalama betiği de değişti: olay adı artık hook girdisinden **argüman olarak**
geliyor, yükten ayrıştırılmıyor. Ayrıştırma zaten yalnızca Claude Code için
çalışıyordu.

### 2026-10-05 — Codex atlandı

Üç oturum açma denemesi: tarayıcı akışı iki kez, `--device-auth` bir kez.
Cihaz kodu kabul edilmedi; muhtemelen hesabın Codex CLI erişimiyle ilgili,
yani bizim tarafımızda bir iş değil. Kullanıcı atlamaya karar verdi.

Codex CLI'nin eşlemesi depoda duruyor, `verified_against: ""` ile, ve
`docs/compatibility.md` onu **"A, doğrulanmamış"** diye gösteriyor. Copilot'ta
dokümantasyondan türetilen iki varsayımın da yanlış çıktığı düşünülürse, bu
eşlemenin de yanlış olduğu varsayılmalı. Tabloda öyle duruyor.

Faz 4 ve Copilot'un dersi Codex'te de beklenmelidir: dokümantasyondan yazılan
eşleme yanlış çıkacak, ve yanlışın bedeli bir JSON düzenlemesi olacak.

### Faz 7 — açık maddeler

- `[SEN]` Codex CLI doğrulanmadı — oturum açılamadı, atlandı. Erişim
  olduğunda `./scripts/capture-agent-events.sh codex-cli` ve aynı döngü.
- Gemini CLI ve Cursor ayrı turlar. Cursor'ın olay modeli (kabuk öncesi /
  dosya sonrası) dört anlamsal olaya en uzak duran; `command_failed`
  karşılığı olup olmadığı bakılmalı.
- Plan `docs/uyumluluk.md` diyor; dosya `docs/compatibility.md` olarak
  yazıldı. Faz 0 kamuya açık doküman adlarını İngilizceye çevirmişti, planın
  bu satırı güncellenmemiş.
- Kullanım okuyucusu hâlâ yalnızca Claude Code'da. Diğerlerinde karşılığı
  görünmüyor; ADR-0012'nin dürüstlük sınırı zaten kapsıyor.

---

## Faz 8 — Dağıtım ve yayın

**Durum:** Adım 1–4 tamam. Adım 5 ve 6 `[SEN]`.

### Alınan kararlar

| # | Karar | Gerekçe |
|---|---|---|
| 1 | Release **etiketle tetiklenen CI**'da derleniyor, **taslak** çıkıyor | Yerelde derlemek imzalamanın önünü kapatıyor; taslak birinin okumasını zorunlu kılıyor |
| 2 | npm sarmalayıcı **ertelendi** | Plan "isteğe bağlı" diyor; `install.sh` aynı ihtiyacı karşılıyor ve npm yeni bir yayın yüzeyi açıyor |
| 3 | `~/.local/bin`'e kuruyor, PATH'te değilse **söylüyor** | Bugün `codex`/`copilot` ile tam bunu yaşadık: ikili oradaydı, kabuk görmüyordu |
| 4 | İlk sürüm `v0.1.0` | |
| 5 | İlk release'te **imzalama yok** | cosign keyless CI kimliğine bağlı; karar 1 sayesinde sonraki release'te ucuz |
| 6 | `docs/versioning.md` | Faz 0'ın İngilizceye çevirme kararı; plan `surumleme.md` diyor, `uyumluluk.md` gibi güncellenmemiş |

### Yazılanlar

- `scripts/release.sh` — kirli ağacı reddediyor, `ci.sh`'i koşuyor, altı
  hedefi etiket sürümüyle derliyor, `SHA256SUMS` yazıyor ve **kendi ürettiğini
  doğruluyor**. Yayımlamıyor; komutu basıyor.
- `scripts/install.sh` — platformu saptıyor, indiriyor, doğruluyor, kuruyor.
  `FORGELORE_BASE_URL` ile başka bir kaynağa yönlendirilebiliyor, kabul testi
  bunu kullanıyor.
- `.github/workflows/release.yml` — `v*` etiketinde, action'lar SHA ile
  sabitli, `contents: write`, taslak release.
- `docs/versioning.md` — üç bağımsız sürüm numarası ve 1.0 için gereken üç
  şey.

#### Eşleme formatı neden hâlâ 1

Faz 7'de iki kez genişledi (cevap şekli, çağıran-verir olay adı) ama ikisi de
**ekleme**ydi: ikisini de kullanmayan bir eşlemeyi eski Forgelore hâlâ
okuyabiliyor. Numara, var olan bir alanın anlamı değiştiğinde 2 olur.
`versioning.md` bunu yazılı hâle getirdi.

### Kabul kriteri — ölçüldü

Plan: "Temiz bir makinede tek komutla kurulum çalışıyor, checksum
doğrulanıyor."

`dist/` yerel bir HTTP sunucusundan servis edildi ve `install.sh` boş bir
`HOME` ile, kırpılmış `PATH` ile koşuldu: indirdi, doğruladı, kurdu,
`forgelore v0.1.0-test darwin/arm64 go1.27.1` bastı, ve `~/.local/bin`'in
PATH'te olmadığını söyledi.

**Asıl test reddetme tarafıydı:** ikiliye bir bayt eklendi, aynı koşu
checksum uyuşmazlığını iki hash'i de yazarak bildirdi ve **hiçbir şey
kurmadı**. GitHub olmadan gerçek test; gerçek release sonrası tekrarlanacak.

### Sırada

Faz 9 — kayma dedektörü ve bakım.

### Faz 8 — açık maddeler

- `[SEN]` İlk release'i yayımla: `./scripts/release.sh v0.1.0` çıktısını oku,
  etiketle, push et. Workflow taslağı üretir, yayımlamak sende.
- `[SEN]` Forgeprint marketplace kaydı, ayrı depoda.
- İmzalama sonraki release'e bırakıldı.
- npm sarmalayıcı yazılmadı.
- `release.sh` `crosscheck`'i iki kez koşuyor (biri `ci.sh` içinde, biri
  etiket sürümüyle). Birkaç saniye; bölmeye değmedi.

---

## Faz 9 — Kayma dedektörü ve bakım

**Durum:** Tamam. Adım 1–4 bitti, kabul kriteri karşılandı.

### Planın 1. adımı kaymayı göremezdi

Adım 1 "son sürümleri kurar ve **toplanan olay örnekleriyle** sözleşme
testlerini çalıştırır" diyordu. Ama toplanan örnekler ajan güncellendiğinde
değişmiyor — o testler bizim değişikliklerimizi yakalar, ajanınkileri değil.
Yeni sürümü kurmak hiçbir şeye dokunmaz.

Gerçek kayma ancak canlı ajandan **yeniden yakalayarak** görülür, o da oturum
açmış bir hesap ister. CI'da kimlik yok ve olmamalı. İkiye bölündü:

| | Ne görür | Nerede koşar |
|---|---|---|
| `drift-versions.sh` | "Yeni sürüm çıktı, kimse bakmadı" | CI, haftalık, kimliksiz |
| `drift-payloads.sh` | Alan gerçekten taşındı mı | Bakımcının makinesi, ajan oturum açmış |

Üç sürüm kaynağı da kimlik doğrulamadan sorgulanabiliyor: Claude Code
`downloads.claude.ai/.../latest`, Copilot ve Codex npm registry.

`drift-payloads.sh` **değerleri değil alan yollarını** karşılaştırıyor. Oturum
kimliği ve zaman damgası her koşuda değişir ve hiçbir şey anlatmaz; kaybolan
ya da beliren bir alan ise hikâyenin tamamı, çünkü eşlemenin bağlı olduğu şey
tam olarak bir alan yolu.

### Kabul kriteri — ve açtığı kör nokta

Plan: "Bir eşleme alanı bilerek bozulduğunda kayma dedektörü bunu yakalıyor."

İki bozma denendi:

1. `tool_input.command` → `tool_input.cmd`: **yakalandı**, dosya adıyla
   birlikte ("the command did not resolve").
2. `common.event_name` → `hookEvent`: **yakalanmadı.**

İkincisi gerçek bir kör noktaydı. Replay olay adını yedek olarak geçiriyordu
ve bu, eşlemenin bozuk `event_name` yolunu maskeliyordu — test geçerdi, ajanın
önünde çalışmazdı, çünkü Claude Code hook'u `--event` göndermiyor. Yedek artık
yalnızca eşleme "adı çağıran verir" (`event_name: ""`) dediğinde veriliyor.
Düzeltmeden sonra ikisi de yakalanıyor.

Canlı kontrol: `./scripts/drift-payloads.sh copilot-cli` gerçek ajana karşı
koşuldu, üç olayda da "unchanged" dedi.

### Yazılanlar

- `scripts/drift-versions.sh` — "hiç doğrulanmadı" ayrı bir durum olarak
  raporlanıyor, sessizce düşmüyor; Codex şu an onu tetikliyor.
- `scripts/drift-payloads.sh` + `capture-agent-events.sh`'e
  `FORGELORE_CAPTURE_OUT` (külliyatı ezmeden yakalamak için).
- `.github/workflows/drift.yml` — haftalık, **ajan başına tek issue**,
  başlıktan bulup yorum ekliyor. Her pazartesi yeni issue açan dedektör,
  insanların filtrelediği dedektördür.
- `doctor` artık kurulu ajan sürümlerini `verified_against` ile
  karşılaştırıyor ve **uyarıyor** — çıkış kodunu değiştirmiyor.
- `docs/mappings.md` + `CONTRIBUTING.md`'nin adaptör bölümü güncellendi (eski
  `adapters/` yolunu gösteriyordu).

İkisi de bilerek `ci.sh`'in dışında: biri ağ, diğeri oturum açmış ajan
istiyor, `ci.sh` ise her makinede çevrimdışı koşabilmeli.

### Sırada

Dokuz fazın tamamı bitti. Kalan işler `[SEN]`:

- İlk release (`./scripts/release.sh v0.1.0`, etiket, push).
- Forgeprint marketplace kaydı.
- İki kişilik bir haftalık ekip denemesi (`docs/team-trial.md`).
- Codex CLI doğrulaması, erişim olduğunda.
- 11 commit hâlâ yerelde; release workflow'u etikete bağlı, yani önce
  `git push origin main` gerekiyor.

### Faz 9 — açık maddeler

- Sürüm kayması yalnızca üç ajanı biliyor. Dördüncüsü eklenince
  `drift-versions.sh`'e bir kaynak eklenmeli.
- Dokümantasyon kayması ölçülmüyor. Bir ajan hook sayfasını değiştirirse
  sürüm değişmeden kayma olabilir; bunu yakalayan tek şey yeniden yakalama.

---

## 2026-10-05 — v0.1.0 taslak release, ve CI'ın hangi Go'yu kullandığı

Dokuz faz push edildi (`a8cd53f..0d1f0da`), CI yeşil. `v0.1.0` etiketlendi,
release workflow altı ikiliyi ve `SHA256SUMS`'u **taslak** release'e koydu.
Yayımlamak `[SEN]`.

### CI ile yerel derleme aynı baytları üretmiyor

Taslağı kontrol ederken çıktı: CI'ın ürettiği ikililer yereldekilerle
eşleşmiyor.

```
CI    go1.26.8 linux   darwin_arm64 -> 860b20b7...
yerel go1.27.1 darwin  darwin_arm64 -> 754facb9...
```

Sebep: `go-version-file: go.mod` + `check-latest: true`, ve `go.mod` `go 1.26`
diyor. Yani bu "**1.26'nın en son yaması**" demek, Go 1.27 değil. Üç sonuç:

1. Kullanıcıların indirdiği ikililer, desteklediğimizi iddia ettiğimiz asgari
   Go ile derleniyor. Kasıtlı değildi ama **doğru olan bu** — yoksa
   "Go 1.26 veya üstü" iddiası hiçbir yerde sınanmıyor.
2. `release.sh`'i daha yeni bir toolchain'le koşan biri eşleşmeyen baytlar
   üretir ve tekrarlanabilirlik sanıp kafası karışır.
3. En yeni Go yalnızca benim makinemde 1.27 olduğu için **kazara** kapsanıyordu.

Karar: asgari sürümle derlemeye devam, ve `docs/versioning.md`'ye "Which Go
builds a release" bölümü. Ayrıca `ci.yml`'ye ikinci bir iş eklendi
(`latest-go`, `go-version: stable`) — en yeni Go artık kaza eseri değil,
gerçekten test ediliyor. Matris yerine iki iş: setup-go ikisi için farklı
girdi alıyor, iki kez okunması gereken bir koşuldansa açık açık yazmak daha
iyi.

### v0.1.0 yayımlandı

https://github.com/forgeprint/forgelore/releases/tag/v0.1.0 — altı ikili ve
`SHA256SUMS`, CI'ın temiz runner'ında derlenmiş.

Yayımlarken iki karar:

- **Notlar yeniden yazıldı.** `--generate-notes` önceki etiket olmadığı için
  tek satırlık bir bağlantı üretmişti. İlk release'in neyin çalıştığını ve
  **neyin doğrulanmadığını** söylemesi gerekiyordu; Codex CLI'nin eşlemesinin
  dokümantasyondan yazıldığı ve aynı şekilde yazılan diğer her eşlemenin
  yanlış çıktığı notlarda açıkça yazılı.
- **Prerelease işaretlenmedi.** 0.x zaten erken olduğunu söylüyor, ama asıl
  sebep pratik: GitHub `/releases/latest` uç noktasından prerelease'leri
  dışlıyor ve `install.sh` tam olarak onu kullanıyor. İşaretlenseydi kurulum
  betiği release'i bulamazdı.

#### Faz 8 kabul kriteri artık gerçek

Gerçek release'e karşı, tek komutla, boş `HOME` ve kırpılmış `PATH` ile
koşuldu: indirdi, checksum'ı doğruladı, kurdu, `forgelore v0.1.0
darwin/arm64 go1.26.8` bastı, PATH uyarısını verdi, `init` ve `doctor`
çalıştı. Daha önce yalnızca yerel bir HTTP sunucusuyla test edilmişti.

`go1.26.8` damgası da yukarıdaki kararın kanıtı: kullanıcının indirdiği ikili
asgari Go ile derlenmiş.

### Kalanlar — hepsi `[SEN]`

- Forgeprint marketplace kaydı (ayrı depo).
- İki kişilik bir haftalık ekip denemesi (`docs/team-trial.md`).
- Codex CLI doğrulaması, erişim olduğunda:
  `./scripts/capture-agent-events.sh codex-cli`, sonra eşlemeyi düzelt,
  `verified_against`'i doldur, `docs/compatibility.md`'yi güncelle.
- Sonraki release'te imzalama (cosign keyless, artık CI'da derlendiği için
  mümkün) ve istenirse npm sarmalayıcı — ikisi de Faz 8'de bilerek ertelendi.

---

## 2026-10-05 — README demosu yazarken bulunan hata: `go vet` tanınmıyormuş

README'ye gerçek bir transcript koymak için demoyu koştururken çıktı:
**Forgelore `go vet` çıktısını hiç tanımıyordu.**

```
vet: ./main.go:3:2: undefined: greet
```

Konum eşleştiricisi satırın **ilk jetonuna** çapalı (bu, Node yığın
karelerini dışarıda tutan şeydi) ve ilk jeton `vet:` olunca nokta-uzantı
taşımadığı için hiç eşleşmiyordu. Hata `v0.1.0`'da var.

Acısı şurada: Faz 2'de fiili parmak izinden çıkarmanın **gerekçesi** "aynı
derleme hatası `go build`, `go test`, `go run` ve `go vet` altında çıkar"
idi. Fiili çıkardık, ama `go vet`'in çıktı biçimi yüzünden o yola hiç
varamıyorduk. Karar doğruydu, uygulaması yarımdı.

### Düzeltme

- `scripts/capture-errors.sh`'e `go/vet-undefined` ailesi eklendi, iki
  varyantla. Külliyat 58 dosya / 29 aile.
- `fileLineDiag` artık konumun önünde isteğe bağlı bir `kelime: ` öneki
  kabul ediyor. Çapa korunuyor, yani Node kareleri hâlâ dışarıda.

### Test düzeltmeyi yasaklıyordu

Düzeltmeden sonra `go/undefined-identifier` ile `go/vet-undefined` **aynı
parmak izini** üretti — `cd023fb609411574` — ve `TestFamiliesSeparate` bunu
çakışma sayıp kırıldı. Oysa istenen tam olarak buydu: iki aile, bir hata, iki
komut.

Test yeniden yazıldı. `sameError` listesi "bunlar bilerek aynı" diyor,
ayrışma kontrolü onları muaf tutuyor, ve yeni `TestOneErrorThroughTwoCommands`
**paylaştıklarını zorunlu kılıyor** — kesişim alarak, deterministik olarak.
Yani test artık özelliği yasaklamak yerine kanıtlıyor.

### v0.1.1 çıkarıldı

Karar: çıkarılsın. `go vet` kullanan biri hiçbir şey görmez ve sebebini
anlayamaz; bir sonraki sürümü beklemek, sessiz kalan bir aracı sahada
bırakmak olurdu.

https://github.com/forgeprint/forgelore/releases/tag/v0.1.1

Gerçek release'ten kurulup doğrulandı: `go vet` çıktısıyla kaydedilen
düzeltme, `go build` altında ve başka bir dosyada bulundu. Kurulan ikili
`v0.1.1 darwin/arm64 go1.26.8` — asgari Go kararı bu sürümde de geçerli.

Notlarda "nothing to migrate" yazılı ve doğru: `v0.1.0`'ın yazdığı kayıtlar
olduğu gibi okunuyor, o dönemde kaydedilmiş parmak izleri hâlâ geçerli.
`v0.1.0` yalnızca `go vet`'ten hiç parmak izi üretmiyordu.

README'ye gerçek transcript eklendi (demo GIF'in birinci alternatifi) ve
durum paragrafı düzeltildi: "yayımlanmış release yok" diyordu, oysa v0.1.0
yayımlanmıştı.

### Yeniden üretme külliyatı bozuyordu — iki şey daha çıktı

`./scripts/capture-errors.sh go` yalnızca yeni aileyi eklemedi, **mevcut
sekiz aileyi de yeniden üretti** ve macOS'ta koştuğu için Windows yol
biçimlerini sildi: `.\main.go` → `./main.go`, `C:/Users/...` →
`/var/folders/...`. Oysa külliyatın README'si tam olarak "bir Windows ev
dizini Windows ev dizini olarak kalır, çünkü normalleştirmenin başa çıkması
gereken şey budur" diyor. Eski dosyalar geri alındı, yalnızca
`go/vet-undefined` tutuldu.

İkincisi: frontmatter'daki `platform` alanı betikte **sabit** yazılmıştı
(`windows/amd64`). macOS'ta üretilen dosyalar darwin yolları taşırken
Windows olduklarını iddia ediyordu. Artık `go env GOOS/GOARCH`'tan türüyor.

Sonuç beklenmedik biçimde daha iyi: paylaşan iki aile artık farklı
işletim sistemlerinde yakalanmış durumda, yani
`TestOneErrorThroughTwoCommands` parmak izinin işletim sistemi, yol biçimi,
satır numarası ve alt komuttan **aynı anda** bağımsız olduğunu kanıtlıyor.

---

## 2026-10-05 — README demosu: transcript evet, GIF hayır

Soru "kurulum rozetinin altına kısa bir demo GIF ekleyelim mi" idi. İki
alternatif önerildi ve sırayla denendi.

### 1. Gerçek transcript — yapıldı

README'ye `## What it looks like` bölümü eklendi: `go build` patlıyor,
`recall` bilmiyor, kayıt yapılıyor, `go vet` aynı hatayı **farklı biçimde**
raporluyor, `recall` buluyor. Parmak izi ikisinde de `cd023fb609411574`.

Her bayt gerçek bir koşudan, yayımlanmış `v0.1.1` ikilisiyle üretildi.
Bağımlılık yok, `cat README.md` ile okunuyor, ekran okuyucuda çalışıyor,
kopyalanabiliyor, diff'leniyor.

### 2. VHS ile GIF — yapılmadı

Varsayım "VHS kurulur, `.tape` commit edilir, yeniden üretilebilir olur"
idi. Maliyet ölçülünce başka çıktı:

| Engel | Bulgu |
|---|---|
| `ttyd` (VHS'in zorunlu bağımlılığı) | macOS ikilisi **yayımlamıyor** — son sürümün 12 varlığının hepsi Linux. Kaynaktan derleme CMake + libwebsockets + json-c + openssl ister, brew de yok |
| Docker yolu | Resmî VHS imajı 10 dakikada **sıfır bayt** indirdi, imaj gelmedi |
| İmaj gelseydi | İçinde Go yok. Demonun can alıcı kısmı `go build`'in gerçekten patlaması; ya committed bir Dockerfile ile Go eklenecek ya da derleyici hiç gösterilmeyecekti |

Karar: **yapılmadı.** Üçünün toplamı bir README GIF'i için fazla, ve asıl
mesele şu: o GIF bu depodaki **yeniden üretilemeyen tek artefakt** olurdu.
Hata külliyatının betiği var, ajan yüklerinin betiği var, release'in betiği
var; Dockerfile'lı kurguda bile bu makinede regenerate edilemediği sürece
çizgi bozulurdu.

İstenirse iki makul yol var, ikisi de bu makinenin dışında: Linux'ta
üretmek (VHS'in bağımlılıkları paket yöneticisinde var), ya da ttyd'nin
macOS ikilisi yayımlamasını beklemek. `.tape` dosyası her hâlükârda commit
edilmeli.

### Rozetler

İki rozet eklendi: CI (**GitHub'ın kendi uç noktası**, üçüncü taraf yok) ve
release sürümü (shields.io — bunun için pratik tek seçenek, ama README'yi
açan herkesin tarayıcısı üçüncü tarafa istek yapıyor; bilerek kabul edildi).

---

## 2026-10-05 — ekip denemesi protokolü gözden geçirildi

Faz 6'da yazılmıştı; o günden beri v0.1.1 yayımlandı ve Faz 7 geldi. Üç
**gerçek hata** bulundu, üçü de denemeyi sessizce bozardı:

1. **"`init --with-git-hook` bir kişi yapar, sonra commit edilir" yanlıştı.**
   Hook `.git/hooks/` altına yazılıyor ve git orayı hiç commit etmiyor.
   İkinci kişi sırrı durduran kontrolden tamamen yoksun kalırdı. Protokol
   artık herkesin kendi makinesinde koşmasını söylüyor ve sebebini yazıyor.

2. **MCP ölçüme hiç girmiyor.** `internal/mcp` deftere tek satır yazmıyor —
   `grep -c "measure\." internal/mcp/*.go` → 0. Protokol MCP'yi hook'a
   denk bir seçenek gibi sunuyordu; o yolu seçen kişi bir hafta çalışıp
   **boş bir rapor** alırdı ve sebebini anlayamazdı.

3. **Elle `forgelore recall` de ölçülmüyor.** `--session` varsayılanı boş ve
   `loopsPerSession` boş oturumları atlıyor (Faz 3'te bilerek böyle
   yapılmıştı, alakasız çağrılar döngü uydurmasın diye). Doğru davranış, ama
   protokolde yazılı olması gerekiyordu.

Artık hangi yolun ölçüldüğünü gösteren bir tablo var.

### Eklenen: "gün sıfır" duman testi

Protokolün en büyük eksiğiydi. Hafta sonunda "kimse bir şey promote etmedi"
bulgusu ile "hook'lar hiç tetiklenmedi" ayırt edilemiyordu. Artık gerçek işe
başlamadan önce kasıtlı bir hata kırdırıp `report --days 1` ile hook'un
gördüğü doğrulanıyor, `doctor` ile ajanın doğrulanmış olup olmadığına
bakılıyor.

Bu sezgi bu oturumdan geliyor: Copilot'un `.github/hooks`'u geçici dizinde
okumaması ve Claude Code'un eklentiyi etkinleştirme gerektirmesi, ikisi de
"çalışıyor sandım, çalışmıyormuş" vakasıydı.

### Diğer düzeltmeler

- Kurulum satırı güncellendi (release artık var).
- Copilot CLI eklendi (Faz 7'de doğrulandı).
- Günlük not defterinin depo **dışında** tutulması söylendi.
- Günde en az bir kez pull — iki kişinin ağacı buluşmazsa merge iddiası
  sınanmamış olur.
- Maliyet karşılaştırmasının status line bağlanmadan çıkmayacağı yazıldı.
- A/B oranı 20'de bırakıldı ama **gerekçesi yazıldı**: bir hafta zaten
  anlamlılık üretemez, büyük kontrol kolu kullanışlılıktan götürür ve
  istatistik getirmez.

### MCP ölçüm boşluğu kapatıldı

`recall_error` isteğe bağlı bir `session` argümanı aldı ve artık deftere
yazıyor: inject / control / miss, bayt, tahmini token, A/B kolu.

Durumsuzluk ihlali değil — spesifikasyonun kendi "Stateful Tools" bölümü
tam bunu söylüyor: birden fazla isteğe yayılan durum, istemcinin her çağrıda
geçirdiği açık bir tanımlayıcıyla taşınmalı. Sunucu hiçbir şey çıkarsamıyor.

Kontrol kolu burada da görünmez: eşleşme varsa bile `hits` boşaltılıyor ve
çağrı bir miss'ten ayırt edilemiyor; defter eşleşmenin var olduğunu yine de
kaydediyor. Boş liste olarak, `nil` olarak değil — JSON her hâlükârda dizi
kalsın diye.

Oturum kimliği verilmezse arama yine cevap veriyor, yalnızca harcama
tarafına sayılıp oturum başına karşılaştırmalardan düşüyor. Araç açıklaması
bunu açıkça söylüyor.

Gerçek istemciyle doğrulandı: MCP inspector'dan `session` geçirilerek
yapılan bir `recall_error`, `report --days 1` çıktısında "hints injected 1,
bytes injected 42" olarak göründü.

Protokolün tablosu güncellendi: MCP artık "istemci `session` geçirirse
ölçülür". Modelin bunu güvenilir biçimde geçirip geçirmediği denemenin
kendi bulgularından biri olacak — gün sıfırda bakılmalı, cuma günü değil.

---

## 2026-10-05 — v0.1.2

https://github.com/forgeprint/forgelore/releases/tag/v0.1.2

MCP üzerinden yapılan aramalar artık ölçülüyor (`recall_error`'ın `session`
argümanı) ve `docs/team-trial.md`'nin üç düzeltmesi de içeride.

Yayımlanan ikiliyle doğrulandı: temiz bir `HOME`'a kuruldu, MCP inspector'dan
`session` geçirilerek bir `recall_error` yapıldı, ve `report --days 1` onu
"hints injected 1, bytes injected 31" olarak gösterdi — yani ölçüm yolu
gerçek release'te uçtan uca çalışıyor.

### Üç sürümün ritmi

`v0.1.0` → `v0.1.1` → `v0.1.2` aynı günde çıktı, ve ikisi de bir **dokümanı
yazarken** bulunan şeyler yüzünden:

- `v0.1.1`: README demosu yazılırken `go vet`'in hiç tanınmadığı görüldü.
- `v0.1.2`: ekip denemesi protokolü gözden geçirilirken MCP'nin hiçbir şey
  ölçmediği görüldü.

İkisi de sessiz hatalardı — hiçbir test kırılmıyordu, hiçbir kullanıcı hata
mesajı görmeyecekti. Ortak nokta: **aracı anlatmaya çalışmak, onu
kullanmaktan daha iyi bir kontrol yöntemi çıktı.** Testler yazdığımız şeyin
çalıştığını kanıtlıyor; bir doküman, yazmadığımız şeyin eksik olduğunu
gösteriyor.

---

## 2026-10-05 — compatibility.md gözden geçirildi

Dört düzeltme, biri kopuk bağlantı.

**1. Copilot CLI "A" olarak fazla iddia ediyordu.** A'nın tanımı "hook + MCP
+ CLI". Copilot'un hook'ları doğrulandı ama **MCP istemcisi hiç
bağlanmadı** — yani A hak edilmemişti. **B**'ye indirildi, gerekçesi
tabloda yazılı. Tam olarak bu tablonun engellemek için var olduğu hata.

**2. `docs/team-trial.md` buraya yönlendiriyordu ama kurulum talimatı
yoktu.** Protokol "Copilot CLI: hooks per `docs/compatibility.md`" diyor,
burada ise tek satır yazmıyordu. Artık tam hooks dosyası var — olay başına
bir girdi, her biri kendi adını `--event` ile geçiriyor.

`COPILOT_HOME` tercihinin sebebi de dürüstçe yazıldı: `.github/hooks`
yakalama betiğinin geçici dizininde okunmadı, ama o dizin git deposu değil,
yani sebep bu da olabilir. "Genel kural" diye yazmak elimizdeki kanıtı
aşardı.

**3. Codex'in ikinci tahmini de işaretlendi.** Eşleme `interrupted` /
`is_interrupt` alanını atlıyor — bu Claude Code'dan kopyalandı. Oysa Codex
`Interrupt`'ı bir **olay** olarak dokümante ediyor, alan olarak değil. Yani
atlama hiç tetiklenmeyebilir ve iptal edilen komut hata olarak
hatırlanabilir. Üç oturum açma denemesinin başarısız olduğu da yazıldı.

**4. "MCP sunucusu ajana özgü değil" bölümü eklendi.** Tablodaki "never
connected" o ajanın istemcisi hakkında bir ifade, Forgelore hakkında değil
— sunucu iki revizyonu da sunuyor ve iki gerçek istemciyle sınandı.
v0.1.2'nin `session` argümanı da burada: MCP üzerinden yapılan arama ancak
çağıran onu geçerse rapora giriyor.

Ayrıca "Keeping this table true" bölümü eklendi — iki drift betiği ve
`doctor`, tabloyu dürüst tutmanın yolu olarak. "Adding an agent" listesine
de "kontrol **etmediğin** sütunları da güncelle" maddesi kondu.

---

## 2026-10-05 — v0.1.3 çıkarılmadı, v0.1.2'nin notları güncellendi

`v0.1.3` istendi. `git diff --stat v0.1.2..HEAD` yalnızca iki doküman
gösterdi, tek satır Go kodu yok. Yeni bir etiket, işlevsel olarak aynı
ikilileri farklı bir sürüm dizesiyle yayımlamak olurdu — patch sürüm
yazılımdaki bir düzeltme için, ve burada yazılımda düzeltilen bir şey yoktu.
Dokümanlar zaten `main`'den canlı.

Bunun yerine **`v0.1.2`'nin notlarına** düzeltilmiş uyumluluk bilgisi
eklendi: hangi ajanın neye karşı sınandığı, Copilot'un A değil B olduğu ve
neden, ve Codex'in iki spesifik tahmininin neden güvenilmemesi gerektiği.
Notun kendisi "bu yayımlandıktan sonra eklendi, hiçbir ikili değişmedi"
diye başlıyor.

Varlıklar ve checksum'lar değişmedi (7 dosya yerinde); yalnızca metin
düzenlendi.

### Kural olarak

Doküman düzeltmesi sürüm çıkarmaz. Okuyucunun release sayfasında görmesi
gereken bir şeyse, o sürümün notu güncellenir. Bir sonraki kod
değişikliğinde doküman değişiklikleri zaten onunla gider.

---

## 2026-10-05 — marketplace kaydı: yayında

Plan bunu `[SEN]` ve "ayrı depoda" işaretlemişti; kullanıcı buradan
yapılmasını istedi.

Marketplace zaten `forgeprint/forgeprint/.claude-plugin/marketplace.json`'da
duruyordu, dört eklentiyle. Forgelore beşinci girdi olarak eklendi:

**PR: https://github.com/forgeprint/forgeprint/pull/207** — altı kontrol de
yeşildi (validate, DCO, lint-setup, render-check, similarity, setup-test).
Kullanıcı 12:37 UTC'de birleştirdi; `forgeprint/forgeprint` main'de `5df475d`.
Girdi canlı, ve `forgelore.git` içindeki `plugin/` yolu çözülüyor.

### Kaynak tipi ve açıklama kasıtlı

`git-subdir`, `github` değil: eklenti **başka** bir deponun alt dizininde
(`plugin/`), ve spesifikasyon tam bu durum için `git-subdir` diyor. `ref`
sabitlenmedi — `plugin/` bir manifest ve bir hooks dosyasından ibaret,
nadiren değişiyor; etikete bağlamak her Forgelore release'inde katalogu
güncellemeyi gerektirirdi.

Açıklama **"Requires the forgelore binary on your PATH"** ile bitiyor. `bin/`
dizinini bilerek koymamıştık (o dizin varsa claude.ai eklentiyi kurmuyor), ve
bunun sonucu şu: kullanıcı eklentiyi kurar, binary yoksa hook'lar sessizce
hiçbir şey yapmaz. Bugünün tekrar eden hatası bu aileden; cümle o yüzden
orada.

### Eklenti ilk kez doğrulandı

`claude plugin validate ./plugin` — geçti, `--strict` dahil. Dokümantasyon
bunu yetkili kontrol diye tanımlıyor ama Faz 4'te manifesti yazarken hiç
çalıştırmamışız. Birleşmiş `marketplace.json` da aynı doğrulayıcıdan
geçirildi.

### Üç şey öğrenildi, ikisi benim hatamdı

1. **Prettier.** O depo `marketplace.json`'ı Prettier ile biçimlendiriyor;
   `json.dump(indent=2)` çıktısı uymadı ve `validate` kırıldı. Kendi
   `.prettierrc.json`'larıyla (printWidth 100) yeniden biçimlendirildi —
   `tags` tek satıra indi.
2. **DCO imzası yazarla eşleşmeli.** Forgelore'un biçimini
   (`aliosman.mho@gmail.com`) kullandım, oysa o deponun commit'leri
   `aliosmanmho@users.noreply.github.com` ile imzalanıyor. İki deponun iki
   ayrı konvansiyonu var.
3. **Dalı base'e sıfırlamak PR'ı kapattırdı.** Düzeltilmiş commit'i atmak
   için önce dalı main'e force ettim; o anda fark sıfır olunca GitHub PR'ı
   otomatik kapattı. Geri açıldı, dal tek temiz commit taşıyor. Doğrusu tek
   adımda yeni commit'e force etmekti.

### Neden doğrudan main'e değil PR

`branches/main/protection` 404 veriyor ("Branch not protected"), yani klasik
API'ye göre koruma yok. Ama son beş commit'in hepsi bir PR numarası
taşıyordu, ve ruleset API'si bakılınca `protect-main` gerçekten `pull_request`
ve `required_status_checks` dayatıyor. Klasik koruma uç noktasının 404'üne
bakıp "doğrudan yazılabilir" demek yanlış olurdu.

### Birleştikten sonra: kurup denedik, ve `ref` notu yanlıştı

Marketplace'ten kurulum denendi: `claude plugin marketplace add
forgeprint/forgeprint`, sonra `claude plugin install forgelore@forgeprint`.
İkisi de sorunsuz. `git-subdir` doğru çözüldü — önbelleğe deponun tamamı
değil yalnızca `plugin/` içeriği indi.

Yukarıda "`ref` koymadığımız için her commit kuranları anında etkiler"
yazmıştım. **Yanlış.** `~/.claude/plugins/installed_plugins.json` kurulumu
commit sha'sına sabitliyor ve `installPath`'i sürüme göre adlandırıyor:

```
"installPath": ".../cache/forgeprint/forgelore/0.1.0",
"version": "0.1.0",
"gitCommitSha": "08658f5da309f0790244cdc9c05fe47b15e790c7"
```

Yani `ref` koymamanın etkisi **yeni** kurulumlarda: onlar main'in o anki
hâlini alıyor. Mevcut kurulumlar güncellenene kadar sabit kalıyor.

Doğrulanmayan kısmı olduğu gibi bırakıyorum: `plugin.json` içindeki sürüm
`0.1.0` dururken sha ilerlerse `claude plugin update` ne yapıyor? Dizin adı
aynı sürüme düşüyor, ve bunu denemeden söylemeyeceğim. Pratik sonuç: `plugin/`
altında bir şey değişirse `plugin.json` sürümünü de artır.

### Hook'un enjekte ettiği ölçüldü

Scratchpad'de kırık bir Go projesi (`undefined: greet`), bir kayıt, sonra
gerçek yakalanmış `PostToolUseFailure` payload'ı o dizine yönlendirilip
çalıştırıldı:

```
{"hookSpecificOutput":{"additionalContext":"forgelore has seen this error
before:\n- fix: greet lives in ... (01M461QEF80ZCNA9ADQ0PV57P8)",
"hookEventName":"PostToolUseFailure"}}
```

Bilinmeyen bir hatada çıktı boş. `report --days 1`: 3 arama, 2 enjeksiyon,
1 "bilinmiyor", hook gecikmesi p95 9 ms.

Canlı oturum içinde denenemedi: `claude -p` "OAuth session expired" verdi.
Eklenti kurulu ve binary PATH'te olduğunda gerçek bir oturumda `go build`
kırmak kalan tek adım — `[SEN]`, çünkü CLI'ya giriş kullanıcıya ait.

### Canlı oturumda denendi: çalışıyor, ve bir kör nokta çıktı

CLI'ya giriş yapıldı, ikili `~/.local/bin`'e kuruldu, ve marketplace'ten
kurulu eklenti gerçek oturumlarda çalıştırıldı. Beş oturumda kayıt enjekte
edildi, hepsi gerçek Claude Code oturum kimlikleriyle ledger'da:
`report --days 1` 8 arama, 7 enjeksiyon, 6 oturum, hook p95 10 ms.

İlk iki oturum hiçbir şey üretmedi ve bunu "hook çalışmıyor" diye okudum.
Yanlıştı. Yakalanan payload'da model komutu şöyle yazmıştı:

```
go build ./... 2>&1 | head -40
```

Pipeline'ın çıkış kodu `head`'inki, yani sıfır. Komut gerçekten başarıyla
bitmiş; Claude Code doğru şekilde `PostToolUse` göndermiş, Forgelore da doğru
şekilde "başarı" deyip hiçbir şey yazmamış. Düz `go build ./...` ile
`PostToolUseFailure` geliyor ve enjeksiyon oluyor — `default`, `acceptEdits`
ve `bypassPermissions` ile ayrı ayrı doğrulandı.

**Kör nokta bu:** çıkış kodu payload'ın hiçbir yerinde yok. Başarısızlık
yalnızca hangi olayın tetiklendiğinden anlaşılıyor. Ajan build'i `head`,
`tail` veya `tee`'ye borularsa hata Forgelore'a hiç ulaşmaz, ve bu sessizce
olur — ajanın hiç hata görmemesinden ayırt edilemez. Bugünün ailesinden bir
hata daha: doğru çalışan kod, hiçbir testin yakalayamayacağı bir boşluk.

Düzeltmek mümkün ama bedava değil: başarısızlığı olaydan değil çıktı
metninden okumak gerekir, yani Copilot CLI eşlemesindeki
`failure.output_matches`. Bu, bu ajan için "başarısız" tanımını değiştirmek
demek; tek bir gözleme dayanarak yapılmadı. `docs/compatibility.md`'ye de
yazıldı.

### Yöntem notu

Teşhis, tahminle değil, aynı komutu üç izin modunda çalıştırıp payload'ları
yan yana koyarak çıktı. Çıktıyı dosyaya döken geçici bir `settings.json`
hook'u, eklentininkinin yanında çalıştı ve ikisi birbirine karışmadı —
sonraki bir ajan için en ucuz teşhis aracı bu.

---

## 2026-10-05 — borulanmış komutun gizlediği hata kapatıldı (ADR-0022)

Karar kullanıcınındı: `output_matches`'a geç. Ama Claude Code'un
`PostToolUse` payload'ında Copilot'taki `completed with exit code [1-9]`
gibi bir işaret yok — çıkış kodu hiçbir alanda geçmiyor. Geriye tek soru
kalıyor: çıktı hataya **benziyor** mu?

İki cevap vardı ve ikisi aynı değildi:

1. **Mapping'e regex.** Go'ya dokunmaz, format 1'de kalır. Ama
   fingerprinter'ın desenini JSON'da ikinci kez yazmak demek, ve
   dosya:satır deseni `testdata/errors`'taki 29 ailenin ancak yarısına
   yetişir. Panic'ler, `ModuleNotFoundError`, `npm ERR!` yine sessiz kalır.
   Tam görünüp yarısını kapatan bir düzeltme, hiç düzeltmemekten kötü —
   kimse geri dönmüyor.
2. **Fingerprinter'a sor.** Bilgi tek yerde, 29 ailenin hepsi kapsanıyor.

İkincisi seçildi: `failure.output_has_diagnostic`.

### Sürüm numarası neden 2'ye çıktı

`docs/versioning.md` "ekleme ise 1'de kalır, mevcut bir alan anlam
değiştirirse 2 olur" diyordu. Bu kural bu vakada yanlış cevap veriyor ve
düzeltildi. Doğru soru şu: **eski bir ikili bu dosyayı okuyup yine de haklı
olabilir mi?** Yanıt şekli ve çağıran-kaynaklı olay adı için evet. Bunun
için hayır — eski ikili alanı tanımaz, düşürür, ve her borulanmış
başarısızlığa "başarı" der. Sessizce. O yüzden dosya 2 diyor ve o ikili onu
yüksek sesle reddediyor. Sürüm 1 iddia edip alanı kullanan bir mapping de
reddediliyor; yoksa numara garanti değil etiket olurdu.

`copilot-cli.json` ve `codex-cli.json` 1'de kaldı — ihtiyaçları yok, ve eski
bir Forgelore onları hâlâ okuyabiliyor.

### Bedeli açıkça yazıldı

Format artık saf bildirimsel değil: bir failure testini Go cevaplıyor.
ADR-0020'nin "ajan eklemek Go gerektirmez" sözünün artık bir istisnası var.
Ayrıca başarılı ama hata biçimli çıktı veren komut (`cat build.log`) artık
başarısız sayılıyor. Zarar sınırlı — `onCommandFailed`, `Scan` bir şey
bulamazsa hiçbir şey yapmadan dönüyor — ve üç yerde görünür: `report`,
enjekte edilen metin, `review`. Hiçbiri kayıt yazmıyor.

### Doğrulama

Yeni ikili kuruldu, canlı Claude Code oturumunda `go build ./... 2>&1 | head
-40` çalıştırıldı: ledger 8 → 9, enjeksiyon oldu. Korpusa gerçek payload
`PostToolUse-2.json` olarak eklendi; `TestAPipedBuildStillCountsAsAFailure`
onu ve başarılı `go version`'ı yan yana tutuyor, çünkü davranışın tamamı
ikisinin farkında.

Dokunulan dosyalar: `internal/agent/mapping.go`,
`internal/agent/mappings/claude-code.json`,
`internal/agent/contract_test.go`,
`testdata/agents/claude-code/2.1.289/PostToolUse-2.json`,
`docs/adr/0022-a-failure-test-that-is-not-data.md`, `docs/versioning.md`,
`docs/mappings.md`, `docs/compatibility.md`.

---

## 2026-10-05 — v0.1.3

https://github.com/forgeprint/forgelore/releases/tag/v0.1.3

Borulanmış komutun gizlediği hata (ADR-0022) ve mapping formatının 2'ye
çıkışı. Tag push → CI taslağı kurdu → notlar yazılıp yayımlandı, `latest`
işaretlendi.

Yayımlanan ikiliyle doğrulandı — v0.1.0 dersinin gereği: indirildi, checksum
tuttu, `v0.1.3 darwin/arm64 go1.26.8`. Borulanmış build payload'ı enjeksiyon
üretti, başarılı `go version` payload'ı sessiz kaldı. Yani düzeltme ağaçta
değil, **release'te** var.

Notların söylemesi gereken şeyler `docs/versioning.md`'nin listesinden
geldi; bu sefer kritik olan madde "mapping formatı taşındı mı, hangi
ajanları etkiliyor" idi. Not açıkça `--mapping` ile kendi dosyasını ezen
kullanıcıya da sesleniyor: hiçbir şey kırılmıyor, yeni testi istiyorsa
`"mapping_version": 2` yazacak.

### Dördüncü sürümün ritmi

`v0.1.1` ve `v0.1.2` bir **doküman yazarken** bulunan hatalardan doğmuştu.
`v0.1.3` farklı bir yerden geldi: **kendi aracımızı bir kullanıcı gibi
kurup çalıştırmaktan.** Marketplace'ten kurulmasaydı ve gerçek bir oturumda
denenmeseydi borulanmış komut sonsuza kadar sessiz kalırdı — hiçbir test
kırılmıyordu, korpustaki dört payload da geçiyordu.

Dört sürümün dördü de aynı aileden: kod doğru çalışıyordu, eksik olan şey
kodun hiç görmediği bir girdiydi. Sırayla anlatmak, kullanmak, kurmak.

---

## 2026-10-05 — imzalama ve npm (ADR-0023, ADR-0024)

Faz 8'in bilerek ertelediği iki madde. İkisi de araştırıldı, iki karar
soruldu, ikisi de önerilenle gitti.

### İmzalama: attestation, cosign değil

`SHA256SUMS` "bu baytlar mı yayımlandı" sorusunu cevaplıyor ama "kim
yayımladı"yı cevaplayamıyor — çünkü aynı release'in içinde duruyor; birini
değiştirebilen ikisini de değiştirir. `install.sh` de checksum'ları aynı
release'ten okuyor.

`actions/attest-build-provenance` (SHA'ya sabitlendi, her action gibi),
`subject-path: dist/*`, ve **release oluşturulmadan önce** çalışıyor: adım
patlarsa düzeltilecek bir release kalmıyor.

cosign elenmedi, tartıldı: GitHub'dan bağımsız doğrulama ve Rekor kaydı
veriyordu, ama bir installer adımı, iki yeni varlık ve kullanıcıda cosign
şartı getiriyordu. Attestation tek adım ve sıfır varlık.

Bedeli ADR-0023'te: kanıt GitHub'da duruyor (çevrimdışı doğrulanamıyor),
kolay yol `gh` istiyor, ve release job'ı artık `id-token: write` taşıyor —
job'ın başka hiçbir şey yapmamasının sebebi bu. Bir de bu **işletim sistemi
imzalaması değil**: Gatekeeper ve SmartScreen Sigstore'u tanımıyor.

### npm: platform başına bir paket

Plan "postinstall indirip checksum doğrulasın" diyordu. Yapılan o değil, ve
sebebi yazıldı: `npm install --ignore-scripts` artık yaygın, ve o bayrakla
paket "başarıyla" kuruluyor ama içinde binary olmuyor. Sessiz başarısızlık —
bu projenin bütün gününü aldığı tür.

Onun yerine esbuild'in yöntemi: altı platform paketi, `os`/`cpu` alanlarıyla,
ve hepsine `optionalDependencies` ile bağlanan tek bir `forgelore`
sarmalayıcısı. Kurulumda script yok, indirme yok, registry dışında ağ yok.

Sarmalayıcı tek dosya. `stdio: "inherit"` — araya girmemek bilinçli, çünkü
`forgelore mcp` stdin/stdout üzerinde protokol konuşuyor ve araya giren her
şey framing konusunda sonsuza kadar haklı kalmak zorunda olurdu.

`scripts/npm-pack.sh` paketleri `dist/`'ten kuruyor ve `release.sh` gibi
yayımlamıyor, komutları yazdırıyor. Platform paketleri **önce** yayımlanmalı:
sarmalayıcı bağımlılıklarını tam sürümle sabitliyor, registry'de yoksa
binary'siz kuruluyor.

### Yerelde uçtan uca denendi

`v0.1.3` artefaktlarıyla yedi paket üretildi ve boş bir dizine tarball'dan
kuruldu:

- `--ignore-scripts` ile kuruldu ve `forgelore v0.1.3 darwin/arm64` dedi;
- çalıştırma bitinin tarball'da korunduğu doğrulandı;
- çıkış kodu sarmalayıcıdan birebir geçti (ikili 1 → sarmalayıcı 1);
- MCP isteği sarmalayıcıdan ve ikiliden **aynı** baytları döndürdü;
- `--no-optional` ile kurulduğunda modül çözümleme hatası değil, ne olduğunu
  anlatan bir cümle çıktı.

### İki şey açık

- **Attestation v0.1.3'te yok.** Bir sonraki etiketten itibaren geçerli;
  yayımlanmış v0.1.3 varlıkları attestation taşımıyor.
- **npm paketleri yayımlanmadı.** `dist/npm/` hazır ama `npm publish` bir
  hesap ve giriş istiyor — `[SEN]`. CI'dan yayımlamak bir automation token
  istiyor, ve o yapılırsa npm provenance de gelir; şu hâliyle release
  attested, npm paketleri değil.

---

## 2026-10-05 — v0.1.4

https://github.com/forgeprint/forgelore/releases/tag/v0.1.4

Programda değişiklik yok; bu release nereden geldiğini kanıtlamakla ve
ikinci bir kurulum yolu eklemekle ilgili. Notlar da bu cümleyle açılıyor.

### Attestation gerçek bir release'te doğrulandı

İndirilen `forgelore_darwin_arm64` için `gh attestation verify` 0 ile
çıktı, ve `--format json` kanıtın ne iddia ettiğini gösterdi:

```
workflow : .../.github/workflows/release.yml@refs/tags/v0.1.4
repo     : https://github.com/forgeprint/forgelore
sha      : 95c60b49cc051f0ecda0e88309a0e2370066ff4d
issuer   : https://token.actions.githubusercontent.com
```

Olumlu sonuç tek başına yeterli değildi, negatif de denendi: **yerelde
derlenmiş** aynı kaynaklı ikili 404 ile reddedildi — o digest için
attestation yok. Doğrulamanın gerçekten bir şey kontrol ettiğini gösteren
şey bu, "başarılı" demesi değil.

Bu arada bir kullanılabilirlik ayrıntısı çıktı ve dokümana girdi: gh 2.102
başarıda **hiçbir şey yazmıyor**. Komutu çalıştıran biri hiçbir şey olmadı
sanabilir; cevap çıkış kodunda. README ve SECURITY.md bunu söylüyor, bir de
attestation'ın v0.1.4'ten önce bulunmadığını.

### `npm-pack.sh` ilk gerçek kullanımında iki hata verdi

Kullanıcı script'i v0.1.4 için çalıştırınca ikisi de çıktıda görünür oldu:

1. **"dist/npm holds 14 packages"** — `ls | wc -l` dizinleri *ve* tarball'ları
   sayıyordu. Altındaki liste yedi satırdı; sayı ile liste birbirini
   yalanlıyordu.
2. **Yazdırılan publish komutu çalışmazdı.** `for p in dist/npm/forgelore-*`
   artık tarball'ları da yakalıyor (tarball'ları `dist/npm` içine taşıdığım
   için), ve `cd` bir `.tgz`'ye giremez. Dahası glob sarmalayıcıyı da
   yakalıyordu — yani şans eseri çalışsaydı **sarmalayıcı önce** yayımlanmış
   olabilirdi, ki bu tam olarak kaçınılması gereken sıra.

İkisi de yerel testimden kaçtı, çünkü ben publish komutunu hiç çalıştırmadım
— sadece tarball'lardan kurulumu denedim. Script'in *yazdırdığı* şey de
teslim edilen şeyin parçası.

Düzeltme: sayım yalnızca dizinleri sayıyor, ve publish komutları glob yerine
**tek tek adlandırılıyor** — script platform listesini zaten biliyor.
Yazdırılan her tarball'ın diskte var olduğu doğrulandı.

### npm'e yanlış ikili gitti — benim hatam

`forgelore@0.1.4` npm'de `v0.1.4-1-g9b2beba-dirty` damgalı bir ikili
taşıyor, ve baytları GitHub release'indeki attested baytlarla **aynı değil**.

Zinciri: kullanıcı `release.sh v0.1.4` çalıştırdı, doğru ikililer `dist/`'e
yazıldı. Sonra ben `npm-pack.sh`'deki hatayı düzeltirken `./scripts/ci.sh`
çalıştırdım — ci.sh `crosscheck.sh` çağırıyor, o da `dist/`'teki ikilileri
`git describe` damgasıyla **yerinde yeniden yazıyor**. `SHA256SUMS` eski
haliyle kaldı. Ardından `npm-pack.sh v0.1.4` o ikilileri paketledi.

Go kodu aslında aynı: `9b2beba` yalnızca dokümanlara dokundu, `-dirty` de
benim `npm-pack.sh` düzenlememdi. Ama kullanıcının gördüğü sürüm dizesi
yanlış, ve baytlar attestation'ın kapsadığı baytlar değil — provenance
üzerine çıkılmış bir release için kabul edilemez.

**Koruma eklendi.** `npm-pack.sh` artık paketlemeden önce `dist/`'i kendi
`SHA256SUMS`'ına karşı doğruluyor (çalıştırılamayan çapraz hedefler dahil
her şeyi yakalar) ve host hedefinin sürüm damgasını okuyup istenen sürümle
karşılaştırıyor. Bayat `dist/` ile denendi: reddediyor.

Asıl ders şu: `dist/` bir yapı çıktısı değil, **paylaşılan mutable durum**.
release.sh onu bir anlık görüntü sanıyordu, ci.sh ise çalışma alanı.

---

## 2026-10-05 — v0.1.5: npm düzeltildi, Windows paketleri yeniden adlandırıldı

https://github.com/forgeprint/forgelore/releases/tag/v0.1.5

Kod değişikliği yok; 0.1.4'ün npm tarafını düzelten bir paketleme sürümü.

### npm'e giden baytlar artık attested baytlar

Bu sefer `dist/`'i yerelde derlemedim, **release'ten indirdim**. Yedi paket o
baytlardan kuruldu. Kanıt, kurulum sonrası karşılaştırmayla:

```
npm      70df577968205691cee51949d3b3640c38b4d188d529ecfd0727aa5fe94890d3
release  70df577968205691cee51949d3b3640c38b4d188d529ecfd0727aa5fe94890d3
```

Bu, `npm-pack.sh`'in `dist/` korumasından daha güçlü bir alışkanlık: koruma
yalnızca `dist/`'in kendi içinde tutarlı olduğunu söylüyor, release'ten
indirmek ise yayımlananla aynı olduğunu garanti ediyor. Bir sonraki sürümde
de böyle yapılmalı.

`forgelore@0.1.4` npm'de deprecated, mesajı 0.1.5'e yönlendiriyor. GitHub
release'i v0.1.4 zaten doğruydu, ona dokunulmadı.

### `win32` adı npm'de yayımlanamıyor

`forgelore-win32-x64` iki ayrı gün, iki ayrı yayın turunda **403 "Package
name triggered spam detection"** aldı. Aynı turda diğer dördü sorunsuz
geçti, yani hız sınırı değil, isme bağlı. `win32` token'ı typosquat
paketlerinde sık geçiyor; muhtemel sebep bu.

Paketler `forgelore-windows-x64` ve `-arm64` olarak yeniden adlandırıldı ve
ilk denemede geçti. **İçerideki `"os": ["win32"]` olduğu gibi kaldı** —
npm o alanı `process.platform` ile eşleştiriyor, "windows" yazsaydık paket
Windows'ta hiç seçilmezdi. Yani ad ile alan bilerek uyuşmuyor ve çeviriyi
sarmalayıcı yapıyor. ADR-0024'e bu cümleyle yazıldı, çünkü ileride biri
"tutarsızlık" diye düzeltmeye kalkar.

Filtreyi üçüncü kez yoklamamak üzerine anlaşılmıştı; bir deneme yetti.

### Son kontrol

Boş dizinde registry'den: `npm install forgelore@0.1.5` →
`forgelore v0.1.5 darwin/arm64 go1.26.8`, `--ignore-scripts` ile de aynısı.
Yedi paket de registry'de, `latest` 0.1.5.

### npm yayını yapıldı — sıradaki sefere not

`forgelore` artık npm'de: yedi paket, `latest` 0.1.5, `0.1.4` deprecated.
Yayımlarken öğrenilen, dokümandan çıkmayan şeyler:

**Sıra gerçekten önemli, ve ilk seferinde kaçırıldı.** 0.1.4'te sarmalayıcı
**önce** yayımlandı. Bağımlılıklar `optionalDependencies` olduğu için npm
eksik olanları sessizce atlıyor: o pencerede kuran biri binary'siz bir
sarmalayıcı alırdı. Script bunu yazdırıyordu, yine de oldu — sıralamayı bir
insanın okumasına bırakmak yeterli değil. CI'a taşımanın asıl gerekçesi bu,
provenance ikincil.

**Her `npm publish` ayrı bir tarayıcı doğrulaması istiyor** (2FA açıkken).
Etkileşimsiz bir kabuktan çalıştırınca `EOTP` ile düşüyor, `--otp=` bekliyor.
Yani yayımlama adımı otomatikleştirilmeden önce automation token şart; bu da
CI yolunu kendiliğinden gerektiriyor.

**Registry hemen görünür olmuyor.** `npm publish` "+ paket@sürüm" dedikten
sonra `npm view` birkaç dakika `MISSING` diyebiliyor. Yayımdan hemen sonraki
doğrulama buna takılmamalı; beklemek gerekiyor.

**Aynı `npm deprecate`'i ikinci kez çalıştırmak 422 veriyor.** Zaten
uygulanmış bir deprecation'ı tekrar yazmaya çalışmak hata döndürüyor;
durum bozulmuyor. `npm view <pkg>@<sürüm> deprecated` gerçek cevabı veriyor.

### npm yayımı CI'a taşındı — Trusted Publishing

Kullanıcı Trusted Publisher'ı sordu, ve bir önceki maddenin ("automation
token gerekiyor") varsayımını geçersiz kıldı: **token gerekmiyor.** npm
Temmuz 2025'ten beri OIDC ile yayımlamayı destekliyor, ve provenance'ı
kendiliğinden üretiyor. Yani eksik bıraktığımız iki şey tek çözümle
kapanıyor.

Belgelerden doğrulananlar (ezberden değil):

- npm **≥ 11.5.1**, Node **≥ 22.14**. Node 22 npm 10.x ile geliyor, workflow
  yayımlamadan önce npm'i yükseltmek zorunda — yoksa olmayan bir token
  aranmaya düşüyor.
- Yapılandırma **paket başına** npmjs.com'da: depo + workflow dosya adı.
  Yedi kayıt. Ve kayıtta **`npm publish`'e açıkça izin verilmeli**;
  varsayılan yalnızca `npm stage publish`.
- `repository.url` depoyla birebir eşleşmeli. Bizimki uyuyor.
- Tarball (`.tgz`) ile yayımlamak destekleniyor; provenance tarball'ın
  içinden değil OIDC iddialarından üretiliyor, ve registry'ye yüklenen
  baytlar bizim tarball'ımızın baytları.
- Reusable workflow (`workflow_call`) doğrulamayı şaşırtabiliyor — publish
  komutu üst seviye dosyada kalmalı. Kaldı.

**Tetikleyici `release: published`**, etiket değil. Sen taslak release'i
okuyup yayımladıktan sonra çalışıyor. npm'de geri alma penceresi 72 saat
olduğu için insan kapısının önce gelmesi GitHub'dakinden daha değerli.

İki yapısal kazanç, ikisi de bu oturumda yaptığımız hataların tekrarını
imkânsız kılıyor:

1. İkililer release'ten **indiriliyor**, yeniden derlenmiyor. npm'in
   gönderdiği baytlarla attestation'ın kapsadığı baytlar ayrışamaz.
2. Sıra `npm-pack.sh`'in yazdığı `dist/npm/PUBLISH_ORDER` dosyasından
   okunuyor. Doğru sırayı bir insanın okuyup uygulamasına bırakmak zaten
   yetmemişti.

**Henüz kanıtlanmadı.** Hiçbir release bu yoldan geçmedi; `SECURITY.md`
bunu "ayarlandı, gösterilmedi" diye yazıyor. İlk gerçek sınav bir sonraki
release.

---

## 2026-10-05 — v0.1.6: yayım zinciri uçtan uca otomatik, ve kanıtlandı

https://github.com/forgeprint/forgelore/releases/tag/v0.1.6

Kullanıcı yedi Trusted Publisher kaydını girdi. İlk gerçek sınav:

1. Etiket itildi → `release.yml` derledi, attest etti, taslak çıkardı.
2. Taslak okunup yayımlandı → `release: published` `npm.yml`'ı tetikledi.
3. `npm.yml` 1 dk 7 sn'de bitti: npm yükseltildi, artefaktlar **release'ten
   indirildi**, `npm-pack.sh` paketledi, `PUBLISH_ORDER` sırasıyla yedi
   paket yayımlandı. Tek bir token yok.

Doğrulamalar:

- Yedi paket de registry'de `0.1.6`, `latest` 0.1.6.
- `npm audit signatures` → **"2 packages have verified attestations"**.
  npm provenance artık var; 0.1.4 ve 0.1.5'te yoktu.
- `gh attestation verify` GitHub artefaktında 0 ile çıktı.
- npm'deki ikilinin SHA-256'sı release'inkiyle birebir aynı:
  `8784370d71d74fe2…`.

Yani bu oturumda elle yaparken düşülen iki hata — yanlış baytlar ve yanlış
sıra — artık yapılabilir değil, sakınılması gereken şeyler değil.

### Bugünün özeti: altı sürüm, hepsi aynı aileden

`v0.1.1` ve `v0.1.2` doküman yazarken, `v0.1.3` aracı kullanıcı gibi
kurarken, `v0.1.4` imzalama/npm eklerken, `v0.1.5` npm'e yanlış bayt
gönderdiğim için, `v0.1.6` o hatayı imkânsız kılmak için çıktı.

Ortak nokta: hiçbirinde kod yanlış değildi. Hepsinde eksik olan, kodun hiç
görmediği bir girdi ya da insanın tekrarlaması beklenen bir adımdı. Testler
yazdığımız şeyin çalıştığını kanıtlıyor; anlatmak, kurmak ve yayımlamak
yazmadığımız şeyin eksik olduğunu gösteriyor.

---

## 2026-10-05 — Gemini CLI eklendi, yarısı doğrulanmış

Eşleme resmî kaynaktan yazıldı: `google-gemini/gemini-cli`,
`docs/hooks/reference.md`, **v0.62.0 etiketinde** — main'den değil, kurulu
sürümün etiketinden. Kurulu CLI de 0.62.0.

Yapı Claude Code'a benziyor: `settings.json` içinde `hooks`, olay başına
dizi, araç olaylarında `matcher` **regex**. İki fark: `timeout`
**milisaniye** (Claude Code'da saniye), ve **`scratchpad_dir` yok** — oturum
durumu store'un kendi cache'ine düşüyor.

Kabuk komutu `AfterTool` ile geliyor, `tool_name: run_shell_command`, komut
`tool_input.command`, sonuç `tool_response` içinde `llmContent`,
`returnDisplay` ve **isteğe bağlı** `error`. Ayrı bir hata olayı yok, çıkış
kodu payload'ın hiçbir yerinde yok — yani Claude Code'un durumu. Eşleme
başarısızlığı **iki kez** test ediyor: `field_present: tool_response.error`
ve `output_has_diagnostic`. ADR-0022 ikinci kez işe yaradı.

### Yarısı doğrulandı, ve yarısı neden doğrulanamadı

`SessionStart` ve `SessionEnd` gerçek payload olarak korpusa girdi; ikisi de
`scratchpad_dir`'in yokluğunu doğruluyor.

`AfterTool` **yok**, sebebi hesapta:

```
IneligibleTierError: This client is no longer supported for Gemini Code
Assist for individuals.
```

Google hesabıyla giriş bu istemciden artık modele ulaşmıyor. Oturum hook'lar
ateşlendikten sonra, hiçbir araç çalışmadan duruyor. `verified_against` boş
kaldı, `doctor` "no mapping has been verified against any version" diyor,
tablo "unverified". Önemli olan tek olay belgeye dayanıyor — ve bu projede
belgeye dayanan her olay yanlış çıktı.

Sentetik bir `AfterTool` ile eşlemenin kendi içinde çalıştığını gördüm:
enjeksiyon üretti, `error` alanı olmadan — işi `output_has_diagnostic`
yaptı. Bu eşlemenin tutarlı olduğunu gösterir, Gemini'nin o şekli
gönderdiğini **değil**. Korpusa konmadı.

### Dokümanın iki yanlışı

İkisi de çalıştırınca çıktı, okuyunca değil.

1. **Folder trust varsayılan olarak açık.** `docs/cli/trusted-folders.md`
   "disabled by default" diyor; kod `settings.security?.folderTrust?.enabled
   ?? true` okuyor. Güvenilmeyen klasörde workspace `settings.json` hiç
   okunmuyor — hook'lar yüklenmezdi — ve `--yolo` sessizce onay istemeye
   düşüyor. Geçici dizin hiçbir zaman güvenilmez, o yüzden profil kaynaktan
   bulduğum `GEMINI_CLI_TRUST_WORKSPACE=true` kullanıyor; kullanıcının
   `~/.gemini/trustedFolders.json`'ına dokunmuyor.
2. **`run_shell_command` manuel onay istiyor**, etkileşimsiz yakalamada
   verecek kimse yok — `--yolo`.

Beşinci ajan, beşinci kez: dokümandan okunan ile ajanın yaptığı ayrışıyor.
Bu sefer ayrışma hook formatında değil, ona ulaşmanın önündeki koşullardaydı.

---

## 2026-10-06 — Gemini CLI doğrulandı: tier B, 0.62.0

`AfterTool` yakalandı ve dokümandan yazdığım eşlemeyi **yine** yalanladı.
Beşinci ajan, beşinci kez, ve yine sessizce başarısız olacak türden.

### Belgedeki `error` alanı başarısız komutta hiç yok

Hooks reference "`tool_response` ... ve isteğe bağlı `error`" diyor. `go
build` 1 ile çıktığında payload yalnızca `llmContent` ve `returnDisplay`
taşıyor. Yalnız o alanı test eden bir eşleme her kırık build'i başarı
sayardı.

Ama çıkış kodu payload'da **var**, `llmContent` metninin içinde:

```
<untrusted_context>
Output: # example.com/broken/cmd/app
cmd/app/main.go:4:2: undefined: greet
Exit Code: 1
Process Group PGID: 72958
</untrusted_context>
```

Başarılı komutta `Exit Code` satırı hiç yok. Yani Copilot CLI'ın
(`completed with exit code [1-9]`) tam karşılığı: `output_matches` ile
`Exit Code: [1-9]`.

**`output_has_diagnostic` kaldırıldı ve mapping_version 1'e indi.** Gerçek
bir çıkış kodu varken "çıktı hataya benziyor mu" testini kullanmak gereksiz
yanlış pozitif demekti; `compatibility.md` zaten "ajan sana daha dar bir
test veriyorsa onu tercih et" diyor. Yazdığımız kurala kendimiz uyduk:
mapping yalnızca ihtiyacı olan sürümü taşıyor, yani eski bir Forgelore de
bu dosyayı okuyabiliyor.

### Ölçülen asıl şey: sarmal parmak izine ulaşmıyor

`llmContent` çıktı olarak seçildi (çünkü çıkış kodunu o taşıyor), ama içinde
her çalıştırmada değişen bir `Process Group PGID` satırı var. O parmak izine
girseydi aynı hata her seferinde farklı hash'lenirdi — ve bu **sessizce**
olurdu, çünkü tek tek her arama yine başarılı görünür, sadece hiçbir zaman
eşleşme bulunmazdı.

Ölçüldü: PGID değişse de parmak izi sabit, **ve Gemini'nin sarmalı
çıktısı Claude Code'un temiz çıktısıyla aynı parmak izini veriyor**
(`cd023fb609411574`). Projenin dayandığı "bir hata, bir parmak izi, hangi
ajan bildirirse bildirsin" özelliği ilk kez iki ajan arasında ölçüldü.
`TestGeminiWrappingDoesNotReachTheFingerprint` bunu tutuyor.

### Çözülmemiş bir nokta

Kullanıcının durdurduğu komut 130 ile çıkıyor ve `[1-9]` ile eşleşiyor.
Claude Code bunu `is_interrupt` ile işaretliyor, Gemini'de karşılığı
belgelenmemiş. İptal edilen bir komut başarısızlık olarak hatırlanabilir.
`compatibility.md`'ye yazıldı.

### Yakalamanın önündeki iki engel

Hiçbiri hook'larla ilgili değildi, ikisi de script içinde çözüldü:

- Google hesabıyla giriş modele ulaşmıyor (`IneligibleTierError`).
  `GEMINI_API_KEY` gerekiyor.
- Değişkeni ayarlamak da yetmiyor: CLI yöntemi **birleşik** ayarlardan
  (`security.auth.selectedType`) okuyor, ve bir kez Google ile girmiş makine
  onu kullanmaya devam ediyor. Script artık `GEMINI_API_KEY` varken geçici
  workspace'in kendi `settings.json`'ına `"gemini-api-key"` yazıyor —
  kullanıcının `~/.gemini/settings.json`'ı `oauth-personal` olarak kalıyor.

### Bir güvenlik notu

Kullanıcı anahtarı `export GEMINI_API_KEY="..."` diye terminale yazdı; o
pane benim okuyabildiğim bir yer ve değer orada göründü. Söylendi,
değiştirildi. Bir sonraki sefer için doğru biçim komutun önüne tek
kullanımlık değişken koymak ve satırı geçmişe sokmamak.

---

## 2026-10-06 — Claude Code 2.1.290: drift aracı ilk kez iş gördü

Kullanıcı "claude çalışıyor değil mi" diye sordu. Ölçmeye kalkınca canlı
oturumda bir **ıskalama** çıktı: arama yapıldı (sayaç 16 → 17) ama eşleşme
bulunamadı. Ve kurulu CLI 2.1.290'dı, korpus ise 2.1.289'a doğrulanmış.

`./scripts/drift-payloads.sh claude-code` dört satır döndü ve ikisi gerçek:

### 1. Yanlış alarm: "PostToolUseFailure no longer captured"

Bunu ciddi bir regresyon sandım ve kullanıcıya öyle söyledim. **Değildi.**
Yeni yakalanan payload gösterdi ki model komutu yine borulamış
(`go build ./... 2>&1 | head -40`), yani komut gerçekten 0 ile çıkmış ve
doğru şekilde `PostToolUse` gelmiş.

Bu, aynı tuzağa **üçüncü** düşüşümüz: borulanmış komut önce ADR-0022'ye,
sonra v0.1.3'e, şimdi de bir yanlış drift alarmına yol açtı. Yakalama
promptu artık açıkça "no pipes, no redirection, no extra flags" diyor.
Düzeltince gerçek `PostToolUseFailure` payload'ı ilk denemede geldi.

Ders: drift aracı "alan kayboldu" diyebilir ama bunun iki sebebi olabilir —
ajan değişmiştir, ya da o oturumda o olay hiç oluşmamıştır. Araç ikisini
ayıramaz; ayıran şey yeni payload'a bakmak.

### 2. Gerçek değişiklik: `scratchpad_dir` gitti

Üç olaydan da kalkmış. 2.1.289'da dördü de taşıyordu, 2.1.290'da hiçbiri
taşımıyor. Hiçbir şey kırılmıyor — oturum durumu store'un cache'ine
düşüyor, ki bu zaten belgelenmiş davranıştı — ama **yedek yol artık
olağan yol**.

Bu alan hakkında iki kez yanıldık: önce "hiç gelmiyor" dedik (oturum
açılmamış bir yakalamadan), sonra düzeltip "gerçek oturumda dördü de
taşıyor" dedik. Şimdi üçüncü hali. Test bu yüzden yeniden yazıldı:
`TestTheScratchpadCameAndWent` **iki korpusu birden** tutuyor — 2.1.289'da
var, 2.1.290'da yok, ve her iki durumda durum dosyasının nereye gittiğini
doğruluyor. Artık hiçbir okuma sessizce varsayıma dönüşemez.

### Yakalama script'inde iki düzeltme

- Prompt borulamayı yasaklıyor.
- Çıktı dizini yazmadan önce **temizleniyor**. Payload'lar olay başına
  numaralandığı için, daha az araç çağıran bir yakalama öncekinin
  dosyalarını bırakıyordu; dizin iki oturumu birden tutuyordu ve tek
  oturum gibi okunuyordu. Gerçekten de bir dosya öyle kalmış, elle silindi.

### Durum

Her iki korpus da replay'den geçiyor, `verified_against` 2.1.290 oldu ve
ikisi de saklanıyor — aralarındaki fark zaten bulgunun kendisi.
`compatibility.md` tabloda "two corpora" diyor.

**Açıklanmadan kalan bir şey var:** kullanıcının canlı oturumundaki
ıskalamanın parmak izi `b9bd373fd700b328` idi, ve ne kayıtlı hata, ne
yakalanan 2.1.290 payload'ları, ne de denediğim beş komut biçimi onu
üretiyor. O oturumun payload'ı kaydedilmediği için geriye dönük
bakılamıyor. Uydurmak yerine açık bırakıldı; bir dahaki denemede o projeye
payload döken geçici bir hook koymak yeterli olur.

---

## 2026-10-06 — zincirli komut: ıskalamanın sebebi bulundu ve düzeltildi

Önceki notta "açıklanamadı" diye bıraktığım `b9bd373fd700b328` çözüldü.
Projeye payload döken geçici bir hook konup oturum tekrarlandı, ve model
şunu çalıştırdı:

```
ls -a && go build ./... 2>&1 | head -40
```

`NormalizeCommand` zincirin **ilk** programını alıyordu, yani `ls`. Hata
`go build`'den geliyor ama parmak izi `ls`'e bağlanıyor. Aynı oturumda model
komutu düz çalıştırınca `go`'ya normalleşti ve enjeksiyon oldu — ledger
ikisini de arka arkaya gösteriyor, bu yüzden teşhis tahmin değil.

Ölçülen tablo:

| komut | araç | eşleşme |
|---|---|---|
| `go build ./...` | `go` | ✅ |
| `go build ./... 2>&1 \| head -40` | `go` | ✅ |
| `cd /x && go build ./...` | `cd` | ❌ |
| `ls -a && go build ./...` | `ls` | ❌ |

Boru zaten doğruydu: `| head` tüketici, ilk program üretici. Kırık olan
zincirdi.

### Neden bir tasarım kararını değil, bir boşluğu düzelttik

`NormalizeCommand`'ın kendi doc-comment'i "komut satırını **çalışan araca**
indirger, teşhis sarmalayıcıya değil araca aittir" diyor ve `npx`'i bu
yüzden atlıyor. `cd x && go build` için aynı mantık `cd`'yi de atlamayı
gerektiriyordu. Kilitli bir karara dokunulmadı; yazılı niyete uyuldu.

`lastInChain` eklendi: `&&`, `||`, `;` ayırıcılarının sonuncusundan sonrası
alınıyor, sonra mevcut mantık (env ataması, sarmalayıcı, boru) o segmente
uygulanıyor. İki kenar durum korundu:

- `find . -exec rm {} \;` — kaçışlı noktalı virgül find'ın argümanı,
  bölünmüyor; yoksa geriye hiçbir şey kalmıyordu.
- `go build ./... &&` — boş kuyruk alınmıyor.

**Sezgi olduğu açıkça yazıldı:** `go build ./... && echo ok` zincirinde hata
ilk halkadan gelir ama normalleştirme `echo` der. Daha iyi tahmin olduğu
için seçildi — zincir son komutuna varmak için yazılır, öncekiler hazırlık.

### Doğrulama

Iskalayan **gerçek payload** düzeltilmiş ikiliyle tekrar çalıştırıldı:
enjeksiyon üretti. Testler `fingerprint_test.go`'ya eklendi, gerçek komut
dizesi dahil. Mevcut parmak izleri etkilenmiyor — yalnızca bugüne kadar
hiçbir şeyle eşleşmeyen zincirli komutlarınki değişiyor.

### Yan bulgu, dokunulmadı

`npm test` → `test`, `npm run build` → `run`. Sarmalayıcı atlama mantığı
`npm`'i de runner sayıyor ve ardındaki fiili araç sanıyor. `npx -p
typescript tsc` için doğru, `npm run build` için değil. Bugünkü işin kapsamı
dışında; ayrı bir karar.

---

## 2026-10-06 — npm da düzeltildi: paket yöneticisi aracın kendisidir

Önceki notta "dokunulmadı" diye bıraktığım yan bulgu kapatıldı.

`runners` listesi `npx`, `bunx`, `pnpm`, `yarn`, `npm`'i aynı kefeye
koyuyordu: hepsini atla, sonraki kelimeyi araç say. `npx -p typescript tsc`
için doğru — npx gerçekten komut satırında adı geçen ikiliyi çalıştırıyor.
Ama `npm run build` → `run`, `npm test` → `test`, `npm ci` → `ci`.

`run` bir araç değil. Ve script'in ne çağırdığı komut satırında **hiç
yazmıyor**. Dolayısıyla bir build hatası `run` altında, aynı hatayı veren
bir test `test` altında parmak izleniyordu — yani tek bir hata, onu
tetikleyen her yola bölünüyordu. Fonksiyonun var oluş sebebi tam olarak bunu
engellemek: "fiil düşer, çünkü aynı derleme hatası `go build`, `go test`,
`go run` ve `go vet` üzerinden gelir."

### Ayrım: çalıştırıcı mı, paket yöneticisi mi

- **`executors`** (`npx`, `bunx`): atlanır, adı geçen ikili araçtır.
- **`packageManagers`** (`npm`, `yarn`, `pnpm`, `bun`): **kendisi araçtır**.
  `npm run build` → `npm`, tıpkı `go build` → `go` gibi.
- **`handsOver`** (`exec`, `dlx`, `x`): paket yöneticisinin ikili adı veren
  alt komutları; bunlarda çalıştırıcı gibi davranılır. `pnpm exec tsc` →
  `tsc`, `yarn dlx tsc` → `tsc`, `bun x tsc` → `tsc`.

### Bedeli

`yarn tsc` artık `yarn` veriyor, eskiden `tsc` veriyordu. Yarn 1 doğrudan
ikili çalıştırmaya izin veriyor, ama `yarn build` (script) ile `yarn tsc`
(ikili) komut satırından ayırt edilemiyor — `package.json`'a bakmak
gerekirdi. İkisinden birini seçmek zorundaydık; script biçimi çok daha yaygın.
Zararı sınırlı: hata mesajı farklı olduğu sürece parmak izleri yine ayrı.

Korpus etkilenmedi — içinde yalnızca `npx --yes -p typescript tsc` var ve o
hâlâ `tsc`.

---

## 2026-10-06 — v0.1.7

https://github.com/forgeprint/forgelore/releases/tag/v0.1.7

İki sessiz ıskalama düzeltmesi (zincirli komut, paket yöneticisi), Claude
Code 2.1.290 doğrulaması ve Gemini CLI desteği.

### Zincir ilk kez uçtan uca kendiliğinden işledi

Etiket itildi → `release.yml` derledi, attest etti, taslak çıkardı. Taslak
yayımlandı → `npm.yml` tetiklendi, artefaktları release'ten indirdi,
paketledi, `PUBLISH_ORDER` sırasıyla yedi paketi yayımladı. **Elle tek bir
komut yok, tek bir token yok.** v0.1.6'da kurulmuştu, burada ikinci kez ve
sorunsuz çalıştı.

### Yayımlanan paketle doğrulandı

- Yedi paket de `0.1.7`, `latest` 0.1.7.
- `npm audit signatures` → "2 packages have verified attestations".
- Registry'den kurulan ikili: `forgelore v0.1.7 darwin/arm64 go1.26.8`.
- **İki düzeltme yayımlanan ikilide çalışıyor:**
  `ls -a && go build ./... | head -40` → araç `go`, eşleşme 1 (eskiden
  `ls`, eşleşme 0); `npm run build` → araç `npm` (eskiden `run`).

### Notlarda açıkça söylenen bir şey

Mevcut kayıtların etkilenip etkilenmediği sorusu release notunda ayrı bir
başlık: düz komutla kaydedilmiş hiçbir parmak izi değişmiyor, yalnızca
bugüne kadar hiçbir şeyle eşleşmeyenler değişiyor. Zincirli ya da `npm run`
ile kaydedilmiş bir fix varsa `forgelore dedupe` çifti bulur.

---

## 2026-10-06 — Forgelore kendi deposunda kuruldu, ve ilk dakikada bir sınır buldu

Kullanıcı "bu yazışmada kendimiz için kullanalım, hem denemiş oluruz"
dedi. `forgelore init --with-git-hook` çalıştırıldı; `.forgelore/records/`
commit'lenir, index/ledger/local gitignore'lu, `pre-commit` her commit'te
`forgelore check` çalıştırıyor.

### Bulgu: bugünkü hataların hiçbiri parmak izi üretmiyor

Kaydetmek istediğim beş gerçek hata vardı. Hiçbiri `fix` olamadı, çünkü
`fingerprint.Scan` hiçbirinde bir şey bulmuyor:

```
Error authenticating: IneligibleTierError: ...   (yok)
npm error code E403 ... spam detection           (yok)
npm error code EOTP                              (yok)
Failed to authenticate: OAuth session expired    (yok)
the working tree is not clean                    (yok)
```

Sınırı ölçtüm: eşleştiriciler **derleyici ve çalışma-zamanı teşhislerini**
tanıyor — `dosya.go:4:2: ...`, `TypeError: ...`, `panic:`, `--- FAIL:`.
CLI'ların düzyazı hatalarını tanımıyor. `IneligibleTierError: ...` satır
başındayken eşleşiyor (`16e2c186...`), ama gerçek çıktıda önünde
`Error authenticating: ` olduğu için eşleşmiyor. `npm error code E403`
küçük harfli ve iki nokta kalıbına uymadığı için hiç görünmüyor.

Bu bir hata değil, kapsamın sınırı: korpus 29 ailenin hepsi derleyici /
runtime hatası. Ama şunu söylüyor — **Forgelore, kendi geliştiricisinin bütün
gün karşılaştığı hataları göremiyor.** Dogfooding'in bulması gereken şey
tam olarak buydu ve ilk dakikada buldu.

### Yapılan

Beşi de `decision` olarak kaydedildi (parmak izi gerektirmiyor, oturum
başlangıcı indeksinde başlıkla görünüyor): claude CLI'ın ayrı girişi,
`dist/`'in paylaşılan değişken durum olması, npm'in `win32` adını
reddetmesi, Gemini'nin API anahtarı ve folder trust gereksinimi, ve
ajanların komut zincirleme/borulama alışkanlığı.

`CLAUDE.md`'ye kısa bir bölüm eklendi: deponun kendi ürününü kullandığı,
hangi komutların işe yaradığı, ve yukarıdaki sınırın açıkça yazılması —
yoksa bir sonraki oturum `fix` kaydetmeye çalışıp sessizce başarısız olur.

### Açık soru

Eşleştiriciler CLI hatalarını kapsayacak şekilde genişletilsin mi? Kazancı
büyük: enjeksiyon yolu bugün yalnızca derleyici hatalarında çalışıyor.
Riski de büyük: `Error:` ile başlayan her satırı yakalamak yanlış pozitif
üretir ve `onCommandFailed`'ın ucuz olmasının sebebi tam da bu darlık.
Yapılırsa korpusa yeni aileler ve bir ADR gerekir. Karar verilmedi.

---

## 2026-10-06 — eşleştiriciler CLI hatalarını öğrendi (ADR-0025)

Dogfooding'in bulduğu boşluk kapatıldı. Dört değişiklik, hepsi **gerçek
araç çıktısından** tasarlandı — önce `npm run nope`, `npm install <yok>`,
`git status`, `git checkout nope` çalıştırılıp çıktıları okundu, sonra
desen yazıldı.

- **`toolErrorCode`** — `npm error code E403`. Yalnızca kod satırı
  alınıyor; npm'in düzyazı satırlarında zaman damgalı log yolu var, onu
  yakalamak her çalıştırmada farklı hash üretirdi.
- **`cliErrorLine`** — satır başında `error:` ya da `fatal:`, rustc'nin
  `error[E0433]:` biçimi dahil. git, cargo, rustc, clang böyle yazıyor.
- **`failedToLine`** — `Failed to …`. Buradaki en gevşek desen, ve tek
  cümleden başka bir şey vermeyen araçları gören tek desen.
- **Öndeki etiket düşürülüyor.** `exceptionLine` artık bir `Şey: ` önekini
  kabul ediyor ve yalnızca tanımlayıcıdan sonrasını saklıyor. Gemini'nin
  `Error authenticating: IneligibleTierError: …` çıktısı, çıplak
  `IneligibleTierError: …` ile **aynı** parmak izini veriyor; test bunu
  tutuyor. Etiket hatayı bildirenin, hatanın değil.

Korpusa iki gerçek aile eklendi: `npm/registry-404` (varyantlar farklı
paket adıyla, parmak izi `code E404`'te buluşuyor) ve
`git/not-a-repository` (varyantlar birebir aynı, `variantsIdentical`'a
yazıldı — test yapmadığı bir karşılaştırmayı yapmış gibi görünmesin).

### Değişen bir güvence, gizlenmeden

`internal/agent`'ta bir test düştü: `TestSuccessIsNotMistakenForFailure`.
Fixture'ı tam olarak `error: nothing is wrong, this word just appears` idi.
Artık satır başındaki `error:` gerçekten hata sayılıyor, ve `claude-code`
eşlemesi `output_has_diagnostic` kullandığı için bu ajan davranışını da
değiştiriyor.

Test silinmedi, **daraltılmış güvenceyi** söyleyecek şekilde yeniden
yazıldı: "error kelimesi hiçbir şey kanıtlamaz" değil, "**cümle ortasındaki**
error kelimesi hiçbir şey kanıtlamaz". İki vakayı birden tutuyor. Satır
başına çapalama, yanlış pozitife karşı tek savunma, o yüzden teste yazıldı.

### Ölçülen yanlış pozitif kontrolü

Dokuz gürültü örneği denendi ve hiçbiri teşhis üretmedi: npm'in log yolu
satırı, boş `npm error`, girintili devam satırı, `exit status 1`, ilerleme
çıktısı, banner, `warning:` satırı, go test özeti, ve cümle içinde geçen
"Error" kelimesi.

### Sonuç: bugünün hataları artık kayıtlı

Dört `fix` kaydı parmak iziyle girdi — npm 403 (win32 adı), npm EOTP,
claude OAuth, Gemini IneligibleTierError. Tekrar denendi: npm 403 **başka
bir paket adıyla** geldiğinde bile eşleşiyor, çünkü parmak izi `code E403`'e
bağlı. Depoda artık 9 kayıt var (4 fix, 5 decision).

`the working tree is not clean` hâlâ tanınmıyor ve bu doğru — diagnostik
bir şekli yok, onu yakalamak her cümleyi yakalamak olurdu. ADR-0025 bunu da
yazıyor.

---

## 2026-10-06 — v0.1.8

https://github.com/forgeprint/forgelore/releases/tag/v0.1.8

ADR-0025: eşleştiriciler CLI hatalarını tanıyor. Zincir yine elle hiçbir
komut olmadan işledi (üçüncü kez).

Yayımlanan paketle doğrulandı:

- Yedi paket de `0.1.8`, `latest` 0.1.8, `npm audit signatures` → verified
  attestations.
- Registry'den kurulan ikili `v0.1.8 darwin/arm64 go1.26.8`.
- **Yeni eşleştiriciler yayımlanan ikilide çalışıyor:** `npm error code
  E403` → `code E403`, `fatal: not a git repository…` → mesaj, etiketli
  `Error authenticating: IneligibleTierError: …` → etiket düşmüş hâli.
- Ve bu depodaki gerçek kayıtla eşleşiyor: `code E403` → "Name the Windows
  npm packages windows, not win32".

Yani bu release'ten itibaren Forgelore kendi geliştirme gününün hatalarını
hatırlayabiliyor; bu sabah hatırlayamıyordu.

### Not: registry yayılımı yine gecikti

Yedi paketten altısı hemen göründü, `forgelore-windows-x64` birkaç dakika
`MISSING` dedi. v0.1.5'te de olmuştu. Yayım sonrası doğrulama bunu
beklemeli; workflow'un "+ paket@sürüm" demesi registry'de görünür olması
demek değil.

---

## 2026-10-06 — Forgelore bu oturumda çalışıyor, ve ölçüm iyi değil

Kullanıcı "şu an bu konuşmada forgelore kullanılıyor mu" diye sordu.
Cevap evet, ve kanıtı ledger'da: bu oturumun kimliğiyle (`80fc4b61-…`)
**13 satır**. Hook'lar benim Bash komutlarımda ateşleniyor.

Ama rakamlar şöyle:

| kaynak | satır | sonuç |
|---|---|---|
| hook (bu oturum) | 13 | **13 ıskalama, 0 enjeksiyon** |
| elle `recall` (benim testlerim) | 15 | 11 ıskalama, 4 enjeksiyon |

**13 aramanın hiçbiri gerçek bir başarısızlık değildi.** Oturum durum
dosyasına bakınca ne olduğu görülüyor: hepsi benim kendi script'lerimin
*çıktısı*. Eşleştirici sınırlarını ölçerken ekrana şunu basmıştım —

```
IneligibleTierError: ... (satır başı) 16e2c18620229d77 IneligibleTierError: …
```

Satır başında bir exception şekli var, dolayısıyla Forgelore kendi test
çıktısını hata saydı. Aynı şekilde `error: pathspec …` içeren fixture'ları
*yazarken* de.

### Bu tam olarak bugün bedel diye yazdığımız şey

ADR-0022 "başarılı ama hata biçimli çıktı veren komut başarısız sayılır"
diyor, ADR-0025 "satır başındaki `error:` artık hata" diyor. İkisi de
doğruydu. Artık tahmin değil, **ölçüm**: `output_has_diagnostic` bu
oturumda 13 arama üretti ve hiçbiri gerçek değildi.

Bugünkü zarar sınırlı — hepsi ıskalama, enjeksiyon yok, sadece ledger
gürültüsü. Ama bir gün o parmak izlerinden biri bir kayda denk gelirse,
hiçbir şeyin başarısız olmadığı bir yere ipucu girer. Ve `nothing known`
sayacı şişerek raporu yanıltır: 28 aramanın 24'ü "bilinmiyor" diyor, oysa
24'ünün çoğu zaten hata bile değildi.

### İkinci ayrıntı: durum dosyası komutu olduğu gibi saklıyor

Bir oturum durum dosyası **17 KB**, en uzun komut girişi **4254 karakter** —
benim heredoc'larım. Diğer 17 dosya 48–155 bayt arasında, yani sorun
komutun uzunluğuyla orantılı ve normalde görünmüyor. Yine de parmak izi
zaten normalleştirilmiş komutu kullanıyor; ham metni saklamanın tek
sebebi "düzeldi" önerisinde komutu geri gösterebilmek.

### Sıradaki

Daraltma konuşulacak. Gerilimin iki ucu da bugün bizim eserimiz:
`output_has_diagnostic` olmazsa borulanmış build yine görünmez olur,
olursa hata hakkında *konuşan* çıktı da hata sayılır. Karar verilmedi.

---

## 2026-10-06 — çıkarsanmış başarısızlık artık iz bırakmıyor

Daraltma kararı: yetenek kalsın, yanlış pozitifin bedeli sıfırlansın.

### Önce bir öneriyi ölçüp çöpe attım

İlk aklıma gelen daraltma şuydu: `output_has_diagnostic`'e yalnızca komut
çıkış kodunu maskeliyorsa (boru ya da zincir varsa) başvur. Oturum durum
dosyasına bakınca bu oturumdaki **12 aramanın 12'si de** zaten boru veya
zincir içeren komutlardandı. Yani kural hiçbir şey kazandırmazdı. Öneri
geri çekildi — tahmin edilmiş bir iyileştirme, ölçülünce sıfır çıktı.

Daha derin sorun: yanlış pozitifler **zayıf** desenlerden değil, güçlü
desenlerden geldi. `IneligibleTierError: …` gerçek bir exception şekli,
sadece veri olarak basılmıştı. Payload'da "bu hata mı, hatadan bahseden
çıktı mı" ayrımını yapacak sinyal yok; desen sıkılaştırmak çözmez.

### Yapılan

`Event.Inferred` eklendi: başarısızlığa ajanın söylediği için değil,
çıktının şekline bakılarak karar verildiyse işaretli. Yalnızca
`output_has_diagnostic` çıkarsıyor; `always`, `field_present` ve
`output_matches` ajanın payload'a koyduğu bir şeyi okuyor.

`onCommandFailed` artık **çıkarsanmış ve eşleşmeyen** bir olayda hiçbir şey
yazmıyor: ne ledger satırı, ne oturum durumu. Arama yine yapılıyor;
eşleşme varsa enjeksiyon da ledger kaydı da duruyor, çünkü o zaman
gerçekten bir şey biliniyordu.

Yan etki, bilinçli: çıkarsanmış bir başarısızlık oturum durumuna
yazılmadığı için o olaydan "düzeldi" önerisi çıkmaz. Başarısız olduğundan
emin değilsek, düzeldiğinden de emin olamayız.

### Kalan risk, saklanmadan

Eşleşen bir tahmin hâlâ hiçbir şeyin bozulmadığı bir yere tek satır
sokabilir. Bu kapanmadı, kapanamaz — ADR-0022'ye bu haliyle yazıldı,
ölçümle birlikte.

---

## 2026-10-06 — CI kırmızı oldu ve altından gerçek bir hata çıktı

v0.1.9 için etiket atmadan önce CI'a baktım: `TestHookRecordsLatency`
düşmüş, "got 0 entries". Yerelde 20/20 geçiyordu.

Sebep zamanlama ama test kusuru değil. `hook.deadline_ms` (varsayılan 500)
bütçesi, `onCommandFailed`'ın döngüsünün **başında** kontrol ediliyordu.
Pahalı iş ise indeksi açmak ve o döngüden **önce** oluyor. Yani bütçeyi
indeksi açarken tüketen bir makinede döngü hiç dönmüyor:

- enjeksiyon yok,
- ledger satırı yok,
- ve `report` gerçekte olandan daha az arama gösteriyor, **farkı hiçbir
  yerde söylemeden.**

1 ms deadline ile birebir üretildi: ledger dosyası hiç oluşmuyor.

Bu, bedelini ödeyip malı çöpe atmak: indeks zaten açılmış, aramanın kendisi
milisaniyeler. Düzeltme — bütçe artık **kaç arama yapılacağını** sınırlıyor,
**hiç yapılıp yapılmayacağını** değil. İlk olay her zaman aranıyor,
deadline ikinciden itibaren işliyor.

Yavaş bir CI runner'ının bulduğu, hiçbir dizüstünün bulamadığı bir hata.
Test artık 1 ms bütçeyle hem enjeksiyonu hem ledger satırını istiyor.

---

## 2026-10-06 — v0.1.9

https://github.com/forgeprint/forgelore/releases/tag/v0.1.9

İki düzeltme, ikisi de Forgelore'u kendi deposunda çalışır bırakıp ne
kaydettiğine bakmaktan çıktı: çıkarsanmış ıskalamanın iz bırakmaması, ve
bütçesi dolmuş bir hook'un hiç ölçmemesi.

Yayımlanan paketle doğrulandı:

- Yedi paket de `0.1.9`, `latest` 0.1.9, `npm audit signatures` → verified
  attestations, ikili `v0.1.9 darwin/arm64 go1.26.8`.
- **1 ms bütçe** ile enjeksiyon geldi ve ledger'a 1 satır yazıldı — düzeltme
  öncesi ikisi de sıfırdı.
- **Çıkarsanmış ıskalama** hiçbir çıktı üretmedi ve ledger dosyası hiç
  oluşmadı.

### Bugünün son dersi

CI kırmızıya döndü ve altından gerçek bir ürün hatası çıktı. Testi "flaky"
diye işaretleyip geçmek mümkündü; bütçe kontrolünün yanlış yerde olduğunu
gösteren şey, 1 ms ile birebir üretebilmek oldu. Yavaş bir runner, hiçbir
dizüstünün bulamayacağı bir şeyi buldu.

### Kalanlar — hepsi `[SEN]`
- İki kişilik bir haftalık ekip denemesi (`docs/team-trial.md`).

---

## 2026-10-06 — günün özeti

51 commit, **on sürüm** (v0.1.0 → v0.1.9), yedi yeni ADR (0019–0025).
Korpus 29 → 31 aile, ajan sayısı 3 → 5, ve proje kendi ürününü kullanmaya
başladı.

### Hataların tek bir ailesi vardı

On sürümün dokuzu bir düzeltme taşıdı ve **hiçbirinde kod yanlış
çalışmıyordu**. Hepsinde eksik olan, kodun hiç görmediği bir girdiydi:

| sürüm | görmediğimiz girdi |
|---|---|
| v0.1.1 | `go vet`'in çıktı biçimi |
| v0.1.3 | `head`'e borulanmış komutun sıfırla çıkması |
| v0.1.5 | `dist/`'in araya giren bir `ci.sh` ile yeniden yazılması |
| v0.1.7 | ajanın komutu `cd x &&` ile zincirlemesi; `npm run build`'in alt komutu |
| v0.1.8 | CLI'ların düzyazı hataları |
| v0.1.9 | bütçesi indeksi açarken dolan bir hook |

Ortak nokta: hiçbiri testle bulunmadı. **Testler yazdığımız şeyin
çalıştığını kanıtlıyor; yazmadığımız şeyin eksik olduğunu başka şeyler
gösteriyor** — ve bugün dördü de ayrı ayrı iş gördü:

1. **Anlatmak.** README demosunu yazarken `go vet` tanınmadığı görüldü.
2. **Kullanmak.** Eklentiyi marketplace'ten kurup gerçek bir oturumda
   çalıştırmak borulanmış komutu ortaya çıkardı.
3. **Yayımlamak.** npm'e yanlış bayt göndermek `dist/`'in paylaşılan
   değişken durum olduğunu gösterdi.
4. **Kendi ürününü kullanmak.** Forgelore kendi deposuna kurulunca o günün
   hatalarının hiçbirini kaydedemedi.

### Dört kez dokümana güvenip dört kez yanıldık

Bugün beşinci ve altıncı ajan turunda da aynısı oldu: Gemini'nin belgelediği
`error` alanı başarısız komutta yok (çıkış kodu `llmContent` metninde), ve
folder trust "varsayılan kapalı" diye yazılmışken kodda açık. Kural —
**ajanın hook'u hakkında ezberden ya da yalnızca dokümandan yazma** — altı
ajanda altı kez karşılığını verdi.

### Benim hatalarım

Saklanacak bir şey değil, bir sonraki oturum için uyarı:

- `gofmt -w .` ile `vendor/`'u yeniden biçimlendirmek (CI yakalayamıyor).
- Marketplace PR'ını dalı base'e sıfırlayarak kapattırmak.
- npm'e geliştirme damgalı ikili göndermek — araya kendi `ci.sh`
  çalıştırmam girdiği için. v0.1.5 bunu düzeltmek için çıktı.
- "`PostToolUseFailure` artık ateşlenmiyor" diye bir regresyon ilan etmek;
  model komutu borulamıştı.
- Komut şekline göre daraltma önermek; ölçünce 12/12 işe yaramadığı çıktı.

Son ikisi aynı dersin iki yüzü: **önce ölç, sonra söyle.** Her ikisinde de
ölçüm kendi önerimi çürüttü, ve ikisinde de bunu söylemek düzeltmekten daha
değerliydi.

### Codex CLI denendi: doğrulanamadı, ama varsayımlar ikiliden sınandı

Giriş yapıldı (ChatGPT; ilk 401'ler girişin hemen ardından geçiciydi),
oturum çalışıyor ve model `go build ./...` ile `go version`'ı gerçekten
koşuyor — yani yakalama promptu iyi. **Ama tek bir payload düşmedi.**

Doküman yardım etmedi: `openai/codex` deposunun `docs/` dizininde **hooks
sayfası yok**. Yani bu eşleme en baştan zayıf temelliydi, ve bunu artık
biliyoruz. Kalan tek güvenilir kaynak ikilinin kendisi oldu.

**İkiliden doğrulananlar** (`strings`, 0.160.0):

- Olay adları: `PreToolUse PermissionRequest PostToolUse PreCompact
  PostCompact SessionStart SessionEnd UserPromptSubmit SubagentStart
  SubagentStop Stop Interrupt`. Eşlemedeki üç ad **doğru**.
- Payload alanları: `session_id`, `transcript_path`, `cwd`,
  `hook_event_name`, `permission_mode`, `turn_id`, `model`, `reason`,
  `tool_input`, `stop_hook_active`. Yanıt `hookSpecificOutput` /
  `additionalContext`. Yani biçim Claude Code ailesinden.
- `hooks` özellik bayrağı 0.160.0'da **varsayılan açık** — `-c
  features.hooks=true` gereksizmiş.
- Yapılandırma `hooks.json`, ve yapı `matcher` + `hooks` gruplarına
  benziyor (`HookHandlerConfig`, `HookStateToml`, `trusted_hash`).

**Çözülemeyen:** dosyanın yeri ve tam biçimi. `$CODEX_HOME/hooks.json` ile
üst düzey olay anahtarları hiçbir şey tetiklemedi. Sarmalanmış
(`{"hooks": {...}}`) biçimi denendiğinde oturum **15 dakika askıda kaldı**
ve öldürüldü — bu da Codex'in dosyayı okuduğunu ve bir şeyin (muhtemelen
hook trust, bypass bayrağına rağmen) bloke ettiğini düşündürüyor.

**Script'e giren düzeltme — bir tane:** `codex exec` model komutlarını
varsayılan olarak kum havuzunda çalıştırıyor, bu yüzden
`--dangerously-bypass-approvals-and-sandbox` eklendi.

**Script'e girmeyen bulgu:** geçici bir `CODEX_HOME` + gerçek `auth.json`'a
**sembolik bağ** oturumu açık tutuyor (elle doğrulandı: `codex login
status` → "Logged in"). Kimlik kopyalanmıyor, kullanıcının yapılandırmasına
dokunulmuyor — Copilot'taki `COPILOT_HOME` kalıbının Codex karşılığı. Koda
geçirilmedi, çünkü `hooks.json`'ın o home içinde nereye konacağı hâlâ
bilinmiyor; yarısı bilinen bir yolu script'e yazmak işe yarıyormuş gibi
görünür.

`verified_against` boş kaldı, tablo "unverified" diyor. Her deneme bir model
çağrısı yakıyor; şekli tahmin ederek devam etmek pahalı. Ucuz yol TUI'nin
hook ekranı: ikilide "No hooks installed for this event", "New hook — review
required", "Trust" dizgeleri var, yani etkileşimli arayüz dosyanın okunup
okunmadığını model çağırmadan gösteriyor.

### Eşleme şemaya göre yeniden yazıldı — üç yanlış, biri sessiz olurdu

Yakalama başarısız olsa da ikilideki JSON şemaları (`post-tool-use.command.input`
ve çıktı ikizi) eşlemenin dayandığı üç varsayımı çürüttü:

1. **Üst düzey `error` yok.** Başarısızlık testimiz `field_present: error`
   idi; hiç ateşlenmezdi. Artık `output_has_diagnostic` kullanılıyor, çünkü
   şema `tool_response`'un içini kısıtlamıyor.
2. **`tool_output` yok**, sonuç `tool_response`'ta.
3. **`interrupted` alanı yok**, `Interrupt` ayrı bir olay. Taşıdığımız
   `skip_when` de hiç ateşlenmezdi.

Ve sessiz olacak olanı: eşleme bağlamı **düz `additionalContext`**'e
yazıyordu. Codex onu `hookSpecificOutput.additionalContext` içinde okuyor —
Claude Code'un yeriyle aynı. Yani enjekte edilen her ipucu Codex'in hiç
bakmadığı bir yere gidecek, ve **hiçbir hata vermeyecekti.** Beşinci kez
aynı aile; bu sefer yayımlanmadan önce yakalandı.

İki test de buna göre yeniden yazıldı. `TestDocumentedPayloadsTranslate`
artık `TestSchemaShapedPayloadsTranslate`, ve `interrupted` testinin yerini
bağlamın doğru yere yazıldığını kontrol eden bir test aldı.

`verified_against` hâlâ boş: şema bir sözleşme, yakalanmış payload değil.
Ama eşleme artık dokümandan değil, ajanın kendi şemasından türetilmiş.

### Kapanmayan bir risk

`output_has_diagnostic` bir tahmindir. Payload'da "bu hata mı, hatadan
bahseden çıktı mı" ayrımını yapacak sinyal yok. Bedeli sıfırlandı (eşleşmeyen
tahmin iz bırakmıyor) ama kapatılmadı: eşleşen bir tahmin hâlâ hiçbir şeyin
bozulmadığı bir yere tek satır sokabilir. ADR-0022 bunu ölçümle birlikte
yazıyor.
- Codex CLI doğrulaması, erişim olduğunda:
  `./scripts/capture-agent-events.sh codex-cli`, sonra eşlemeyi düzelt,
  `verified_against`'i doldur, `docs/compatibility.md`'yi güncelle.
- Cursor: araştırıldı, başlanmadı.
