# Demonstracija X-Forwarded-For Ranjivosti u Gin Framework-u

## Uvod

Demonstracija sigurnosne ranjivosti koja nastaje kada web aplikacije napisane u Go Gin framework-u ne konfigurišu pravilno funkciju `SetTrustedProxies()`. Projekat pokazuje kako napadač može da iskoristi HTTP headere da zaobiđe sigurnosne mehanizme ili izazove Denial of Service (DoS) napad. Link do video demontracije -> **[Demo](https://youtu.be/2_nkGjmMd1U)**

## Pretnja: Manipulacija X-Forwarded-For Headera

### Kontekst

HTTP header `X-Forwarded-For` se koristi kada je aplikacija iza reverse proxy-ja (Nginx, HAProxy). Proxy dodaje ovaj header sa originalnom IP adresom klijenta:

```
Klijent (192.168.1.100) → Proxy (203.0.113.50) → Aplikacija

HTTP zahtev:
X-Forwarded-For: 192.168.1.100    ← Proxy dodao
```

**Problem**: Napadač može sam da postavi ovaj header:

```bash
curl -H "X-Forwarded-For: 10.0.0.1" http://example.com/api
```

Ako aplikacija ne verifikuje da zahtev zaista dolazi od pouzdanog proxy-ja (`SetTrustedProxies()`), napadač može da lažira bilo koju IP adresu. 

Bez pravilno konfigurišanog `SetTrustedProxies()`, aplikacija će koristiti IP adresu iz `X-Forwarded-For` headera umesto stvarne IP adrese odakle zahtev zapravo dolazi (TCP konekcija). Ako je rate limiting implementiran unutar aplikacije i zasnovan na IP adresi, napadač može da lažira različite IP adrese i tako zaobiđe ograničenja.

### Tipovi napada

**Scenario 1: Rate limiter bypass**
- Aplikacija koristi `X-Forwarded-For` za rate limiting
- Napadač lažira različite IP adrese → svaka ima svoj rate limit → bypass

**Scenario 2: DoS legitimnih korisnika**
- Aplikacija ima IP whitelisting (dozvoljava samo proxy IP)
- Napadač lažira proxy IP → zaobilazi whitelist
- Moguće posledice: 
  - Napadač exhaustuje rate limit proxy IP-a (ako je rate limit implementiran na web serveru) → svi legitimni korisnici blokirani
  - Napadač samo ima pristup web serveru bez prolaska kroz proxy (naravno ako zna adresu servera)

**Scenario 3: Sigurna implementacija**
- `SetTrustedProxies()` pravilno konfigurisan → ignoriše lažne headere
- Napadi ne uspevaju

## Naš Sistem: Subscriptions & Ticket Service

### Opis aplikacije

Kreiran je jednostavan web servis za upravljanje pretplatama i ticketima koji bi mogao da se koristi za sistem prodaje karata, članarina ili subscription paketa. Aplikacija pruža REST API sa sledećim funkcionalnostima:

Resursi:
- Subscriptions (Pretplate) - CRUD operacije za pretplate korisnika
- Tickets (Tiketi) - CRUD operacije za tikete

Endpointi:
- `GET /api/subscriptions/` - Lista svih pretplata
- `POST /api/subscriptions/` - Kreiranje nove pretplate
- `PUT /api/subscriptions/` - Izmena pretplate
- `DELETE /api/subscriptions/{id}` - Brisanje pretplate
- 
- `GET /api/tickets/` - Lista svih tiketa
- `POST /api/tickets/` - Kreiranje novog tiketa
- `PUT /api/tickets/` - Izmena tiketa
- `DELETE /api/tickets/{id}` - Brisanje tiketa

### Važnost Subscriptions & ticket servisa

Sistem za upravljanje pretplatama i ticketima je kritična komponenta aplikacije za prevoz. Servis čuva informacije o:
- Kupljenim kartama i pretplatama
- Istoriji plaćanja korisnika
- Važećim pretplatama i njihovim datumima

**Zašto je dostupnost ovog servisa kritična:**

Ako servis nije dostupan, korisnici ne mogu:
- Proveriti važeće pretplate prilikom korišćenja usluge (npr. ulazak u prevoz)
- Dokazati da su platili uslugu
- Pristupiti elektronskim kartama/ticketima

**Biznis uticaj DoS napada:**

Pad ovog servisa direktno utiče na krajnje korisnike koji su već platili uslugu ali je ne mogu koristiti. Ovo dovodi do:
- Nezadovoljstva korisnika i gubitka poverenja
- Finansijskih gubitaka (refundiranje, izgubljeni prihodi)
- Reputacionih šteta za kompaniju

Zbog kritičnosti servisa, zaštita od DoS napada je obavezna.

## Problem: Nevalidovani c.ClientIP() u Gin Framework-u

### Kako funkcioniše ClientIP() u Gin-u

Gin framework koristi metodu `c.ClientIP()` da utvrdi IP adresu klijenta. Ova metoda čita različite HTTP headere sledećim redosledom:

1. `X-Forwarded-For` (ako je proxy trusted)
2. `X-Real-IP` (ako je proxy trusted)
3. `RemoteAddr` (stvarna TCP konekcija)

Problem: Ako `SetTrustedProxies()` nije konfigurisan, Gin veruje bilo kom `X-Forwarded-For` headeru koji primi, bez provere da li zahtev zaista dolazi od pouzdanog proxy-ja.

### Ranjiva aplikacija

U prvom i drugom scenariju namerno nismo koristili `SetTrustedProxies()`:

```go
router := gin.Default()
// PROBLEM: Ovde nedostaje SetTrustedProxies()!

// Rate limiter middleware koristi c.ClientIP()
router.Use(RateLimiterMiddleware())

// ClientIP() veruje X-Forwarded-For headeru bez provere
```

Naš rate limiter middleware računa broj zahteva po IP adresi.

## Tri Scenarija Demonstracije

Projekat implementira tri odvojena scenarija:

### Scenario 1: Zaobilaženje Rate Limiting-a

Implementacija: `main.go`

Rate limit: 10 zahteva po minuti po IP adresi

Napad:
Napadač šalje zahteve sa različitim lažnim IP adresama u `X-Forwarded-For` headeru. Rate limiter vidi svaki zahtev kao da dolazi od različitog korisnika, pa svaka lažna IP adresa ima svoj separatni limit.

```bash
curl -H "X-Forwarded-For: 10.0.0.1" http://localhost:8080/api/subscriptions/  # 1/10 za IP 10.0.0.1
curl -H "X-Forwarded-For: 10.0.0.2" http://localhost:8080/api/subscriptions/  # 1/10 za IP 10.0.0.2
curl -H "X-Forwarded-For: 10.0.0.3" http://localhost:8080/api/subscriptions/  # 1/10 za IP 10.0.0.3
# Napadač može da šalje neograničen broj zahteva
```

Posledica:
Rate limiting je potpuno zaobiđen. Napadač može da šalje neograničeno zahteva u minuti umesto maksimalno 10. Ovo može izazvati pad web servera.

---

### Scenario 2: Denial of Service Legitimnih Korisnika

Implementacija: `main_proxy_dos.go`

Dodatni mehanizam: IP Whitelisting - dozvoljava samo zahteve sa proxy IP `203.0.113.50`

Arhitektura:
```
Legitimni korisnici → Reverse Proxy (203.0.113.50) → Aplikacija
Napadač (direktno) → Aplikacija (blokiran whitelistom, ALI...)
```

Ranjivost:
Aplikacija koristi `c.ClientIP()` i za whitelisting proveru:

```go
func IPWhitelistMiddleware() gin.HandlerFunc {
    allowedIP := "203.0.113.50"  // Samo proxy sme da pristupa
    
    return func(c *gin.Context) {
        clientIP := c.ClientIP()  // Ranjivo!
        
        if clientIP != allowedIP {
            c.JSON(403, gin.H{"error": "Access denied"})
            c.Abort()
            return
        }
    }
}
```

Napad:
1. Napadač lažira `X-Forwarded-For: 203.0.113.50` (pretvaraju se da su proxy)
2. Whitelisting ga propušta jer veruje lažnom headeru
3. Napadač šalje 10+ zahteva brzo
4. Rate limiter vidi sve zahteve kao da dolaze sa IP `203.0.113.50` (proxy IP)
5. Nakon 10 zahteva, rate limiter blokira IP `203.0.113.50`
6. SVI legitimni korisnici kroz pravi proxy su blokirani

```bash
# Napadač exhaustuje rate limit proxy IP-a
for i in {1..15}; do
  curl -H "X-Forwarded-For: 203.0.113.50" http://localhost:8080/api/subscriptions/
done

# Legitimni korisnik pokušava pristup kroz pravi proxy
curl http://localhost:8080/api/subscriptions/
# HTTP 429 Too Many Requests - BLOKIRAN!
```

Posledica:
Potpun Denial of Service. Svi legitimni korisnici ne mogu da pristupe aplikaciji.

---

### Scenario 3: Sigurna Implementacija

Implementacija: `main_secure.go`

Rešenje: Korišćenje `SetTrustedProxies()` funkcije

```go
router := gin.Default()

// REŠENJE: Ne veruj nijednom proxy serveru
router.SetTrustedProxies(nil)

// Sada c.ClientIP() ignoriše X-Forwarded-For i koristi samo RemoteAddr
router.Use(RateLimiterMiddleware())
```

Efekat:
Gin framework ignoriše sve proxy headere (`X-Forwarded-For`, `X-Real-IP`) i koristi samo stvarnu IP adresu iz TCP konekcije (`RemoteAddr`). Ovo je IP adresa sa koje je zahtev zaista stigao.

**Rezultat napada:**
```bash
# Napadač pokušava da lažira IP
curl -H "X-Forwarded-For: 10.0.0.1" http://localhost:8080/api/subscriptions/  # Request 1/10
curl -H "X-Forwarded-For: 10.0.0.2" http://localhost:8080/api/subscriptions/  # Request 2/10
# ...
curl -H "X-Forwarded-For: 10.0.0.11" http://localhost:8080/api/subscriptions/  # HTTP 429

# Server vidi SVE zahteve kao da dolaze sa iste IP adrese (127.0.0.1)
# Rate limiter blokira samo napadača
# Legitimni korisnici nisu pogođeni
```

Napomena:
Ako je aplikacija stvarno iza proxy servera u produkciji, koristiti:
```go
router.SetTrustedProxies([]string{"203.0.113.50"})  // IP tvog proxy servera
```


## Struktura projekta

```
Gin_Attack/
├── main.go                      # Scenario 1: Rate limiter bypass
├── main_proxy_dos.go            # Scenario 2: DoS legitimnih korisnika
├── main_secure.go               # Scenario 3: Sigurna implementacija
├── db/
│   └── database.go              # SQLite database setup
├── models/
│   └── models.go                # Subscription & Ticket modeli
├── middlewares/
│   ├── rate_limiter.go          # IP-based rate limiting
│   ├── ip_whitelist.go          # Ranjiv IP whitelist (Scenario 2)
│   └── secure_ip_whitelist.go   # Siguran IP whitelist (Scenario 3)
├── handlers/
│   ├── subscription_handlers.go # API handlers za subscriptions
│   └── ticket_handlers.go       # API handlers za tickets
├── test_scripts/
│   ├── dos_attack.py            # Test za Scenario 1
|   ├── dos_attack_agressive.py  # Test za Scenario 1, veliko opterećenje servera
|   ├── normal_test.py           # Test za normalno funkcionisanje Scenarija 1
│   ├── proxy_dos_attack.py      # Test za Scenario 2
│   ├── legitimate_user_test.py  # Test legitimnog korisnika
│   └── test_secure_server.py    # Test za Scenario 3, sve bezbedno
```

## Pokretanje i Testiranje

### Opcije pokretanja

Aplikacija se može pokrenuti na dva načina:

**1. Lokalno (bez Docker-a)**
- Svi scenariji koriste port **8080**
- Koristi SQLite bazu podataka
- Potrebno je imati Go instaliran (1.23+)

**2. Docker (preporučeno)**
- Scenario 1: port **8081**
- Scenario 2: port **8082**
- Scenario 3: port **8083**
- Koristi PostgreSQL bazu podataka
- Svi scenariji rade istovremeno

**VAŽNO Prilagođavanje test skripti:**
U Python test skriptama prilagoditi `BASE_URL` prema načinu pokretanja:
- Lokalno: `BASE_URL = "http://localhost:8080"`
- Docker: `BASE_URL = "http://localhost:8081"` (ili 8082/8083)

---

### Docker pokretanje

```powershell
# Build i pokretanje svih servisa (PostgreSQL + sva 3 scenarija)
docker-compose up -d

# Provera statusa
docker-compose ps

# Zaustavljanje
docker-compose down
```

Svi scenariji su odmah dostupni:
- Scenario 1: http://localhost:8081
- Scenario 2: http://localhost:8082
- Scenario 3: http://localhost:8083

---

### Instalacija zavisnosti (za lokalno pokretanje)

```powershell
# Go zavisnosti
go mod tidy

# Python zavisnosti (za test skripte)
cd test_scripts
pip install -r test_scripts/requirements.txt
```

### Test Scenario 1: Rate Limiter Bypass

Kompajliranje i pokretanje aplikacije:
```powershell
go run main.go
```
Ili preko docker-a (`http://localhost:8081`). Prethodno objašnjeno.

Aplikacija će biti dostupna na `http://localhost:8080`

Manuelni test:
```bash
# Normalan zahtev - server vidi pravu IP adresu
curl http://localhost:8080/api/subscriptions/

# Napad - lažiranje različitih IP adresa
curl -H "X-Forwarded-For: 10.0.0.1" http://localhost:8080/api/subscriptions/
curl -H "X-Forwarded-For: 10.0.0.2" http://localhost:8080/api/subscriptions/
curl -H "X-Forwarded-For: 10.0.0.3" http://localhost:8080/api/subscriptions/
# Rate limiter se zaobilazi - svaka "fake IP" ima svoj limit
```

Automatski test:
```powershell
cd test_scripts
python dos_attack.py

python dos_attack_agressive.py
```

Skripta će poslati 1000 zahteva sa različitim lažnim IP adresama. Očekivano: svih 1000 zahteva uspešno (HTTP 200). Agressive skripta salje 10000 zahteva, naravno ovi parametri se mogu podesiti u samim skriptama.

---

### Test Scenario 2: DoS Legitimnih Korisnika (lokalno pokretanje)

Kompajliranje i pokretanje aplikacije:
```powershell
go run main_proxy_dos.go
```
Ili preko docker-a (`http://localhost:8082`). Prethodno objašnjeno.

Manuelni test:
```bash
# Pokušaj direktnog pristupa - odbijen (nije sa proxy IP-a)
curl http://localhost:8080/api/subscriptions/
# HTTP 403 Forbidden

# Napadač lažira proxy IP
curl -H "X-Forwarded-For: 203.0.113.50" http://localhost:8080/api/subscriptions/
# HTTP 200 OK (whitelisting prevaren)

# Ponovi 10+ puta da exhaustuješ rate limit
for i in {1..15}; do
  curl -H "X-Forwarded-For: 203.0.113.50" http://localhost:8080/api/subscriptions/
done

# Sada čak i sa "pravim" proxy IP-jem:
curl -H "X-Forwarded-For: 203.0.113.50" http://localhost:8080/api/subscriptions/
# HTTP 429 Too Many Requests - SVI korisnici blokirani!
```

Automatski test:
```powershell
cd test_scripts

# Simulacija napada
python proxy_dos_attack.py

# Provera da li legitimni korisnik može pristupiti
python legitimate_user_test.py
# Očekivano: HTTP 429 (blokiran)
```

---

### Test Scenario 3: Sigurna Implementacija (lokalno pokretanje)

Kompajliranje i pokretanje aplikacije:
```powershell
go run main_secure.go
```
Ili preko docker-a (`http://localhost:8083`). Prethodno objašnjeno.

Manuelni test:
```bash
# Pokušaj napada sa lažnim IP adresama
curl -H "X-Forwarded-For: 10.0.0.1" http://localhost:8080/api/subscriptions/  # 1/10
curl -H "X-Forwarded-For: 10.0.0.2" http://localhost:8080/api/subscriptions/  # 2/10
curl -H "X-Forwarded-For: 10.0.0.3" http://localhost:8080/api/subscriptions/  # 3/10
# ...
curl -H "X-Forwarded-For: 10.0.0.11" http://localhost:8080/api/subscriptions/  # HTTP 429

# Server ignoriše lažne headere i vidi samo stvarnu IP adresu (127.0.0.1)
# Napad ne uspeva!
```

Automatski test:
```powershell
cd test_scripts
python test_secure_server.py
```

Skripta će pokušati da pošalje 50 zahteva sa lažnim IP adresama. Očekivano: prvih 10 uspešno, ostali blokirani (HTTP 429).

## Poređenje Rezultata

| Aspekt | Scenario 1 | Scenario 2 | Scenario 3 |
|--------|-----------|-----------|-----------|
| **SetTrustedProxies()** | Nije korišćen | Nije korišćen | Korišćen (nil) |
| **Rate limiter funkcioniše** | NE | DA, ali pogrešno | DA, pravilno |
| **50 zahteva sa lažnim IP-ovima** | Svih 50 uspešnih | Svih 50 uspešnih | 10 uspešnih, 40 blokirano |
| **ClientIP() veruje X-Forwarded-For** | DA | DA | NE |
| **Napad uspešan** | DA | DA | NE |
| **DoS legitimnih korisnika** | NE | DA | NE |


## Tehnički Detalji

### Kako SetTrustedProxies() rešava problem

**Bez SetTrustedProxies():**
```go
router := gin.Default()
// ClientIP() slepo veruje X-Forwarded-For headeru
```

Gin framework čita `X-Forwarded-For` header bez provere da li zahtev dolazi sa pouzdane adrese.

**Sa SetTrustedProxies(nil):**
```go
router := gin.Default()
router.SetTrustedProxies(nil)
// ClientIP() ignoriše sve proxy headere
```

Gin framework ignoriše `X-Forwarded-For`, `X-Real-IP` i druge proxy headere. Koristi samo `RemoteAddr` (stvarnu TCP IP adresu).

**Sa SetTrustedProxies([proxy-ip]):**
```go
router := gin.Default()
router.SetTrustedProxies([]string{"203.0.113.50"})
// ClientIP() veruje X-Forwarded-For samo ako dolazi od 203.0.113.50
```

Gin framework veruje `X-Forwarded-For` headeru SAMO ako zahtev dolazi od navedene proxy IP adrese. Ovo je pravi način za produkciju kada je aplikacija iza reverse proxy-ja.

