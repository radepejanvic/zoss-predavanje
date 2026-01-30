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
