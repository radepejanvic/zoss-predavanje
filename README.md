---
date: 2025-12-24
---
## 1. Domen problema
Softver se nalazi u domenu pametne urbanog transporta (Smart Mobility).  
Cilj sistema je da korisnicima omogući planiranje svakodnevnog putovanja, izbor i korišćenje različitih vidova prevoza, kao i kupovinu karata i pretplata, kroz jedinstvenu platformu koja integriše više nezavisnih provajdera prevoza.

Sistem objedinjuje podatke o javnom prevozu, servisima za iznamljivanje bicikala i skutera, kao i dodatnim eksternim servisima (plaćanje, autentifikacija), i pruža podršku za analizu i optimizaciju kretanja na nivou grada.
### Učesnici

| Učesnik                                        | Opis                                                                                                                                                      |
| ---------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Putnici                                        | Koriste sistem za planiranje putovanja, izbor prevoznih sredstava i kupovinu karata.                                                                      |
| Provajderi prevoza                             | Organizacije koje upravljaju javnim prevozom, servisima za iznajmljivanje bicikala i trotineta i izlažu podatke o dostupnosti i rasporedima putem API-ja. |
| Administratori sistema                         | Održavaju i konfigurišu sistem, upravljaju integracijama i nadgledaju rad platforme.                                                                      |
| Eksterni servisi                               | Servisi za autentifikaciju i autorizaciju korisnika, kao i servisi za obradu plaćanja.                                                                    |
| Sistemi za analitiku i optimizaciju saobraćaja | Koriste agregirane i anonimizovane podatke za analizu obrazaca kretanja i unapređenje saobraćajnih tokova.                                                |
### Poslovni procesi

| Poslovni proces                   | Opis                                                                                         |
| --------------------------------- | -------------------------------------------------------------------------------------------- |
| Planiranje putovanja              | Kombinovanje više vidova prevoza u optimalnu rutu na osnovu dostupnosti i vremenskih uslova. |
| Integracija prevoznih servisa     | Prikupljanje i sinhronizacija podataka od eksternih provajdera prevoza putem API-ja.         |
| Upravljanje kartama i pretplatama | Kupovina, validacija i upravljanje pravima korišćenja prevoznih usluga.                      |
| Obrada plaćanja                   | Bezbedna obrada transakcija putem eksternog payment provajdera.                              |
| Prikupljanje i analiza podataka   | Skladištenje i obrada istorijskih podataka o saobraćaju radi analitike i optimizacije.       |
| Korisnička podrška (sekundarno)   | Automatizovana podrška korisnicima kroz chat servis zasnovan na LLM-u.                       |
## 2. Arhitektura sistema
Zamišljeni softver je projektovan kao distribuiran sistem zasnovan na mikroservisnoj arhitekturi, gde su ključne funkcionalnosti podeljene u nezavisne servise. Na slici 1 prikazan je pregled arhitekture sistema na apstraktnom nivou. 
![[arhitektura.svg]]
*Slika 1: Arhitektura sistema*
### Osnovne karakteristike
Mikroservisna arhitektura sa sledećim ključnim osobinama:
- **Distribuirani sistem** - Svaki servis (Route Planning, Transportation Booking, User Service, Subscription & Ticket Service, Analytics Service) je nezavisan mikroservis
- **Database per Service pattern** - Svaki servis ima svoju bazu (Users DB, Travel History DB, Subscriptions & Tickets DB, Analytics Data Storage)
- **API Gateway pattern** - Centralna tačka ulaza koja rutira zahteve ka odgovarajućim servisima
- **Adapter pattern za eksterne integracije** - Mobility Data Adapter i Traffic Data Adapter enkapsuliraju komunikaciju sa eksternim providerima 
- **Event-driven komponente** - Analytics Data Streaming radi asinhronu obradu podataka za real-time analitiku
- **Caching strategy** - Routes Cache i Analytics Cache poboljšavaju performanse za često korišćene podatke
### Korišćene tehnologije
- **Web/Mobile App** (Frontend): 
	- **React.js** (Web) - Moderni UI framework
	- **React Native** (Mobile) - Cross-platform razvoj (iOS + Android) sa deljenjem koda sa web aplikacijom
- **Traefik** (API Gateway) - Golang implementacija - visoke performanse, mali *memory footprint*, *rate limiting*, *authentication*, *load balancing*, *service discovery* i *health checks*, idealan kao *single entry* point za mikroservise
- **Gin + Golang** (Backend Microservices) - visoke performanse, *low latency*, brzo pokretanje, kompaktne Docker slike, statički tipiziran, nizak *memory footprint*
- **Redis** (Cache) - *blazing-fast-in-memory* performanse, *low latency*, podržava širok spektar struktura, *sub-millisecond latency* perzistencija (RDB + AOF) za *crash recovery*
- **MongoDB** (Databases) - dokument orijentisana NoSQL baza podataka, pogodna za čuvanje podataka promenljive strukture (rute sa različitim brojem prevoznih opcija)
- **GTFS** (General Transit Feed Specification) API - rasporedi javnog prevoza, procenjeno vreme dolaska, cene, dostupnost
- **Uber/Bolt/Yandex API** - taxi i uber provajderi, procenjeno vreme dolaska, cene, dostupnost
- **Lime Micromobility API** - provajder bicikala i trotineta, cene, dostupnost
- **Google Maps Traffic API** - stanje i prohodnost puteva u realnom vremenu 
- **Auth0** (Auth Provider) - OAuth 2.0 / OpenID Connect, Social login (Google, Facebook, Apple)
- **Stripe** (Payment Provider) - kartice, digitalni novcanici, PCI DSS kompatibilnost eksterno ishendlana
- **Apache Kafka** (Analytics Data Streaming) - distributed event streaming, travel-events, subscription-events, payment-events, visoka dostupnost
- **Hadoop** (Analytics Data Storage) - distribuirano skladište za petabyte-scale podatke
## 3. Slučajevi korišćenja
- **Planiranje putovanja** - pretraživanje i planiranje optimalne rute između lokacija koristeći različite vidove prevoza (javni prevoz, taxi/Uber, bicikli/skuteri). Sistem uzima u obzir real-time saobraćajne podatke i dostupnost prevoznih sredstava.
- **Rezervacija i kupovina karata** - rezervisanje i kupovina karata/pretplate za različite vidove transporta. Sistem omogućava upravljanje aktivnim pretplatama, istorijom kupovina i plaćanje preko eksternih payment providera.
- **Upravljanje korisničkim nalogom** - registracija, prijava (preko Auth Provider), upravlja profilom i preferencama, čuva omiljene rute i pristupa istoriji svojih putovanja.
- **Praćenje i analitika putovanja** - sistem prikuplja i analizira podatke o putovanjima korisnika, obrascima kretanja i korišćenju različitih vidova prevoza. Omogućava personalizovane preporuke i uvide u navike putovanja.
## 4. Osetljivi resursi sistema

**Users DB (baza korisničkih podataka)**
- **Bezbednosni ciljevi:** poverljivost, integritet, dostupnost
- **Osetljivi podaci:** Lični podaci korisnika (ime, prezime, email, telefon, adresa, datum rođenja), kredencijali za autentifikaciju
- **Regulativa:** GDPR (General Data Protection Regulation) - zahteva enkripciju ličnih podataka, pravo na brisanje, kontrolu pristupa
- **Pretnje:** Neovlašćen pristup, curenje podataka, identity theft

**Subscriptions & Tickets DB (baza pretplata i karata)**
- **Bezbednosni ciljevi:** integritet, poverljivost, autentičnost
- **Osetljivi podaci:** Aktivne pretplate korisnika, kupljene digitalne karte, istorija plaćanja, statusи pretplata (active/expired), validacioni tokeni za digitalne karte
- **Regulativa:** GDPR - podaci o kupovinama i pretplatama korisnika se smatraju ličnim podacima
- **Pretnje:** Neovlašćena aktivacija/produženje pretplata bez plaćanja, onemogućavanje validnih pretplata drugih korisnika, krađa/falsifikovanje digitalnih karata, manipulacija statusom pretplata

**Travel History DB (baza istorije putovanja)**
- **Bezbednosni ciljevi:** poverljivost, integritet
- **Osetljivi podaci:** Obrasci kretanja korisnika, lokacijski podaci, navike putovanja
- **Pretnje:** Profilisanje korisnika, praćenje lokacije, narušavanje privatnosti

**Routes Cache** (keš isplaniranih ruta)
- **Bezbednosni ciljevi:** integritet, dostupnost
- **Osetljivi podaci:** Keširane rute korisnika
- **Pretnje:** Cache poisoning, curenje informacija o korisničkim navikama

**Analytics Data Streaming (Kafka/Message Broker)**
- **Bezbednosni ciljevi:** integritet, dostupnost, poverljivost
- **Osetljivi podaci:** Real-time podaci o putovanjima korisnika, događaji o rezervacijama, obrasci kretanja
- **Pretnje:** Neovlašćeno čitanje event stream-a, message tampering, DoS napadi na broker, replay napadi

**Analytics Data Storage (Hadoop/Data Lake)**
- **Bezbednosni ciljevi:** poverljivost, integritet, dostupnost
- **Osetljivi podaci:** Istorijski podaci o svim putovanjima, agregirana analitika korisničkog ponašanja, big data skupovi
- **Pretnje:** Masovno curenje podataka, neovlašćena analiza korisničkih obrazaca, data mining bez saglasnosti

**Analytics Cache**
- **Bezbednosni ciljevi:** integritet, poverljivost, dostupnost
- **Osetljivi podaci:** Keširani rezultati analitike, često tražene statistike, personalizovani insights
- **Pretnje:** Cache poisoning, curenje agregiranih podataka, neovlašćen pristup analitičkim rezultatima

**Configuration Files (Konfiguracioni fajlovi)**
- **Bezbednosni ciljevi:** poverljivost, integritet, dostupnost
- **Osetljivi podaci:** API ključevi za eksterne providere (mobility, traffic, payment), database connection stringovi, OAuth client secrets, enkripticioni ključevi, servisni tokeni
- **Pretnje:** Neovlašćen pristup API ključevima, kompromitovanje kredencijala za baze, izloženost secrets u version control sistemima, hardcoded credentials