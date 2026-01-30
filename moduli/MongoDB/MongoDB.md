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

### MongoDB Server (mongod)
mongod je primarni sistemski proces (daemon) koji upravlja svim kritičnim aspektima baze podataka. Njegova glavna uloga je direktna interakcija sa Storage Engine slojem kako bi se osiguralo da su BSON dokumenti pravilno zapisani na disk ili pročitani iz memorije. Pored samog skladištenja, ovaj proces je odgovoran za sprovođenje sigurnosnih polisa, upravljanje indeksima i obradu svih dolaznih zahteva od strane klijentskih aplikacija.

### MongoDB Shell (mongosh)
mongosh predstavlja interaktivni JavaScript interfejs koji služi kao primarni alat za administraciju i brzu manipulaciju podacima direktno iz komandne linije. Za razliku od aplikativnih drajvera, shell omogućava testiranje kompleksnih upita u realnom vremenu, menjanje konfiguracije baze u hodu ili inspekciju indeksa bez potrebe za pisanjem eksternog koda.

### Cluster formacija
U produkcionim okruženjima, MongoDB se retko koristi kao jedna instanca; umesto toga, više mongod procesa se povezuje u klaster kako bi se osigurala visoka dostupnost i horizontalno skaliranje. Klaster formacija omogućava sistemu da nastavi sa radom čak i ako dođe do fizičkog otkaza jednog ili više servera, koristeći mehanizam automatskog prebacivanja na ispravne čvorove (failover) unutar Replica Setova. Osim otpornosti na otkaze, klaster omogućava distribuciju ogromnih količina podataka kroz Sharding proces. Na ovaj način, opterećenje upita se deli između više mašina, što sprečava da jedan server postane usko grlo. 

### Klijentske biblioteke (Drivers)
Drajveri predstavljaju komunikacioni sloj koji omogućava aplikacijama napisanom u različitim jezicima da razgovaraju sa MongoDB serverom. Njihova primarna uloga je serijalizacija i deserijalizacija podataka, odnosno prevođenje izvornih programskih struktura u BSON format koji baza razume i obrnuto. 

Pored same konverzije podataka, drajveri upravljaju i kompleksnim mrežnim operacijama kao što je održavanje kolekcije konekcija (connection pooling) i automatsko rutiranje upita ka odgovarajućim čvorovima u klasteru. Oni su svesni topologije klastera, pa ako jedan čvor otkaže, drajver će automatski preusmeriti zahtev na drugu dostupnu repliku, čime se osigurava neometan rad aplikacije bez potrebe za ručnom rekonfiguracijom koda.

### Mehanizam za skladištenje (Storage Engine)
Storage Engine je unutrašnja komponenta MongoDB-a zadužena za direktno upravljanje načinom na koji se podaci čuvaju na fizičkom disku ili u RAM memoriji. On deluje kao posrednik između mongod procesa i memorijskog podsistema, a od njegove konfiguracije direktno zavise performanse čitanja i pisanja, kao i efikasnost kompresije podataka. MongoDB je jedinstven po tome što podržava promenu mehanizma skladištenja u zavisnosti od potreba aplikacije, što omogućava optimizaciju baze za različite scenarije upotrebe.

Podrazumevani mehanizam od verzije 3.0 je WiredTiger, koji je dizajniran za sisteme sa visokim intenzitetom upisa podataka. On koristi napredne tehnike poput document-level zaključavanja, što omogućava da više korisnika istovremeno menja različite dokumente u istoj kolekciji bez blokiranja sistema. Pored toga, WiredTiger nudi efikasnu kompresiju koja može smanjiti zauzeće diska i do 80%, dok istovremeno koristi keširanje u memoriji kako bi obezbedio brzinu pristupa najčešće korišćenim podacima koja se meri u milisekundama.

### Bezbednost 
MongoDB pruža višeslojni sigurnosni model za zaštitu podataka u sistemu:
- Autentifikacija: Provera identiteta korisnika i servisa pre pristupa bazi.
- Autorizacija (RBAC): Dodela specifičnih rola korisnicima, osiguravajući da svaki servis pristupa samo onim podacima koji su mu neophodni.
- Enkripcija: Podrška za TLS/SSL protokole za zaštitu podataka u tranzitu i enkripcija podataka na samom disku (at rest).

## Primena MongoDB-a u arhitekturi sistema
U okviru Smart Mobility sistema, MongoDB služi kao primarno skladište podataka.
- Users DB: Čuva profile korisnika
- Travel History DB: Skladišti istoriju putovanja, različiti tipovi transporta (bus, taxi, bike) mogu imati različite atribute, lako se rešava kroz fleksibilnu šemu.

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
