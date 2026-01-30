# MongoDB
## Uvod
MongoDB je open-source, dokument-orijentisana NoSQL baza podataka napisana u programskom jeziku C++. Za razliku od relacionih baza, MongoDB koristi fleksibilnu šemu, što znači da dokumenti unutar iste kolekcije ne moraju imati identičnu strukturu. Podaci se skladište u BSON formatu (Binary JSON), koji omogućava podršku za dodatne tipove podataka i bržu pretragu u poređenju sa običnim JSON-om.

Osnovna organizacija podataka prati hijerarhiju: Database > Collection > Document.
- Dokument: Jedinični zapis podataka u formatu ključ-vrednost, sličan JSON objektu.
- Kolekcija: Skup dokumenata koji je funkcionalni ekvivalent tabeli u relacionim bazama, ali bez fiksne strukture.
- Schema-less priroda: Struktura podataka se može menjati u hodu, mada se validacija šeme može nasilno primeniti radi održavanja konzistentnosti.

Modelovanje relacija između entiteta se u MongoDB-u izvodi na dva načina:
- Embed (Ugnježdavanje): Skladištenje povezanih podataka unutar jednog dokumenta, što je idealno za podatke koji se često čitaju zajedno i čija je veličina ograničena.
- Reference: Povezivanje dokumenata putem jedinstvenih identifikatora (_id), što je pogodno za velike skupove podataka i sprečavanje prekomernog dupliranja.

## Arhitektura
<img width="1366" height="788" alt="sharding-and-replica-sets" src="https://github.com/user-attachments/assets/12fe0bfc-42f0-4abe-9cf6-3ff9a199847d" />

Arhitektura MongoDB-a je dizajnirana da podrži visoku dostupnost i horizontalno skaliranje, uz oslanjanje na efikasne mehanizme za upravljanje memorijom i upitima. Sastoji se od nekoliko osnovnih komponenti koje zajednički omogućavaju efikasno skladištenje, preuzimanje i obradu podataka.

Terminologija vezana za MongoDB:
- Shard: Predstavlja server koji skladišti deo baze podataka, upravljajući specifičnim blokovima (chunks) informacija.
- Server: Hostuje MongoDB instance, primarni server obrađuje upise, dok sekundarni repliciraju podatke radi sigurnosti.
- MongoDB Engine: Izvršava operacije čitanja, pisanja, ažuriranja, upita i agregacija nad podacima.
- Config Server: Čuva metapodatke i informacije o šardovanju koji su neophodni za pravilno rutiranje upita unutar klastera.
- Mongos (Query router): Služi kao ruter koji prosleđuje zahteve klijenata ka odgovarajućem šardu na osnovu ključa.
- Replica Sets: Grupa servera koji održavaju identičan skup podataka kako bi se obezbedila redundansa i automatski oporavak u slučaju otkaza (failover).

### Jezgro sistema i upravljanje (mongod & mongosh)
MongoDB arhitektura se oslanja na specifične procese i alate koji omogućavaju rad baze i interakciju sa podacima:
- MongoDB Server (mongod): Centralni proces (daemon) koji predstavlja srce sistema. Njegove odgovornosti su skladištenje podataka, upravljanje pristupom i izvršavanje svih operacija nad bazom.
- MongoDB Shell (mongosh): Interaktivni JavaScript interfejs (CLI) koji omogućava administratorima i developerima da direktno upravljaju bazom, vrše upite i održavaju sistem.
- Cluster formacija: Više mongod instanci se može povezati u klaster, čime se postiže distribucija podataka i otpornost na otkaze.

## Primena MongoDB-a u arhitekturi sistema
U okviru Smart Mobility sistema, MongoDB služi kao primarno skladište podataka.
- Users DB: Čuva profile korisnika
- Travel History DB: Skladišti istoriju putovanja, različiti tipovi transporta (bus, taxi, bike) mogu imati različite atribute, lako se rešava kroz fleksibilnu šemu.

### Implementacija (Golang, Gin)
Povezivanje sa MongoDB-om se vrši putem zvaničnog `mongo-go-driver` paketa:
- Connection Pooling: Drajver održava kolekciju konekcija ka klasteru, čime se optimizuje broj otvorenih soketa prema mongod procesima.
- BSON Mapping: Go strukture (structs) koriste tagove (npr. bson:"name") kako bi se podaci automatski mapirali iz baze u objekte unutar Gin hendlera.
- Konkurentnost: Svaka operacija nad bazom unutar Gin rute se izvršava asinhrono, koristeći context paket za kontrolu tajmauta i otkazivanje zahteva.

## Reference
- [Geeks for Geeks](https://www.geeksforgeeks.org/mongodb/mongodb-architecture/)
- [GitHub](https://github.com/minhhungit/mongodb-cluster-docker-compose?tab=readme-ov-file)
- [Hello Interview](https://www.hellointerview.com/learn/system-design/in-a-hurry/key-technologies)
- [ChatGPT](https://chatgpt.com/share/696bac70-8ff4-8002-bd99-f380f831fbdb)
- [Hello Interview](https://www.hellointerview.com/learn/system-design/core-concepts/sharding)
