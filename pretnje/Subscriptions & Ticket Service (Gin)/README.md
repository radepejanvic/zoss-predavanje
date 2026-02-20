# Pretnje, napadi i mitigacije na Subscriptions & Ticket Service (Gin)
## Uvod

Ovaj dokument analizira sigurnosne pretnje servisa za pretplate i tikete. Fokus je na ranjivostima koje proizilaze iz načina na koji Gin rukuje HTTP zahtevima, concurrency-jem i integracijom sa drugim servisima.

## Kontekst Sistema

Subscription & Ticket Management sistem omogućava:
- Kupovinu i upravljanje pretplatama
- Prodaju tiketa
- Real-time validaciju pretplata i tiketa
- Payment processing

**Kritični resursi:**
- Subscription records (active/expired status)
- Single-use ticket identifikatori
- Revenue i payment integritet
- Dostupnost servisa

---

## Katalog Napada

### 1. X-Forwarded-For Header Manipulation

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Rate limiting mehanizmi
- IP-based access control
- Dostupnost servisa


Gin framework koristi `ClientIP()` metodu za ekstraktovanje IP adrese klijenta. Problem je što Gin po default-u veruje proxy header-ima (`X-Forwarded-For`, `X-Real-IP`) koje klijent pošalje. Pošto HTTP headeri mogu biti lažirani, napadač može manipulisati svojom IP adresom i zaobići sigurnosne mehanizme.

Napadač mora imati mrežni pristup serveru i server mora koristiti Gin ClientIP() metodu bez pravilno konfigurisanih trusted proxy adresa.

#### Napadi

**Rate Limiter Bypass:** Napadač šalje zahteve sa različitim lažnim X-Forwarded-For header-ima. Svaki zahtev izgleda kao da dolazi sa različite IP adrese, čime se zaobilazi rate limiting (npr. 10 zahteva/min postaje 1000+ zahteva/min).

**Denial of Service:** Napadač exhaustuje rate limit proxy IP adrese. Kada se proxy IP blokira, **svi legitimni korisnici** koji dolaze kroz taj proxy gube pristup servisu.

#### Mitigacije

**SetTrustedProxies() Konfiguracija:** Bez proxy-ja koristiti `SetTrustedProxies(nil)`, za aplikacije iza proxy-ja specificirati trusted proxy IP adrese.

Praktična demonstracija:

Za detaljan prikaz napada, implementaciju i testiranje pogledati:  
📍 **[GitHub - X-Forwarded-For Attack Demo](https://github.com/radepejanvic/zoss-predavanje)**

Demonstracija sadrži:
- Tri scenarija
- Automatske test skripte
- Docker environment
- Kompletnu dokumentaciju napada i odbrane

---

### 2. Race Condition: Concurrent Ticket Validation

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Single-use ticket integrity  
- Prihod (jedan tiket = jedna upotreba)
- Database consistency


Gin je dizajniran za visoke performanse i obrađuje svaki HTTP zahtev u odvojenoj goroutine-i. Kada dva zahteva stignu gotovo istovremeno za validaciju istog ticketa, izvršavaju se paralelno u različitim goroutine-ima.

Ako handler nije pažljivo implementiran sa database locking mehanizmima, javlja se race condition - oba zahteva pročitaju tiket status prije nego što bilo koji update-uje, rezultirajući u duploj validaciji.

#### Napadi

**Scenario: Dupla Validacija Single-Use Ticketa**

Korisnik kupi ticket za $10 sa unique ID i statusom `used: false` u bazi.

Napadač istovremeno pošalje dva zahteva:
```http
Tab 1: POST /api/tickets/validate?id=TICKET123
Tab 2: POST /api/tickets/validate?id=TICKET123
```

**Ranjiva handler logika:**
1. Pročitaj ticket iz baze
2. Proveri da li je `used == false`
3. Nastavi ako je false
4. Postavi `used = true`
5. Sačuvaj u bazu

**Execution timeline:**
```
T0: Goroutine 1 čita ticket → used = false ✓
T1: Goroutine 2 čita ticket → used = false ✓  (Još nije update-ovan!)
T2: Goroutine 1 → used = true, save()
T3: Goroutine 2 → used = true, save()
```

Oba zahteva pročitala `used = false` pre nego što je prvi update-ovao. Oba dobiju "Valid ticket" response.

Šta napadač postiže:
- Jedan ticket validiran dvaput
- Korisnik platio $10, dobio dva ulaza
- Direktan gubitak prihoda
- Ako se automatizuje (script sa 10 simultanih zahteva), isti tiket se može iskoristiti višestruko

#### Mitigacije

**Database Transaction sa Row Locking:**

Korišćenjem `SELECT FOR UPDATE` u database transakciji, red u bazi se zaključava tokom čitanja i update-a. Drugi zahtev koji pokuša da čita isti ticket mora čekati dok prvi ne završi. Kada dobije pristup, ticket će već biti označen kao `used = true` i validacija će failovati.

**Optimistic Locking sa Version Counter:**

Dodavanje `version` polja u Ticket model koje se inkrementira pri svakom update-u. Update query uključuje proveru verzije: `WHERE id = ? AND version = ?`. Samo prvi zahtev uspešno update-uje verziju jer će drugi naići na promenjenu verziju i failovati sa `RowsAffected = 0`.

**Idempotency Key Pattern:**

Klijent generiše jedinstveni validation key i šalje ga sa svakim zahtevom. Server čuva key i ignoriše duplicate zahteve sa istim key-om, čime se osigurava da se validacija izvršava samo jednom bez obzira na broj paralelnih zahteva.

---

### 3. Insecure Parameter Binding (IDOR)

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Tuđe aktivne pretplate
- Istorija plaćanja drugih korisnika  
- Tiketi koji pripadaju drugim korisnicima


Gin omogućava ekstremno lak pristup URL parametrima kroz `c.Param()` i query parametrima kroz `c.Query()`. Developeri često direktno koriste ove parametre za database query bez provere vlasništva.

Gin olakšava routing ali ne pruža built-in authorization layer. Rezultat: IDOR (Insecure Direct Object Reference) ranjivosti su česte u Gin aplikacijama.

#### Napadi

**Scenario 1: Pristup Tuđoj Pretplati**

Korisnik vidi svoju pretplatu na:
```
GET /api/subscriptions/45678
```

Primećuje numerički ID u URL-u. Pokušava da menja: `45679`, `45680`, `45681`...

Ranjiv handler direktno koristi `c.Param("id")` za database query **bez provere** da li korisnik ima pravo pristupa tom resursu.

Što napadač postiže:
- Vidi sve pretplate u sistemu (privacy breach)
- Identifikuje korisnike sa premium pretplatama (targetiranje za phishing)
- Payment method details, renewal dates

**Scenario 2: Bruteforce Tuđih Tiketa**

```http
POST /api/tickets/validate
{"ticket_id": "TKT-99999"}
```

Napadač automatski probava ID-ove:
```
TKT-10000, TKT-10001, TKT-10002...
```

Čim pogodi validan neiskorišćen tiket, može ga koristiti iako ga nije kupio. Besplatan pristup servisu.

**Scenario 3: Brisanje Tuđih Pretplata (DoS)**

```
DELETE /api/subscriptions/12345
```

Ako handler ne proverava vlasništvo, napadač briše tuđe pretplate → Denial of Service.

#### Mitigacije

**Authorization Check u Handler-u:**

Ključna odbrana je dodavanje ownership provere u database query. Handler ekstraktuje user ID iz JWT tokena (middleware-om) i query mora uključiti: `WHERE id = ? AND user_id = ?`. Time se osigurava da korisnik može pristupiti samo svojim resursima, čak i ako pogodi tuđe ID-ove.

**Alternativni API Design (eliminacija IDOR-a):**

```
Ranjivo:  GET /api/subscriptions/12345
Sigurno:  GET /api/users/me/subscriptions
```

Server automatski filtrira po authenticated user-u iz JWT-a. Nema ID-a u URL-u koji se može manipulisati.

**UUID umesto Sequential ID-ova:**

```
GET /api/subscriptions/f47ac10b-58cc-4372-a567-0e02b2c3d479
```

UUID-ovi su teži za pogađanje (128-bit random), ali nisu zamena za authorization check - samo dodatni sloj security through obscurity.

---

### 4. Payment Amount Manipulation (Price Tampering)

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Prihod
- Payment integrity
- Business logic bypass


Gin's `c.Bind()` i `c.ShouldBindJSON()` metode automatski deserializuju request body u Go struct bez ikakve validacije podataka. Developeri često direktno koriste bind-ovane vrednosti za kreiranje payment-a sa Stripe/PayPal API-jem. 

Ako frontend šalje iznos plaćanja u request-u i Gin handler ga direktno koristi bez server-side validacije, napadač može manipulisati cenu proizvoda/pretplate pre slanja zahteva.

#### Napadi

**Scenario: Price Tampering u Payment Intent**

**Normalan flow:**
1. Korisnik odabere Premium plan ($99/mesec) na frontend-u
2. Frontend šalje: `POST /api/payment/create {plan_id: "premium", amount: 9900}` (u cents)
3. Gin handler bind-uje request
4. Handler kreira Stripe PaymentIntent sa tim amount-om
5. Korisnik plaća $99

**Attack flow:**

Napadač otvara DevTools → Network tab i presreće request pre slanja:

```json
POST /api/payment/create
{
  "plan_id": "premium",
  "amount": 100    // Changed from 9900 to 100
}
```

Ranjivi Gin handler direktno koristi `request.Amount` iz bind-ovanog struct-a za Stripe API poziv. Stripe kreira PaymentIntent za $1 umesto $99.

Šta napadač postiže:
- Premium pretplatu ($99/mesec) za $1
- Direktan gubitak prihoda: $98/mesec po napadaču
- Kompletno zaobilaženje pricing logike

**Varijacija: Plan Switching**

Napadač menja `plan_id` ali ostavlja stari niži `amount`:
```json
{
  "plan_id": "enterprise",     // $299/mesec
  "amount": 999                // $9.99 (basic plan cena)
}
```

Ako handler samo validira da li plan postoji, ali ne validira da li amount odgovara planu, Enterprise pretplata se dobija za Basic plan cenu.

#### Mitigacije

**Server-Side Price Validation:**

Handler **nikada** ne sme verovati amount-u koji dolazi od frontend-a. Pravilna implementacija:
1. Frontend šalje samo `plan_id`
2. Handler izvlači cenu iz sopstvene baze: `SELECT price FROM plans WHERE id = ?`
3. Handler koristi tu cenu za Stripe PaymentIntent kreiranje
4. Amount se generiše server-side, ne prima se od klijenta

**Integrity Check:**

Ako je tehnički neophodna da frontend šalje amount (npr. za custom donation), koristiti HMAC signature:
- Backend generiše `signature = HMAC(plan_id + amount + secret_key)`
- Frontend šalje amount i signature
- Backend verifikuje signature pre procesiranja
- Sprečava se modifikacija amount-a jer napadač ne zna secret_key

**Stripe Price Objects:**

Umesto dinamičkih amount-ova, koristiti Stripe Price objekta sa fiksnim cenama. Handler šalje samo `price_id` Stripe-u - amount je već definisan na Stripe dashboard-u i ne može se manipulisati od strane klijenta.

---

### 5. Idempotency Key Mishandling

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Payment integrity (dupla naplaćivanja)
- Subscription consistency (duplicate records)
- Prihod i poverenje korisnika


Gin rutira svaki HTTP zahtev u zasebnu goroutine za maksimalne performanse. Kada korisnik klikne "Pay" dugme i network timeout se desi, browser automatski retry-uje zahtev. Bez idempotency zaštite, oba zahteva će se izvršiti paralelno u različitim goroutine-ima.

Stripe (i drugi payment gateway-i) zahtevaju idempotency key-ove za POST/PATCH/DELETE operacije da spreče duplicate processing. Ako Gin handler ne implementira idempotency mehanizam, može doći do:
- Duplih naplaćivanja istog payment-a
- Kreiranja multiple subscription-a za istu kupovinu
- Race condition-a u payment processing logici

#### Napadi

**Scenario 1: Network Retry Double Charge**

Korisnik klikne "Subscribe to Premium" → Browser šalje:
```http
POST /api/subscriptions/create
{"plan_id": "premium"}
```

Network timeout se desi nakon 30s. Browser automatski retry-uje zahtev.

**Timeline:**
```
T0: Request 1 → Gin goroutine 1 → Stripe PaymentIntent 1 kreiran
T1: Timeout (frontend ne dobije response)
T2: Request 2 (retry) → Gin goroutine 2 → Stripe PaymentIntent 2 kreiran
```

Rezultat:
- User naplaćen dvaput
- Dve subscription-a u bazi sa istim user_id
- Korisnici su besni (refund zahtevi)

**Scenario 2: Deliberate Double-Click Attack**

Napadač namerno klikće "Pay" dugme brzo 5 puta uzastopno. Bez rate limiting-a i idempotency zaštite, 5 payment-ova se kreira paralelno u različitim goroutine-ima.

**Scenario 3: Race Condition Exploit**

Napadač šalje 10 simultanih zahteva za istu subscription purchase koristeći script:
```bash
for i in {1..10}; do
  curl -X POST /api/subscriptions/create &
done
```

Svih 10 zahteva stiže gotovo istovremeno. Gin obrada u 10 različitih goroutine-a. Ako handler nema idempotency check, kreiraju se 10 subscription-a i 10 Stripe PaymentIntent-ova.

#### Mitigacije

**Server-Side Idempotency Key Generation:**

Handler generiše deterministički idempotency key baziran na:
- User ID (iz JWT token-a)
- Plan ID
- Timestamp (round-down na minut/sat)

Formula: `idempotency_key = SHA256(user_id + plan_id + timestamp)`

Pre procesiranja, handler proverava Redis/Database:
```
IF EXISTS(idempotency_key):
  RETURN cached_result
ELSE:
  Process payment
  STORE(idempotency_key, result)
```

Svaki request sa istim user+plan kombinacijom u vremenskom prozoru vraća isti rezultat bez duplicate processing-a.

**Stripe Idempotency Key Header:**

Stripe API podržava `Idempotency-Key` header. Handler generiše unique key per user action i prosleđuje ga Stripe-u:
```
POST /v1/payment_intents
Idempotency-Key: usr_123_premium_20260101
```

Stripe garantuje da će svi zahtevi sa istim key-om kreirati samo jedan PaymentIntent, čak i ako handler pošalje duplicate zahteve.

**Database Unique Constraint:**

Dodati unique constraint u bazu: `UNIQUE(user_id, plan_id, idempotency_key)`. Drugi simultani zahtev će failovati sa database constraint violation umesto da kreira duplicate. Handler hvata error i vraća cached result iz prvog uspešnog zahteva.

