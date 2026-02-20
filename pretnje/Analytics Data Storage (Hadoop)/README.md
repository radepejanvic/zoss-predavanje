# Pretnje, napadi i mitigacije na Analytics Data Storage (Hadoop)

## Uvod

Ovaj dokument analizira sigurnosne pretnje Hadoop ekosistema u kontekstu Smart Mobility platforme. Hadoop čuva milione osetljivih zapisa: GPS koordinate (Special Category Data), istoriju plaćanja, rute putovanja i obrasce ponašanja korisnika.

**Kritični resursi:**
- Zapisi o putovanjima (GPS koordinate, timestamps, korisnički ID-evi)
- Istorija plaćanja (transakcije, iznosi, metode plaćanja)
- Poslovna analitika (izveštaji o prihodima, predviđanja potražnje)

---

## Katalog Pretnji

### 1. Kerberos Keytab Leak (Demonstrirano)

**Opis ranjivosti:**  
Kerberos keytab fajlovi sa permisijama koje dozvoljavaju čitanje svim korisnicima omogućavaju kradnju autentifikacijskih kredencijala. Ovo je česta ranjivost u Hadoop deployment-ima zbog nedostatka post-install security audita.

**Lanac napada:**
Napadač koji ima pristup Docker host-u (lateral movement) može pronaći keytab fajl sa lošim permisijama, kopirati ga van sistema, autentifikovati se sa ukradenim kredencijalima, i izvršiti eksfiltraciju podataka iz HDFS-a.

**Posledica:**  
150,000+ zapisa o putovanjima ukradeno (GPS lokacije, podatci o plaćanju, istorija korisnika).

**Mitigacija:**  
Keytab fajlovi moraju imati restriktivne permisije (samo vlasnik može čitati), pravilnog vlasnika (hadoop system user), i periodičnu rotaciju kredencijala (na svakih 90 dana).

**📍 Za kompletnu demonstraciju (ranjiv klaster, exploit script, test data, automated mitigation):**  
**[GitHub - Hadoop Kerberos Attack Demo](https://github.com/radepejanvic/zoss-predavanje)**  

---

### 2. YARN Remote Code Execution

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Shell pristup NodeManager čvorovima (fajl sistem, memorija, procesi)
- HDFS data bypass (direktan pristup raw blokovima)

**Objašnjenje:**

YARN (Resource Manager) omogućava distribuirano izvršavanje poslova (MapReduce, Spark). Podrazumevana Hadoop instalacija ne zahteva autentifikaciju za slanje poslova na port 8088. Napadač može poslati maliciozni `.jar` fajl sa reverse shell payload-om, koji će biti izvršen na NodeManager čvoru, dajući napadaču potpunu kontrolu nad sistemom.

#### Napadi

**Scenario: Maliciozni MapReduce posao → Shell pristup**

**Lanac napada:**
Napadač skenira mrežu i pronalazi otvoren YARN port (8088) bez autentifikacije. Kreira maliciozni JAR fajl koji sadrži reverse shell kod, submituje ga kao legitiman MapReduce posao. YARN prihvata posao, pokreće ga na NodeManager čvoru, i napadač dobija shell pristup.

**Post-eksploatacija mogućnosti:**
- Direktan pristup HDFS blokovima (zaobilaženje HDFS permisija)
- Modifikacija zapisa o putovanjima i plaćanjima
- DoS napadi (brisanje kritičnih podataka)
- Lateralno kretanje ka drugim servisima (sniffing komunikacije)

**Uticaj na Smart Mobility:**
- Krađa kompletne baze podataka o putovanjima
- Falsifikovanje zapisa o plaćanjima → računovodstvena prevara
- Potpuno zaustavljanje analytics pipeline-a

#### Mitigacije


1. **Kerberos autentifikacija za YARN:** Omogućiti Kerberos za ResourceManager kako bi samo autentifikovani korisnici mogli slati poslove

2. **YARN ACL (Access Control Lists):** Konfigurisati whitelist korisnika/grupa koji smeju slati poslove (npr. samo `analytics-team` grupa)

3. **Izolacija kontejnera:** LinuxContainerExecutor sa zabranjenim privilegovanim korisnicima, resource limiti po job-u, i SELinux/AppArmor politike

4. **Mrežna segmentacija:** Hadoop klaster u internal-only mreži - YARN port (8088) blokiran van klastera

---

### 3. WebHDFS User Impersonation

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- HDFS podaci o putovanjima drugih korisnika
- Istorija plaćanja i transakcioni zapisi
- Zaobilaženje autorizacije (bypass HDFS ACL-ova)

**Tehnička pozadina:**

WebHDFS je REST API za HDFS pristup preko HTTP-a. U nesigurnim Hadoop klasterima (bez Kerberos-a), WebHDFS implicitno veruje `user.name` parametru u URL-u. Pošto je ovaj parametar pod kontrolom klijenta, napadač može lažirati bilo koje korisničko ime i pristupiti podacima drugih korisnika - sistem ne proverava da li je klijent zaista taj korisnik.

#### Napadi

**Scenario: Lažiranje korisničkog identiteta → Krađa podataka**

U nekoj implementaciji moguće je da Smart Mobility sistem organizuje podatke po korisnicima (`/data/users/marko/travels/`, `/data/users/ana/travels/`, itd.). HDFS ACL-ovi štite da samo vlasnik može pristupiti svom direktorijumu.

**Lanac napada:**

Napadač šalje HTTP zahtev ka WebHDFS endpointu i postavlja `user.name=hdfs` parametar, lažirajući identitet HDFS administratora. WebHDFS prihvata ovaj parametar bez provere i vraća listu svih korisničkih direktorijuma.

Napadač zatim može:
- **Enumerisati sve korisnike** (lista svih user_xxx direktorijuma)
- **Čitati osetljive podatke** (GPS koordinate, obrasci putovanja)
- **Masovno prikupljanje** (automatizovani script koji downloaduje podatke svih korisnika)
- **Menjati podatke** (dodavanje lažnih zapisa o putovanjima)

**Posledice:** Krađa podataka, izloženost obrazaca kretanja, računovodstvena korupcija.

#### Mitigacije


1. **Kerberos autentifikacija:** Konfigurisati WebHDFS da zahteva Kerberos token u Authorization header-u. Parametar `user.name` se ignoriše - identitet dolazi iz kriptografski potpisanog tokena.

2. **Delegation Tokens:** Za aplikacije koje ne mogu koristiti Kerberos direktno - koristiti delegation tokene koje potpisuje NameNode (ne mogu se falsifikovati kao obični `user.name` parametar).

3. **Isključiti WebHDFS (ako nije potreban):** Ako Smart Mobility aplikacija ne koristi WebHDFS direktno, potpuno ga isključiti iz konfiguracije - eliminacija attack surface-a.

4. **Firewall pravila:** Blokirati WebHDFS port (9870) van internal mreže - samo local servisi mogu pristupiti.

5. **API Gateway middleware:** Rutiranje kroz middleware koji validira JWT/OAuth tokene + ACL proveru PRE forwarding-a ka WebHDFS-u.

---

### 4. HDFS Permission Misconfiguration

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- HDFS direktorijumi dostupni za čitanje svim korisnicima
- Sirovi podaci o putovanjima (2M+ zapisa), istorija plaćanja, poslovna analitika

**Tehnička pozadina:**

HDFS ima Linux-style permission model sa owner/group/others permisijama. Problem nastaje kada developeri kreiraju direktorijume bez eksplicitnih permisija - sistem primenjuje podrazumevane (755) koje dozvoljavaju čitanje svim korisnicima.

**Ranjivost:** Kerberos je samo autentifikacija. Autorizacija zavisi od HDFS permisija i ACL-ova. Direktorijumi dostupni svima = nema prave autorizacije.

#### Napadi

**Scenario 1: Rudarenje podataka sa keytab-om niskih privilegija**

Smart Mobility sistem ima kompletnu HDFS strukturu sa direktorijumima za putovanja, plaćanja i analitiku - svi su konfigurisani sa permisijama koje dozvoljavaju čitanje svim autentifikovanim korisnicima.

**Lanac napada:**

Napadač kompromituje keytab niskih privilegija (npr. `tester.keytab` ili `app_user.keytab`) - možda iz prvog napada (Keytab Leak). Autentifikuje se sa tim credentials-ima i pošto su direktorijumi dostupni za čitanje svima, može:

- **Čitati sve podatke o putovanjima** - 2 miliona zapisa sa GPS lokacijama i korisničkim ID-evima
- **Ukrasti poslovnu analitiku** - najpopularnije rute, cenovni modeli, predviđanja potražnje

**Impact:**
- Milion+ korisničkih zapisa sa GPS podacima
- Konkurentska prednost izgubljena (analitika ukradena)
- Sve to sa keytab-om niskih privilegija

**Scenario 2: Insider pretnja**

Data scientist sa legitimnim `analyst` principal-om ima zadatak da radi samo sa agregatnim analitikom. Međutim, zbog permisija koje dozvoljavaju čitanje svima, može pristupiti i sirovim GPS podacima korisnika, iako to nije u okviru njegovih ovlaštećenja.

**Motivacija:** Prodaja podataka konkurenciji, praćenje pojedinih korisnika (stalking), sabotaža pre napuštanja kompanije.

#### Mitigacije


1. **Restriktivne podrazumevane permisije:** Direktorijumi moraju biti kreirani sa eksplicitnim permisijama koje dozvoljavaju pristup samo vlasniku (700) ili timu (750 sa group-based pristupom).

2. **HDFS ACL-ovi (Fine-Grained kontrola):** Omogućiti ACL support i konfigurisati whitelist specifičnih korisnika/grupa koji smeju pristupiti podacima. Čak i ako je keytab leak-ovan, pristup zavisi od ACL whitelist-a.

3. **Pathname-Based autorizacija:** Omogućiti proveru permisija i ACL support u HDFS konfiguraciji - bez toga, permisije se ne primenjuju.

4. **Podrazumevana umask konfiguracija:** Konfigurisati sistem da automatski kreira nove fajlove/direktorijume sa restriktivnim permisijama (700) umesto permisivnih (755).

---

### 5. Denial of Service - DataNode Resource Exhaustion

#### Pretnje (Resursi)

**Resursi pod rizikom:**
- Dostupnost Hadoop klastera
- Real-time GPS processing pipeline
- Poslovne operacije (izveštaji, dashboards)

**Tehnička pozadina:**

HDFS čuva podatke u blokovima (128MB podrazumevano) distribuirano na DataNode worker-ima. Ranjivosti nastaju jer sistem podrazumevano nema:
- **Ograničenja disk prostora** - napadač može popuniti ceo storage
- **Ograničenja replikacije** - 1GB upload = 3GB storage zbog 3x replika
- **Ograničenja konekcija** - DataNode može servirati ~4096 simultanih konekcija
- **Ograničenja resursa za YARN** - malicious MapReduce job može zauzeti sve CPU/memoriju

#### Napadi

**Scenario 1: Iscrpljivanje disk prostora**

Napadač generiše ogromne količine garbage podataka i uploaduje ih u HDFS sa faktorom replikacije 3. Kada napadač uploaduje 5TB podataka, čitav storage konzumira 15TB (zbog 3 replike). 

**Impact:** NameNode prelazi u safe mode (read-only), legitimni write operacije failuju, analytics pipeline se zaustavlja.

**Scenario 2: Iscrpljivanje konekcija**

Napadač otvara hiljade simultanih HDFS konekcija automatizovanom skriptom. DataNode dostiže limit maksimalnih transfer threads-a, pa svi legitimni klijenti dobijaju IOException.

**Impact:** Smart Mobility frontend ne može učitati dashboards, background analytics job-ovi failuju, kompletan klaster nedostupan 30+ minuta.

**Scenario 3: Maliciozni YARN posao**

Napadač submituje MapReduce job sa infinite CPU petljom i memory leak-om. Job zauzima sve dostupne YARN kontejnere, legitimni analytics job-ovi ne mogu dobiti resurse, CPU klaster 100%.

**Impact na Smart Mobility:**
- GPS tracking prestaje (gubitak real-time podataka)
- Customer dashboards timeout-uju (SLA kršeće)

#### Mitigacije


1. **HDFS disk kvote:** Postaviti space quota po korisniku (npr. 100GB limit) i namespace quota (max broj fajlova). Jedan korisnik ne može iscrpsti ceo klaster.

2. **DataNode resource limiti:** Povećati max broj transfer threads-a (npr. 8192) i rezervisati minimalnu količinu slobodnog prostora (npr. 10GB).

3. **YARN kvote:** Konfigurisati capacity scheduler da user-submitted job-ovi mogu zauzeti max 30% resursa, a legitimna analitika 60%. Takođe ograničiti broj job-ova po korisniku (npr. max 10).

4. **Monitoring & Alerting:**
   - Alert kada disk usage pređe 80%
   - Alert na neobičane upload rate-ove (>1TB/sat)
   - Alert kada job radi >24h bez progress-a

5. **Rate Limiting:** API Gateway middleware za HDFS upload koji ograničava upload na max 10GB/sat po korisniku.

6. **Politika ubijanja job-ova:** Automatski kill job-ova koji su stariji od 48h bez progress-a.

---

## Zaključak

Hadoop u Smart Mobility platformi čuva milione zapisa o putovanjima, GPS koordinate. Pet pretnji ugrožavaju:

**1. Krađa podataka (Data Breach):**
- Kerberos Keytab Leak (demonstrirano u GitHub repozitorijumu)
- WebHDFS lažiranje identiteta
- HDFS permisije dostupne svima

**2. Kompromitovanje sistema:**
- YARN Remote Code Execution

**3. Uskraćivanje usluge (Service Disruption):**
- DataNode iscrpljivanje resursa

**Minimalna odbrana (Defense-in-Depth):**
1. **Kerberos** za autentifikaciju (sa restriktivnim keytab permisijama)
2. **HDFS ACL-ovi** za precizan pristup podacima
3. **YARN ACL-ovi & kvote** za izolaciju resursa
4. **Mrežna segmentacija** (Hadoop samo u internal mreži)
5. **Monitoring & Audit Logging** (proaktivna detekcija)

