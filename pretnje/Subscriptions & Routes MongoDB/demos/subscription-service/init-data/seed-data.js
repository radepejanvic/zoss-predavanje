db = db.getSiblingDB('subscriptions_db');

db.users.insertMany([
  {
    user_id: "USR001",
    personal_info: {
      first_name: "Marko",
      last_name: "Jovanović",
      email: "marko.jovanovic@email.com",
      phone: "+38164123456",
      address: "Knez Mihailova 10, Beograd",
      jmbg: "1234567890123"
    },
    created_at: new Date("2024-01-15"),
    verified: true
  },
  {
    user_id: "USR002",
    personal_info: {
      first_name: "Jovana",
      last_name: "Petrović",
      email: "jovana.petrovic@email.com",
      phone: "+38164234567",
      address: "Bulevar Oslobođenja 50, Novi Sad",
      jmbg: "2345678901234"
    },
    created_at: new Date("2024-02-20"),
    verified: true
  },
  {
    user_id: "USR003",
    personal_info: {
      first_name: "Nikola",
      last_name: "Đorđević",
      email: "nikola.djordjevic@email.com",
      phone: "+38164345678",
      address: "Strossmayerova 15, Niš",
      jmbg: "3456789012345"
    },
    created_at: new Date("2024-03-10"),
    verified: false
  },
  {
    user_id: "USR004",
    personal_info: {
      first_name: "Ana",
      last_name: "Stojanović",
      email: "ana.stojanovic@email.com",
      phone: "+38164456789",
      address: "Cara Dušana 25, Kragujevac",
      jmbg: "4567890123456"
    },
    created_at: new Date("2024-03-15"),
    verified: true
  }
]);

db.transport_subscriptions.insertMany([
  {
    subscription_id: "SUB001",
    user_id: "USR001",
    type: "city_transport",
    plan: "monthly_unlimited",
    price: 2500,
    currency: "RSD",
    valid_from: new Date("2024-04-01"),
    valid_until: new Date("2024-04-30"),
    payment_method: "credit_card",
    payment_info: {
      card_last4: "1234",
      card_type: "Visa",
      transaction_id: "TXN123456789"
    },
    zones: ["zone_a", "zone_b"],
    status: "active",
    auto_renew: true
  },
  {
    subscription_id: "SUB002",
    user_id: "USR002",
    type: "city_transport",
    plan: "annual_unlimited",
    price: 24000,
    currency: "RSD",
    valid_from: new Date("2024-01-01"),
    valid_until: new Date("2024-12-31"),
    payment_method: "bank_transfer",
    payment_info: {
      bank_account: "160-123456-78",
      reference_number: "REF2024001"
    },
    zones: ["zone_a", "zone_b", "zone_c"],
    status: "active",
    auto_renew: false
  },
  
  {
    subscription_id: "SUB003",
    user_id: "USR003",
    type: "taxi_credits",
    plan: "taxi_starter",
    credits: 500,
    price: 1500,
    currency: "RSD",
    valid_from: new Date("2024-03-15"),
    valid_until: new Date("2024-06-15"),
    payment_method: "credit_card",
    payment_info: {
      card_last4: "5678",
      card_type: "Mastercard",
      transaction_id: "TXN987654321"
    },
    remaining_credits: 350,
    status: "active",
    auto_renew: false
  },
  {
    subscription_id: "SUB004",
    user_id: "USR001",
    type: "taxi_credits",
    plan: "taxi_premium",
    credits: 2000,
    price: 5500,
    currency: "RSD",
    valid_from: new Date("2024-04-01"),
    valid_until: new Date("2024-07-01"),
    payment_method: "paypal",
    payment_info: {
      paypal_email: "marko.jovanovic@email.com",
      transaction_id: "PPAY123456"
    },
    remaining_credits: 2000,
    status: "active",
    auto_renew: true
  },
  
  {
    subscription_id: "SUB005",
    user_id: "USR002",
    type: "combined",
    plan: "mobility_plus",
    price: 6500,
    currency: "RSD",
    valid_from: new Date("2024-03-01"),
    valid_until: new Date("2024-04-01"),
    payment_method: "credit_card",
    payment_info: {
      card_last4: "9012",
      card_type: "Visa",
      transaction_id: "TXN11223344"
    },
    benefits: {
      city_transport: {
        type: "monthly_unlimited",
        zones: ["zone_a", "zone_b"]
      },
      taxi_credits: 300,
      bike_sharing: "unlimited_30min"
    },
    remaining_taxi_credits: 200,
    status: "expired",
    auto_renew: false
  },
  {
    subscription_id: "SUB006",
    user_id: "USR004",
    type: "taxi_credits",
    plan: "taxi_business",
    credits: 5000,
    price: 12000,
    currency: "RSD",
    valid_from: new Date("2024-04-01"),
    valid_until: new Date("2024-10-01"),
    payment_method: "company_account",
    payment_info: {
      company_name: "Transport DOO",
      company_id: "12345678",
      invoice_number: "INV-2024-001"
    },
    remaining_credits: 5000,
    status: "active",
    auto_renew: false
  }
]);

db.payment_history.insertMany([
  {
    payment_id: "PAY001",
    subscription_id: "SUB001",
    user_id: "USR001",
    amount: 2500,
    currency: "RSD",
    payment_date: new Date("2024-04-01"),
    payment_method: "credit_card",
    card_last4: "1234",
    status: "completed"
  },
  {
    payment_id: "PAY002",
    subscription_id: "SUB002",
    user_id: "USR002",
    amount: 24000,
    currency: "RSD",
    payment_date: new Date("2024-01-01"),
    payment_method: "bank_transfer",
    status: "completed"
  },
  {
    payment_id: "PAY003",
    subscription_id: "SUB003",
    user_id: "USR003",
    amount: 1500,
    currency: "RSD",
    payment_date: new Date("2024-03-15"),
    payment_method: "credit_card",
    card_last4: "5678",
    status: "completed"
  }
]);

db.transport_zones.insertMany([
  {
    zone_id: "zone_a",
    name: "Centralna zona",
    cities: ["Beograd", "Novi Sad", "Niš"],
    price_multiplier: 1.0
  },
  {
    zone_id: "zone_b",
    name: "Prigradska zona",
    cities: ["Beograd", "Novi Sad", "Niš", "Kragujevac"],
    price_multiplier: 1.5
  },
  {
    zone_id: "zone_c",
    name: "Međugradska zona",
    cities: ["Srbija - svi gradovi"],
    price_multiplier: 2.0
  }
]);

db.users.createIndex({ "user_id": 1 });
db.transport_subscriptions.createIndex({ "user_id": 1 });
db.transport_subscriptions.createIndex({ "status": 1 });
db.payment_history.createIndex({ "user_id": 1 });
db.payment_history.createIndex({ "subscription_id": 1 });

print("[INFO]: Inserted demo data successfully!");
print(`[INFO]: Inserted ${db.users.count()} users!`);
print(`[INFO]: Inserted ${db.transport_subscriptions.count()} subscriptions!`);
print(`[INFO]: Inserted ${db.payment_history.count()} transactions!`); 
print(`[INFO]: Inserted ${db.transport_zones.count()} zones!`); 
