# Pretnje, napadi i mitigacije na Route Planning Service (Redis)
## Uvod
Ovaj dokument analizira bezbednosne pretnje servisa za planiranje ruta sa fokusom na Redis koji se koristi kao Routes Cache. Analiziran je kritičan propust koji omogućava uskraćivanje usluge (DoS) čak i bez poznavanja lozinke za pristup bazi.

## Kontekst Sistema
Route Planning Service koristi Redis za skladištenje izračunatih ruta kako bi se smanjilo opterećenje na eksterne provajdere (Google Maps, GTFS, Uber) i ubrzao odgovor korisniku.

## Resursi u Redis-u:
- Routes Cache: Privremeno skladištenje optimalnih putanja.
- Session Metadata: Podaci o trenutnim zahtevima korisnika.
- External Provider Quotas: Praćenje broja poziva ka eksternim API-jima.
- Kritični resursi pod rizikom su dostupnost sistema i performanse, jer pad keša uzrokuje preopterećenje backend servisa.

## Katalog Napada
### 1) Redis NOAUTH DoS - Resource Exhaustion (CVE-2025-21605)
Ova ranjivost omogućava napadaču da pošalje veliki broj komandi koje zahtevaju autentifikaciju (npr. GET, AUTH, INFO), ali bez slanja ispravne lozinke. Iako Redis odbija ove komande sa NOAUTH greškom, on i dalje mora da obradi zahtev i smesti odgovor u izlazni bafer klijenta.

#### Resursi pod rizikom:
- CPU Usage: Procesorska snaga potrebna za upravljanje hiljadama konekcija.
- Network I/O & Memory: Izlazni baferi (output buffers) koji se pune podacima.
- Dostupnost servisa: Sposobnost Redisa da opsluži legitimne zahteve aplikacije.

#### Resursi i operativni procesi pod rizikom:
- Lojalnost korisnika: Gubitak milisekundnog odziva ruter-a direktno uzrokuje prelazak korisnika na konkurentske aplikacije zbog lošeg korisničkog iskustva.
- API troškovi: Pad keša forsira direktno pozivanje eksternih provajdera (Google, Uber), što dovodi do nekontrolisanog trošenja budžeta za API kvote.
- Sinhronizacija ponude: Onemogućavanje keša narušava real-time prikaz dostupnosti bicikala i taxi vozila, rezultirajući neuspešnim rezervacijama.
- Stabilnost u "špicu": DoS napad tokom saobraćajnih gužvi onemogućava skaliranje sistema kada je potražnja najveća, što vodi ka potpunom kolapsu usluge.

#### Provera prisustva ranjivosti
Provera se vrši uvidom u broj povezanih klijenata i stanje bafera tokom sumnjive aktivnosti:
```shell
redis-cli -a <password> info clients
```
Znakovi napada su nagli skok `connected_clients` (npr. sa uobičajenih 5-10 na 100+) i povećanje `client_recent_max_output_buffer`-a.

#### Napad
- Scenario: "Ugušivanje" keša putem neautentifikovanih zahteva
- Napadač koristi Python skriptu (`noauth_dos.py`) koja otvara TCP konekcije ka Redis portu (6379) i šalje komande bez čitanja odgovora.
- Mehanizam: Pošto napadač ne čita odgovore (NOAUTH error messages), Redis ih čuva u memoriji (output buffer) čekajući da ih klijent preuzme.
- Rezultat:  CPU Usage naglo skače jer Redis troši cikluse na upravljanje hiljadama malicioznih konekcija.
- Network I/O: Zagušenje mrežnog interfejsa kontejnera.
- Latencija: Legitimni Route Planning Service više ne može da dobije podatke iz keša u milisekundama, što usporava celu aplikaciju za krajnjeg korisnika.


```shell 
py .\noauth_dos.py
```

```shell
docker exec -it redis-cache redis-cli -a SuperSecretPassword123 info clients

# Clients
connected_clients:101 # Broj konekcija 
cluster_connections:0
maxclients:10000
client_recent_max_input_buffer:20480
client_recent_max_output_buffer:20504 # Veličina output buffer-a
blocked_clients:0
tracking_clients:0
clients_in_timeout_table:0
total_blocking_keys:0
total_blocking_keys_on_nokey:0
```
<img width="1097" height="491" alt="image" src="https://github.com/user-attachments/assets/377c6b8f-b14d-4a8b-bb8c-d9e913473dfd" />

<img width="1100" height="489" alt="image" src="https://github.com/user-attachments/assets/046720c6-e063-4956-b3e7-a21b288dcc71" />

#### Mitigacije
- Ažuriranje na bezbednu verziju: Instalacija zakrpe za Redis koja limitira resurse dodeljene neautentifikovanim klijentima.
- Client Output Buffer Limits: Konfigurisanje client-output-buffer-limit u `redis.conf` za normal klijente kako bi Redis automatski prekinuo konekciju koja puni memoriju.
- Mrežna zaštita (ACL & Firewall): Dozvoliti pristup Redis portu isključivo unutar interne Docker mreže za Route Planning Service.
- Rate Limiting na mrežnom nivou: Ograničiti broj novih konekcija po sekundi sa jedne IP adrese.
