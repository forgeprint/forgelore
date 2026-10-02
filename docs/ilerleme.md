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

### Sırada

Faz 3 — ölçüm altyapısı. `docs/plan.md`'deki adımlara bakılacak; Faz 2'den
devreden açık karar kalmadı.

### Faz 2 — açık maddeler

- `[SEN/CLAUDE]` `go/unknown-import` yeniden yakalanmalı: farklı import yolu,
  kayan satır. Windows makinesi gerekiyor.
- Külliyatta gerçek projelerden hata yok — hepsi küçük ve kasıtlı örnekler.
  `testdata/errors/README.md` bunu zaten söylüyor; parmak izi kurallarının
  asıl sınavı o malzeme.
- `config.yaml` okunmuyor: ADR-0018 yazılı ama uygulanmadı, `init` de bu
  yüzden dosyayı oluşturmuyor. Bayrak > ortam > yerel > ekip > varsayılan
  önceliği hâlâ kodda yok.
