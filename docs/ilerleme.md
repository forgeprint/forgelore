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
