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

## Arhitektura
![redis-architectzure](https://github.com/user-attachments/assets/ec8c36a8-ce60-4fb2-9ffc-eaf43d54ba2f)


## Reference
https://architecturenotes.co/p/redis
https://ashutoshkumars1ngh.medium.com/redis-architecture-why-redis-is-so-fast-and-a-deep-dive-into-internals-d153c064a549
