db = db.getSiblingDB('routes_db');

db.transport_routes.insertMany([

  {
    route_id: "RT001",
    type: "bus",
    line_number: "23",
    name: "Zemun - Slavija",
    operator: "GSP",
    stops: [
      "Zemun",
      "Tošin Bunar",
      "Brankov Most",
      "Zeleni Venac",
      "Slavija"
    ],
    zones: ["zone_a", "zone_b"],
    active: true
  },

  {
    route_id: "RT002",
    type: "bus",
    line_number: "45",
    name: "Novi Beograd - Kalemegdan",
    operator: "GSP",
    stops: [
      "Blok 70",
      "Novi Beograd",
      "Brankov Most",
      "Studentski Trg",
      "Kalemegdan"
    ],
    zones: ["zone_a"],
    active: true
  },

  {
    route_id: "RT003",
    type: "tram",
    line_number: "7",
    name: "Ustanička - Blok 45",
    operator: "GSP",
    stops: [
      "Ustanička",
      "Vukov Spomenik",
      "Slavija",
      "Most na Adi",
      "Blok 45"
    ],
    zones: ["zone_a", "zone_b"],
    active: true
  },

  {
    route_id: "RT004",
    type: "bus",
    line_number: "95",
    name: "Borča - Novi Beograd",
    operator: "Privatni Prevoznik",
    stops: [
      "Borča",
      "Pančevački Most",
      "Tašmajdan",
      "Savski Trg",
      "Novi Beograd"
    ],
    zones: ["zone_b"],
    active: true
  },

  {
    route_id: "RT005",
    type: "minibus",
    line_number: "E6",
    name: "Petlovo Brdo - Zeleni Venac",
    operator: "Privatni Prevoznik",
    stops: [
      "Petlovo Brdo",
      "Vidikovac",
      "Banovo Brdo",
      "Ada",
      "Zeleni Venac"
    ],
    zones: ["zone_a"],
    active: true
  }

]);

db.transport_routes.createIndex({ route_id: 1 });
db.transport_routes.createIndex({ line_number: 1 });

print("[INFO]: Transport routes inserted!");
print(`[INFO]: Inserted ${db.transport_routes.count()} routes!`);