# Kafka
## Uvod 
Apache Kafka je distribuisana platforma za strimovanje podataka u realnom vremenu, dizajnirana za izgradnju brzih i otpornih tokova podataka (data pipelines). Za razliku od tradicionalnih sistema za razmenu poruka, Kafka funkcioniše kao distribuirani log-fajl koji garantuje visok propusni opseg (high throughput), nisku latenciju i horizontalnu skalabilnost.

U osnovni Kafka se zasniva na Publish-Subscribe modelu:
- Producers: Aplikacije koje šalju (objavljuju) podatke u Kafku.
- Consumers: Aplikacije koje se pretplate na određene kanale i čitaju podatke.
- Topics: Logički kanali ili kategorije u koje se poruke raspoređuju. Svaki topic je podeljen na particije radi paralelizacije rada.

Pored osnovnog Publish-Subscribe modela, Kafka ima opciju da tretira podatke kao neprekidne događaje (event streaming). Za razliku od standardnih Message Broker-a koji brišu poruku čim je consumer potvrdi, Kafka zadržava podatke na disku prema definisanom periodu (retention policy), omogućavajući consumer-ima da premotaju i ponovo obrađuju iste podatke.

### Primene Kafke u realnom svetu
Zahvaljujući svojoj robusnosti, Kafka se najčešće primenjuje u sledećim scenarijima:
- Log Aggregation: Prikupljanje logova iz stotina različitih mikroservisa i njihovo centralizovano slanje u sisteme za analizu. Kafka ovde služi kao bafer koji sprečava da sistem za analizu bude preopterećen naletima logova.
- Stream Processing: Kontinuirana obrada i transformacija podataka u realnom vremenu. 
- Event Sourcing: Skladištenje svake promene stanja u sistemu kao niza događaja. Ovo omogućava potpunu rekonstrukciju stanja sistema u bilo kom trenutku u prošlosti.
- Commit Log za distribuirane sisteme: Kafka služi kao eksterni log koji garantuje da će svi povezani servisi videti isti redosled događaja, što je ključno za konzistentnost podataka.

### Ključne specifičnosti i prednosti
- Decoupling: Producer-i i consumer-i su potpuno nezavisni. Producer ne mora da zna ko čita podatke, niti koliko consumer-a postoji, što olakšava skaliranje sistema.
- Trajnost i pouzdanost: Podaci su nepromenljivi (immutable) i zapisuju se sekvencijalno na disk. Sekvencijalni upis je drastično brži od nasumičnog, što omogućava Kafki da postigne milionske propusne opsege uz minimalnu latenciju.
- Garantovan redosled: Kafka garantuje redosled poruka unutar jedne particije. 

## Arhitektura
<img width="676" height="444" alt="Kafka-architecture" src="https://github.com/user-attachments/assets/2d1a1f49-175a-44c1-9f9a-3f30406d2688" />

Arhitektura Kafke je projektovana da podrži ekstremno visoku propusnu moć uz zadržavanje integriteta podataka. Sistem funkcioniše kao klaster čvorova koji međusobno sarađuju kako bi omogućili paralelnu obradu i skladištenje dolaznih strimova informacija.

### Kafka Broker
Broker je osnovna radna jedinica ili server unutar Kafka klastera. Njegova primarna uloga je da prima poruke od producer-a, dodeljuje im offset-e i bezbedno ih zapisuje na disk, dok istovremeno opslužuje zahteve consumer-a koji žele te podatke da pročitaju. Klaster se sastoji od više brokera koji dele opterećenje i između sebe, time se postiže da sistem ostane operativan čak i ako pojedinačni serveri prestanu sa radom. Svaki broker je zadužen za određeni skup particija. Iako svaki broker može da obrađuje hiljade poruka u sekundi, on ne radi u izolaciji, brokeri neprestano komuniciraju kako bi osigurali da su podaci replikovani i da je stanje klastera konzistentno.

### Topici i Particije
Topik je logička kategorija u koju se poruke raspoređuju. Svaki topik je podeljen na više particija koje su distribuirane širom brokera u klasteru. Ovo omogućava horizontalno skaliranje. Više producer-a može istovremeno da piše u različite particije istog topika, dok više consumer-a može istovremeno da čita iz njih, čime se postiže ogromna paralelizacija.

Unutar svake particije, poruke su strogo poređane i nepromenljive. Svaka poruka dobija svoj offset (jedinstveni redni broj), koji služi kao jedini dokaz o njenoj poziciji. Ovakav dizajn omogućava konzumentima da čitaju podatke sopstvenim tempom, takođe mogu da nastave tamo gde su stali ili da se vrate unazad i ponovo obrade stare događaje, što je nemoguće kod tradicionalnih redova poruka.

### ZooKeeper (i KRaft)
ZooKeeper je eksterni servis koji se ponaša kao koordinator Kafka klastera. Čuva sve kritične metapodatke:
- informacije o tome koji su brokeri aktivni
- gde se nalaze particije
- ko je trenutno "vođa" (Leader) određene particije koji sme da prima upise.

Bez ZooKeeper-a, brokeri ne bi znali kako da se sinhronizuju u slučaju pada jednog od njih. 

Novije verzije Kafke uvode KRaft protokol koji eliminiše potrebu za ZooKeeper-om, integrišući upravljanje metapodacima direktno u samu Kafku. Ovo pojednostavljuje arhitekturu i omogućava klasteru da se brže oporavi nakon grešaka. Bez obzira na mehanizam, uloga koordinacije ostaje ista – osigurati da sistem uvek zna tačnu topologiju mreže i distribuciju podataka.

### Producer-i i Consumer-i
Producer-i su aplikacije koje generišu podatke i šalju ih u Kafku. Oni su "pametni" jer sami odlučuju u koju će particiju poslati podatak (obično na osnovu ključa), čime se osigurava da svi povezani događaji zavrse na istom mestu. S druge strane, consumer-i su aplikacije koje povlače te podatke. Oni su organizovani u Consumer Groups, gde više instanci deli posao čitanja iz jednog topika tako da svaka instanca dobije svoj deo particija, što sprečava dupliranje posla i omogućava ogromnu brzinu obrade.



## Reference
- [Hello Interview](https://www.hellointerview.com/learn/system-design/deep-dives/kafka)
- [Research Gate](https://www.researchgate.net/figure/Kafka-architecture_fig1_347866161)
- [Apache Kafka](https://kafka.apache.org/11/streams/architecture/)
