# Redis
## Uvod
Redis (Remote Dictionary Server) je open-source sistem za skladištenje struktura podataka u memoriji, napisan u programskom jeziku C. Za razliku od tradicionalnih baza podataka, Redis je dizajniran kao samo opisujući (self-described) sistem, što znači da ne zahteva eksternu šemu (schema-less). Svaki podatak koji se čuva u sebi nosi metapodatke o svom tipu. Njegova arhitektura je optimizovana za visoku brzinu. Operacije obavlja direktno u RAM memoriji bez pristupa SSD-u ili HDD-u. Potencira ekstremnu efikasnost i brzinu operacija u odnosu na trajnost (durability) podataka.

U osnovi je NoSQL baza koja skladišti podatke u formatu ključ-vrednost (key-value store), ali podržava i skup raznih struktura podataka kao što su:
- osnovne strukture: stringovi, hash-evi (objekti), liste i mape.
- napredne strukture: setovi i sortirani setovi, bloom filteri, geoprostorni indeksi i time-series podaci.

Neke od primena Redisa su:
- klasično keširanja sa definisanim vremenom trajanja ključeva (TTL)
- implementacija distribuiranih brava (distributed locks) i zajedničkih brojača koji sprečavaju konflikte u distribuiranim sistemima
- rang-liste (leaderboards) implementirane pomoću sortiranih setova
- ograničavanje protoka (rate limiting) zahteva pomoću sliding window mehanizama
- geoprostornu pretragu
- strimovanje događaja (slično Kafki) putem Append-only logova
- brzu razmenu poruka u realnom vremenu kroz Pub/Sub kanale

U Redisu dizajn ključeva direktno diktira performanse i skalabilnost sistema. Preporučena praksa je korišćenje namespaces šeme (npr. user:123:profile) radi logičkog grupisanja povezanih podataka. Ovakva granulacija omogućava efikasniju evikciju (eviction) i postavljanje vremena isteka (expiration) na nivou pojedinačnih atributa, umesto na nivou celog objekta.

Loš dizajn, koji podrazumeva skladištenje velikih JSON objekata pod jednim ključem (npr. user123:json-blob), stvara uska grla u memoriji i mrežnom saobraćaju, jer svaki upit zahteva preuzimanje celokupnog seta podataka. Nasuprot tome, razbijanje podataka u specifične strukture (npr. user:123:posts kao list ili user:123:settings kao hash) omogućava izvršavanje atomičnih operacija nad delovima podataka bez uticaja na ostatak sistema.

Skaliranje sistema sa velikim brojem ključeva rešava se kroz Redis Cluster pomoću particionisanja (sharding). Prostor ključeva je podeljen na 16384 hash slota, gde algoritam određuje na kom se čvoru nalazi određeni ključ na osnovu njegove vrednosti. Klijenti keširaju mapu ovih slotova kako bi direktno komunicirali sa relevantnim čvorom.

Čvorovi unutar klastera koriste gossip protokol za razmenu informacija o stanju mreže i detekciju otkaza. Ovakva arhitektura omogućava horizontalno skaliranje i sprečava pojavu vrućih ključeva (hot keys) koji bi mogli da preopterete pojedinačne čvorove usled lošeg dizajna prostora ključeva.

## Arhitektura
![redis-architectzure](https://github.com/user-attachments/assets/ec8c36a8-ce60-4fb2-9ffc-eaf43d54ba2f)

Arhitektura Redisa je slojevita i dizajnirana da maksimizira propusnu moć RAM memorije uz minimalni overhead procesora. Podeljen je na komponentu mrežni i klijentski interfejs, jezgra servera, upravljanja memorijom, perzistenciju, klasterovanje i distribuciju. 

### Jezgro Redis Server (daemon)
Centralni deo arhitekture čini single-threaded event loop napisan u C-u. Koristi jedan proces i jednu glavnu nit za izvrsavanje svih komandi, čime se eliminiše potreba za zaključavanjem podataka i sprečava gubitak performansi usled promene konteksta (context switching) između procesa.  
- I/O Multiplexing: Koriste se sistemski pozivi poput epoll (Linux) ili kqueue (BSD) koji omogućavaju jednoj niti da istovremeno nadgleda hiljade mrežnih konekcija i obrađuje ih čim postanu spremne
- Command Execution: svaka komanda prolazi kroz parser i execution engine sekvencijalno, što garantuje atomičnost svake operacije bez dodatnih troškova sinhronizacije

### Sloj perzistencije
Iako je Redis implementiran kao in-memory sistem, obezbeđuje neki vid trajnosti podataka kroz sloj perzistencije i njegovih mehanizama. 
- RDB (Redis Database Snapshots) - periodično kreiranje binarnih snimaka (snapshot-ova) celokupnog skupa podataka na disku. Ovaj posao se izvršava u child niti Redis procesa (sistemski poziv fork()) kako se ne bi blokirao rad glavne niti koja obrađuje komande.
- AOF (Append Only File) - logovanje svake operacije pisanja (write) u realnom vremenu. Redis omogućava rayličite strategije sinhronizacije (svake sekunde, pri svakom upitu ili prepuštanje operativnom sistemu). Ovaj mehanizam logovanja je ključan za sisteme gde je prioritet minimizacija gubitka podataka u slučaju pada servera.

### Visoka dostupnost
Redis Sentinel (sistem za preživljavanje) je poseban proces koji radi sa strane i nadgleda Redis servere. 
- Monitoring: Sentinel stalno šalje ping primarnom (Master) čvoru i replikama da proveri da li su živi.
- Automatic Failover: Ako Master prestane da odgovara, Sentineli glasaju između sebe (postizanjem kvoruma) da potvrde da je Master zaista mrtav.
- Promocija: Kada potvrde pad, Sentinel automatski uzima jednu od replika i unapređuje je u novog Mastera.
- Obaveštavanje klijenata: Sentinel javlja aplikaciji novu adresu Mastera, tako da sistem nastavlja da radi bez ručne intervencije.

### Skaliranje i klasterovanje
Redis Cluster služi za proširenje kapaciteta (skaliranje). Kada podaci postanu preveliki da bi stali u RAM jednog servera, oni se moraju podeliti na više mašina pomoću particionisanja (sharding).
- Hash Slots (16,384): Prostor ključeva je podeljen na tačno 16,384 logičke jedinice (slota). Prilikom unosa ključa (npr. user:123), Redis koristi CRC16 algoritam kako bi izračunao kom tačno slotu taj ključ pripada.
- Distribucija podataka: Ovi slotovi se raspoređuju na dostupne servere. U sistemu sa tri servera, svaki čvor preuzima odgovornost za trećinu ukupnog broja slotova, čime se obezbeđuje da podaci budu ravnomerno raspoređeni.
- Gossip protokol: Čvorovi u klasteru neprekidno razmenjuju informacije (tračare) o svom statusu i statusu susednih čvorova. Na ovaj način svaki čvor u svakom trenutku zna koji su drugi serveri aktivni i koji opseg slotova trenutno opslužuju.
- Pametni klijenti: Aplikacija (klijent) poseduje mapu slotova. Ukoliko klijent greškom pošalje upit pogrešnom serveru, taj server ga neće obraditi, već će klijentu poslati informaciju o tačnoj adresi servera koji čuva traženi podatak.

## Primena Redisa u arhitekturi sistema
U okviru prikazane arhitekture Smart Mobility sistema, Redis je implementiran kao keš sloj postavljen između mikroservisa koji rade kompleksne upite ili računanja i Web aplikacije koja te podatke povlači i prikazuje korisniku, sa ciljem da se smanji latencija.

Integracija se oslanja na `go-redis` klijent koji omogućava efikasnu komunikaciju između Gin hendlera i Redis servera:
- Connection Pooling: Klijent automatski održava kolekciju otvorenih konekcija ka Redisu. Ovo je ključno za Gin jer sprečava otvaranje novog TCP soketa za svaki HTTP zahtev, čime se drastično smanjuje latencija.
- Konkurentnost (Goroutines): `go-redis` je thread-safe, što omogućava da hiljade konkurentnih Go rutina unutar Gin-a istovremeno koriste istu instancu klijenta bez rizika od konflikata.

## Reference
- [Hello Interview](https://www.hellointerview.com/learn/system-design/deep-dives/redis)
- [Redis Docs](https://redis.io/docs/latest/)
- [Architecture Notes](https://architecturenotes.co/p/redis)
- [Medium](https://ashutoshkumars1ngh.medium.com/redis-architecture-why-redis-is-so-fast-and-a-deep-dive-into-internals-d153c064a549)
