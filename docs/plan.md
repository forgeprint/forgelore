# Forgelore — Proje Planı

> Proje adı: **Forgelore**. CLI komutu: `forgelore`. Go modülü: `github.com/forgeprint/forgelore`. Kullanıcı projesindeki klasör: `.forgelore/`.

---

## Claude Code için çalışma kuralları (ÖNCE BUNU OKU)

Bu proje **adım adım, her adımda kullanıcıya sorularak** ilerler. Kurallar:

1. **Her adıma başlamadan önce sor.** Adımı kısaca özetle: ne yapacaksın, hangi dosyalara dokunacaksın, hangi kararlar gerekiyor. Kullanıcı onay verene kadar kod yazma, dosya oluşturma, komut çalıştırma.
2. **Her adımın sonunda rapor ver.** Ne yaptın, hangi testler geçti, kabul kriterlerinin durumu ne, açık kalan ne. Sonraki adıma kendi başına geçme.
3. **`[SEN]` işaretli işler kullanıcıya aittir.** Bunları yapma. Kullanıcıya ne yapması gerektiğini net talimatlarla söyle ve tamamlandığını teyit etmesini bekle.
4. **`[CLAUDE]` işaretli işler sana aittir,** ama yine de 1. kurala tabidir.
5. **"Kilitli kararlar" bölümündeki maddeleri tartışmaya açma.** Bir kararın uygulanamaz olduğunu düşünüyorsan, gerekçesiyle kullanıcıya bildir ve cevabını bekle. Kendi başına değiştirme.
6. **Yeni bağımlılık ekleme yasağı.** Go standart kütüphanesi ve `modernc.org/sqlite` dışında hiçbir bağımlılık eklenmez. Eklenmesi gerektiğini düşünüyorsan önce kullanıcıya sor. Onaylanırsa bir ADR yaz ve bağımlılığı `vendor/` altına al.
7. **Ajanlar hakkında hafızadan bilgi kullanma.** Bir ajanın (Claude Code, Codex, Gemini CLI vb.) hook formatı, olay adları ya da dosya yolları hakkında kod yazmadan önce, o ajanın **güncel resmî dokümantasyonunu** kontrol et, bulduğun kaynağı rapora yaz. Emin olamadığın alanı "doğrulanmadı" olarak işaretle.
8. **Belirsizlikte tahmin yürütme, sor.** Soruyu kod yorumuna gömme, kullanıcıya sor.
9. Her fazın sonunda `docs/ilerleme.md` dosyasını güncelle: tamamlanan adımlar, alınan kararlar, açık sorular. Yeni bir oturumda bu dosyayı okuyarak kaldığın yerden devam et.

---

## 1. Ürün tanımı

Forgelore, kodlama ajanları için **yerel, ekip paylaşımlı, bağımsız** bir hafıza ve bağlam bütçesi katmanıdır. Amacı token kullanımını azaltmaktır ve bunu **ölçerek kanıtlar**.

Ne yapar:
- **Olay tetiklemeli, tam zamanında hafıza.** Bir komut başarısız olduğunda hatanın parmak izini çıkarır. Aynı hata daha önce görülmüş ve çözülmüşse, o an tek satırlık bir ipucu enjekte eder ("Bu hata daha önce görüldü, çözüm: …; denenip işe yaramayanlar: …"). Eşleşme yoksa hiç token harcamaz.
- **Çıkmaz sokakları hatırlar.** Denenmiş ve işe yaramamış yaklaşımlar kod tabanından çıkarılamaz, ama tekrarlanan hata döngüleri en çok token yakan şeydir.
- **Kademeli erişim.** Oturum başında bütçesi sınırlı küçük bir dizin yükler. Ayrıntılar yalnızca istenince getirilir. Bu disiplin araç tasarımıyla zorlanır: dizinden kimlik almadan ayrıntı çekilemez.
- **Fayda takibi.** Her kaydın ne kadar işe yaradığı izlenir. Faydasız kayıtlar zamanla dizinden düşer.
- **Ölçüm (A/B).** İsteğe bağlı olarak bazı oturumlarda enjeksiyonu kapatır ve token kullanımını karşılaştırır. Kullanıcıya "bu araç sana şu kadar kazandırdı ya da kaybettirdi" raporu verir.

Ne yapmaz:
- Arka planda yapay zekâ çağrısı yapmaz. Yakalama deterministiktir.
- Ağa çıkmaz, telemetri göndermez.
- Ajanı asla engellemez (bkz. kilitli kararlar).
- Ajanların yerleşik hafıza özelliklerini (ör. Claude Code auto memory) kopyalamaz. Onların yapmadığını yapar.

---

## 2. Kilitli kararlar

| # | Karar | Gerekçe |
|---|---|---|
| K1 | Dil: **Go**. Çıktı: `CGO_ENABLED=0` ile derlenmiş **tek statik binary**. | Kullanıcının makinesinde hiçbir çalışma zamanı gerekmez. Go 1 uyumluluk sözü. Hook başlangıç süresi düşük. |
| K2 | SQLite: **`modernc.org/sqlite`** (saf Go, FTS5 dahil). | C derleyicisi gerekmez, her platforma çapraz derlenir. |
| K3 | Bağımlılıklar: yalnızca stdlib + K2. Hepsi `go mod vendor` ile depoda. | İnternetteki bir paket değişse ya da silinse bile derleme bozulmaz. |
| K4 | **MCP sunucusu kendi yazımımız:** stdio üzerinden minimal JSON-RPC. SDK kullanılmaz. | Sıfır bağımlılık. İhtiyaç 3–5 araç. |
| K5 | **Adaptörler kod değil, veri:** her ajan için binary'ye gömülü bir eşleme dosyası. Depodaki bir dosyayla yerelde geçersiz kılınabilir. | Ajan bir olay adını değiştirdiğinde yeniden derleme gerekmez. |
| K6 | **Evrensel taban CLI'dır.** Her özellik önce bir CLI komutu olarak var olur. Hook ve MCP bunun üstüne eklenen katmanlardır. | Her kodlama ajanı shell komutu çalıştırabilir. |
| K7 | **Fail-open:** hook'lar her koşulda başarıyla çıkar, katı zaman aşımıyla çalışır. Hata olursa hiçbir şey enjekte etmez, sessizce devam eder. | Hafıza yardımcı katmandır, ajanın çalışmasını asla durdurmaz. |
| K8 | **Arka planda LLM çağrısı yok**, sürekli çalışan servis (daemon) yok. | Maliyet ve kararlılık. |
| K9 | **Doğruluk kaynağı düz markdown**, kayıt başına bir dosya. SQLite indeksi türetilmiştir, git'e girmez, her an yeniden üretilebilir. | Git ile ekip paylaşımı, merge çakışmasız, insan tarafından okunabilir. |
| K10 | İki kapsam: **ekip** (commit edilir, PR ile incelenir) ve **yerel/kişisel** (gitignore). | Bir kişinin yanlış kaydı herkesi kirletmez. |
| K11 | Her kayıtta **`schema` sürümü**. Tanınmayan alanlar okunur ve **korunur**, silinmez. | İleri ve geri uyumluluk. |
| K12 | **Ölçüm ilk günden çekirdekte.** Enjekte edilen her bayt bir defterde kaydedilir. | Tasarruf iddiası ancak ölçülürse gerçektir. |
| K13 | **Dış içerik otomatik kalıcı olmaz.** Web'den ya da ağdan okunan içerikten türeyen kayıt adayları yalnızca öneri olarak tutulur. Gizli anahtar desenleri maskelenir. `<private>` etiketi içindeki içerik hiç kaydedilmez. | Kalıcı prompt injection riskine karşı. |
| K14 | Lisans: **Apache-2.0**. Katkılar: DCO (`git commit -s`). | Açık ve benimsemeyi kolaylaştıran lisans, patent hükmü içerir. |
| K15 | Proje, **forgeprint organizasyonu altında ayrı bir depo**dur. Forgeprint marketplace'inde listelenir ama Forgeprint'in koduna bağımlı değildir. | Bağımsız sürüm döngüsü, iki dilin karışmaması. |

---

## 3. Depo yapısı (hedef)

```
cmd/forgelore/                 main paketi, CLI giriş noktası
internal/
  record/                 kayıt formatı, frontmatter ayrıştırıcı, ULID
  store/                  dosya deposu, kapsamlar, SQLite FTS5 indeksi
  fingerprint/            hata parmak izi çıkarma
  events/                 kanonik olay modeli
  inject/                 enjeksiyon kararları, bütçe
  ledger/                 enjeksiyon defteri, A/B atama
  usage/                  ajan kullanım verisi okuyucuları (adaptör)
  adapter/                eşleme motoru + gömülü eşleme dosyaları
  hookrun/                hook çalıştırıcı (fail-open, zaman aşımı)
  mcp/                    minimal MCP sunucusu
  redact/                 gizli bilgi maskeleme
adapters/                 ajan eşleme dosyaları (binary'ye gömülür)
testdata/                 ajan olay örnekleri, hata örnekleri
docs/
  adr/                    mimari karar kayıtları
  ilerleme.md             oturumlar arası ilerleme notu
vendor/
```

Kullanıcının projesindeki yapı:

```
.forgelore/
  records/<ulid>.md       ekip kayıtları (commit edilir)
  local/                  kişisel kayıtlar (gitignore)
  cache/index.db          türetilmiş indeks (gitignore)
  adapters/               isteğe bağlı yerel eşleme geçersiz kılmaları
  config.toml ya da benzeri  bütçe, A/B, kapsam ayarları
```

---

## 4. Fazlar

Her fazın yapısı aynıdır: **Amaç → Adımlar (`[SEN]` / `[CLAUDE]`) → Claude Code'un soracağı sorular → Kabul kriterleri.**

### Faz 0 — Depo ve yönetişim

**Amaç:** Boş ama doğru kurulmuş bir depo, kilitli kararların ADR olarak yazılması.

Adımlar:
1. `[SEN]` Ad belirlendi: **Forgelore**. Depoyu açmadan önce son kontrolleri yap: `npm view forgelore` (404 = boş), pkg.go.dev'de arama, GitHub'da `forgelore` adı, istenirse alan adı. Claude Code sonucu sorar; çakışma çıkarsa devam etmeden bildir.
2. `[SEN]` forgeprint organizasyonunda `forgelore` adıyla yeni depoyu aç, Apache-2.0 seç, ana dal korumasını ayarla.
3. `[SEN]` Makinene Go'nun güncel kararlı sürümünü kur. Claude Code sana sürüm kontrol komutunu verir.
4. `[CLAUDE]` Depo iskeleti: `go.mod` (modül yolu `github.com/forgeprint/forgelore`), klasör yapısı (bölüm 3), `LICENSE`, `DCO`, `CONTRIBUTING.md`, `SECURITY.md`, `.gitignore`, gitleaks yapılandırması.
5. `[CLAUDE]` `docs/adr/` altında K1–K15 için ADR'ler. Her biri kısa: bağlam, karar, sonuçlar.
6. `[CLAUDE]` Yerel derleme ve test betiği (tüm CI adımları yerelde çalışabilir olmalı; CI yalnızca aynı komutları çağırır). Çapraz derleme hedefleri: linux/darwin/windows × amd64/arm64.
7. `[CLAUDE]` `docs/ilerleme.md` oluştur.

Sorulacaklar: Desteklenecek asgari Go sürümü? Yayın betiği için ek bir araç (ör. GoReleaser) kullanılsın mı, yoksa düz betik mi? (Varsayılan: düz betik, K3'ün ruhuna uygun.)

Kabul kriterleri: Boş bir `main` tüm hedef platformlar için `CGO_ENABLED=0` ile derleniyor. ADR'ler yazılı. Yerel test komutu çalışıyor.

---

### Faz 1 — Kayıt formatı ve depolama çekirdeği

**Amaç:** Kayıtları yazıp okuyan, indeksleyen ve arayan çekirdek. Henüz hiçbir ajanı tanımıyor.

Adımlar:
1. `[CLAUDE]` Kayıt şemasını öner, kullanıcı onaylasın. En az: `schema`, `id` (ULID), `type` (`fix`, `dead_end`, `decision`, `command`, `note`), `scope` (`team`/`local`), `fingerprint` (varsa), `title`, `tags`, `created`, `source` (`user`, `hook`, `import`), `tainted` (bool). Gövde: markdown.
2. `[CLAUDE]` Frontmatter ayrıştırıcı. YAML'ın **kısıtlı bir alt kümesi** (düz anahtarlar, dizeler, sayılar, bool, dize listeleri), kendi yazımımız, katı doğrulamalı. Tanınmayan alanlar korunur (K11).
3. `[CLAUDE]` ULID üretimi (stdlib ile).
4. `[CLAUDE]` Dosya deposu: kayıt başına bir dosya, atomik yazma (geçici dosya + rename), iki kapsam.
5. `[CLAUDE]` SQLite FTS5 indeksi: dosyalardan tam yeniden üretim, değişen dosyalar için artımlı güncelleme (dosya zamanı + içerik hash'i). İndeks bozuksa sessizce yeniden üretilir.
6. `[CLAUDE]` Gizli bilgi maskeleme (`redact`): yaygın anahtar desenleri, `<private>` etiketi. Maskeleme yazmadan önce uygulanır.
7. `[CLAUDE]` Testler: gidiş-dönüş (yaz, oku, eşit mi), bilinmeyen alan korunması, bozuk dosya toleransı, indeks yeniden üretimi, maskeleme.

Sorulacaklar: Kayıt tipleri yeterli mi? Frontmatter için kısıtlı YAML mı, yoksa TOML ya da JSON mu? (Öneri: kısıtlı YAML, AGENTS.md/SKILL.md ekosistemiyle uyumlu.) Config dosyası formatı?

Kabul kriterleri: 10.000 kayıtlık sentetik bir depoda tam indeks üretimi ve arama süresi ölçülmüş ve rapora yazılmış. Tüm testler geçiyor.

---

### Faz 2 — Hata parmak izi ve CLI (evrensel taban)

**Amaç:** Hook ya da MCP olmadan, sadece shell ile tam kullanılabilir bir araç.

Adımlar:
1. `[CLAUDE]` Parmak izi algoritması: hata metnini normalleştir (mutlak yollar, satır/sütun numaraları, hex adresler, zaman damgaları, UUID'ler, geçici dosya adları, sayılar), komut türünü ekle, hash'le. Aynı hatanın farklı makinelerde aynı parmak izini üretmesi hedef.
2. `[SEN]` Kendi projelerinden (Go, .NET, TypeScript, Python) gerçek hata çıktıları topla ve `testdata/errors/` altına koy. Claude Code sana hangi formatta ve hangi hassas bilgileri silerek koyman gerektiğini söyler.
3. `[CLAUDE]` Parmak izi testleri: aynı hatanın varyantları aynı sonucu, farklı hatalar farklı sonucu vermeli. Yanlış eşleşme ve kaçırma oranları raporlanır.
4. `[CLAUDE]` CLI komutları: `init`, `record` (elle kayıt), `recall <hata metni>` (parmak izi ile eşleşme), `search <sorgu>` (kompakt dizin döndürür), `show <id>` (ayrıntı), `index rebuild`, `doctor`, `stats`.
5. `[CLAUDE]` Çıktılar hem insan okuyabilir hem `--json` ile makine okuyabilir. `search` her sonuç için yalnızca kimlik, tip, başlık döndürür (kademeli erişim).
6. `[CLAUDE]` Kullanıcı projesine eklenecek tek satırlık AGENTS.md yönlendirme metni (`forgelore init` bunu önerir, kendi başına yazmaz).

Sorulacaklar: CLI komut adları ve bayraklar uygun mu? `init` hangi dosyaları oluştursun?

Kabul kriterleri: Hiçbir ajan entegrasyonu olmadan, bir ajan sadece AGENTS.md yönlendirmesi ve shell ile hatayı sorgulayıp geçmiş çözümü bulabiliyor. Bu senaryo elle denenmiş ve raporlanmış.

---

### Faz 3 — Ölçüm altyapısı

**Amaç:** Hafıza özelliği geliştikçe her sürümün gerçekten tasarruf edip etmediğini gösteren kalite kapısı.

Adımlar:
1. `[CLAUDE]` Enjeksiyon defteri: her enjeksiyonun zamanı, oturumu, kaydı, bayt ve tahmini token miktarı. Tamamen yerel.
2. `[CLAUDE]` A/B atama: isteğe bağlı (varsayılan kapalı). Açıksa her oturum rastgele "enjeksiyon açık" ya da "kontrol" grubuna atanır. Kontrol grubunda hiçbir şey enjekte edilmez ama olaylar yine kaydedilir.
3. `[CLAUDE]` Kullanım okuyucu arayüzü: ajanın kendi yerel kayıtlarından oturum başına token kullanımını okur. **İlk uygulama Claude Code için.** Format güncel dokümantasyondan doğrulanır (kural 7).
4. `[CLAUDE]` `report` komutu: dönem boyunca enjekte edilen token, iki gruptaki oturum başına ortalama kullanım, tekrarlanan hata döngüsü sayısı. İstatistiksel olarak anlamlı olmayan farklar açıkça "anlamlı değil" olarak işaretlenir.
5. `[CLAUDE]` Dürüstlük sınırı: kullanım okuyucusu olmayan ajanlarda rapor yalnızca "harcanan" tarafı gösterir, kazanç iddia etmez.

Sorulacaklar: Hangi metrikler rapora girsin? A/B oranı (ör. %20 kontrol) ne olsun?

Kabul kriterleri: Sentetik verilerle rapor doğru hesaplıyor. Gerçek bir Claude Code oturumundan kullanım verisi okunabiliyor.

---

### Faz 4 — Claude Code adaptörü ve hook çalıştırıcı

**Amaç:** İlk tam destekli ajan.

Adımlar:
1. `[CLAUDE]` Kanonik olay modeli: SessionStart, PromptSubmit, PreTool, PostTool, Stop, PreCompact, SessionEnd. Alanlar: oturum kimliği, çalışma dizini, araç adı, girdi, çıktı, çıkış kodu, zaman.
2. `[CLAUDE]` Eşleme dosyası formatı (K5): olay adı eşlemesi, alan yolları, cevap formatı, zaman aşımı. Format kendi sürüm numarasını taşır.
3. `[CLAUDE]` Hook çalıştırıcı: stdin'den JSON okur, eşlemeyle kanonik olaya çevirir, çekirdeği çağırır, cevabı ajanın formatına çevirir. **K7:** her durumda başarıyla çıkar, katı zaman aşımı, panik yakalama.
4. `[SEN]` Claude Code'da hook'ları geçici olarak bir kayıt betiğine yönlendirip gerçek olay örnekleri topla. Claude Code sana adım adım nasıl yapılacağını ve hassas verileri nasıl temizleyeceğini söyler. Örnekler `testdata/agents/claude-code/<sürüm>/` altına girer.
5. `[CLAUDE]` Davranışlar:
   - SessionStart: bütçe sınırlı kompakt dizin (varsayılan bütçe config'den).
   - PostTool + başarısız çıkış: parmak izi → eşleşme varsa tek satırlık ipucu.
   - PostTool + önceki başarısızlıktan sonra başarı: çözüm **adayı** oluştur (yerel kapsam, onaysız).
   - Stop/SessionEnd: adayları özetle, kullanıcıya bir sonraki oturumda onay için sun ya da `review` komutuna bırak.
   - Dış içerik (web fetch vb.) içeren oturumdan gelen adaylar `tainted: true`.
6. `[CLAUDE]` Sözleşme testleri: toplanan gerçek örneklerle eşleme doğrulanır. Hangi Claude Code sürümüyle doğrulandığı kaydedilir.
7. `[CLAUDE]` Gecikme ölçümü: hook başına p50/p95 süre. Rapora yaz.
8. `[CLAUDE]` Claude Code plugin manifesti.

Sorulacaklar: Varsayılan enjeksiyon bütçesi? Adaylar ne zaman ve nasıl onaya sunulsun? Hook zaman aşımı kaç ms?

Kabul kriterleri: Gerçek bir Claude Code oturumunda bilinen bir hata tekrarlandığında ipucu enjekte ediliyor. Bozuk indeks, silinmiş klasör ve zaman aşımı senaryolarında Claude Code hiç etkilenmiyor. Gecikme ölçülmüş.

---

### Faz 5 — MCP sunucusu (kendi yazımımız)

**Amaç:** MCP destekleyen her istemciden aranabilir hafıza.

Adımlar:
1. `[CLAUDE]` Hedef MCP spesifikasyon sürümünü belirle (güncel spesifikasyondan doğrula) ve ADR'ye yaz.
2. `[CLAUDE]` stdio JSON-RPC: `initialize` ve sürüm müzakeresi, `tools/list`, `tools/call`, hata kodları. Yalnızca ihtiyaç duyulan kısım.
3. `[CLAUDE]` Araçlar (kademeli erişim araç tasarımıyla zorlanır): `search` (kompakt dizin), `get` (yalnızca `search`'ten gelen kimliklerle ayrıntı), `recall_error` (parmak izi eşleşmesi), `propose` (kayıt önerisi, onaysız).
4. `[CLAUDE]` Araç açıklamaları kısa, en önemli bilgi başta (istemciler uzun açıklamaları kesebilir).
5. `[CLAUDE]` Uyum testleri: spesifikasyondaki örnek mesajlarla.

Sorulacaklar: Araç seti yeterli mi? `propose` kayıtları hangi kapsamda başlasın?

Kabul kriterleri: Claude Code ve en az bir başka MCP istemcisi bağlanıp arama yapabiliyor.

---

### Faz 6 — Ekip akışı

**Amaç:** Yerel adayların güvenli şekilde ekip hafızasına geçmesi.

Adımlar:
1. `[CLAUDE]` `review` komutu: bekleyen adayları listeler, kabul/red/düzenle.
2. `[CLAUDE]` `promote` komutu: yerel kaydı ekip kapsamına taşır. `tainted` kayıtlar ek onay ister.
3. `[CLAUDE]` Commit öncesi kontrol: gitleaks kuralları + kendi maskeleme kontrolümüz ekip kayıtlarına uygulanır.
4. `[CLAUDE]` Tekrar tespiti: aynı parmak izine sahip kayıtlar birleştirme için önerilir.
5. `[CLAUDE]` Fayda istatistikleri **yerel kalır**, ekip kayıtlarına kişisel veri yazılmaz.
6. `[SEN]` Küçük bir ekip denemesi: en az iki kişi, bir hafta. Claude Code sana deneme protokolünü yazar.

Kabul kriterleri: İki geliştirici aynı depoda eşzamanlı kayıt eklediğinde merge çakışması oluşmuyor. Sır içeren bir kayıt commit edilemiyor.

---

### Faz 7 — Diğer ajanlar

**Amaç:** "Herkesi kapsama" hedefi.

Her ajan için aynı döngü:
1. `[CLAUDE]` Güncel resmî dokümantasyondan hook desteğini araştır: olay adları, girdi/çıktı formatı, yapılandırma dosyası yolu, güven/onay mekanizması, eşleştirici kuralları. Kaynakları rapora yaz.
2. `[CLAUDE]` Seviye belirle: **A** (hook + MCP + CLI), **B** (MCP + CLI), **C** (yalnızca CLI + AGENTS.md).
3. `[SEN]` O ajanı kur, Faz 4'teki yöntemle gerçek olay örneklerini topla.
4. `[CLAUDE]` Eşleme dosyası + sözleşme testleri + (varsa) kullanım okuyucusu.
5. `[CLAUDE]` Kurulum talimatı ve gerekiyorsa plugin/marketplace manifesti.

Önerilen sıra: Codex CLI, Gemini CLI, GitHub Copilot CLI, Cursor, ardından diğerleri. Sırayı kullanıcıyla teyit et.

Kabul kriterleri: Her ajan için desteklenen seviye ve doğrulanan sürüm bir uyumluluk tablosunda (`docs/uyumluluk.md`) yayımlanmış.

---

### Faz 8 — Dağıtım ve yayın

Adımlar:
1. `[CLAUDE]` Yayın betiği: tüm platformlar için binary, SHA-256 checksum dosyası.
2. `[CLAUDE]` Kurulum betiği (checksum doğrulamalı).
3. `[CLAUDE]` İsteğe bağlı npm sarmalayıcı: yalnızca doğru binary'yi indirip checksum'ını doğrular, başka hiçbir şey yapmaz.
4. `[CLAUDE]` Sürümleme politikası: semver. Kayıt şeması ve eşleme formatı sürümlerinin binary sürümünden bağımsız olduğu ve nasıl yükseltileceği `docs/surumleme.md`'de.
5. `[SEN]` İlk sürümü yayımla (GitHub Release). İmzalama (ör. cosign) istenirse karar ver.
6. `[SEN]` **Forgeprint deposunda ayrı bir görev olarak:** marketplace'e bu projeyi ekle ve README'ye bağlantı ver. Bu iş bu depoda değil, Forgeprint'in kendi Claude Code oturumunda yapılır.

Kabul kriterleri: Temiz bir makinede tek komutla kurulum çalışıyor, checksum doğrulanıyor.

---

### Faz 9 — Kayma dedektörü ve bakım

**Amaç:** Ajanlar güncellendiğinde bunu kullanıcılardan önce görmek.

Adımlar:
1. `[CLAUDE]` Zamanlanmış bir iş: desteklenen ajanların son sürümlerini kurar ve toplanan olay örnekleriyle sözleşme testlerini çalıştırır. Kırılma olursa issue açar. (Bu iş CI'da çalışır ama yerelde de tek komutla çalıştırılabilir olmalı.)
2. `[CLAUDE]` Eşleme dosyası güncellemesi için katkı rehberi: bir ajan değiştiğinde topluluk üyesinin kod yazmadan eşlemeyi nasıl güncelleyebileceği.
3. `[CLAUDE]` `doctor` komutu kullanıcının kurulu ajan sürümünü desteklenen sürümlerle karşılaştırıp uyarır.

Kabul kriterleri: Bir eşleme alanı bilerek bozulduğunda kayma dedektörü bunu yakalıyor.

---

## 5. Kullanıcıya ait işlerin özeti

| Faz | Senin yapacakların |
|---|---|
| 0 | Ad seçimi, depo açma, dal koruması, Go kurulumu |
| 2 | Gerçek hata örnekleri toplamak |
| 4 | Claude Code olay örneklerini toplamak |
| 6 | Ekip denemesi |
| 7 | Her ajanı kurup olay örneklerini toplamak |
| 8 | Yayın, imzalama kararı, Forgeprint marketplace kaydı |

---

## 6. Bu dokümanın bakımı

Bu doküman proje boyunca kilitli kararların kaynağıdır. Bir kilitli karar değişirse önce bu dosya ve ilgili ADR güncellenir, sonra kod. Claude Code bu dosyayı değiştirmeden önce kullanıcıya sorar.
