db = db.getSiblingDB(process.env.MONGODB_DB || "auctions");
var now = new Date();

db.createCollection("users");
db.createCollection("auctions");
db.createCollection("bids");

db.auctions.createIndex({ status: 1, category: 1 });
db.auctions.createIndex({ product_name: "text" });

db.bids.createIndex({ auction_id: 1, amount: -1 });

db.users.insertMany([
    {
        _id: "adbf0c24-21bb-442c-8f7a-1d4b6826720f",
        name: "John Doe",
    },
    {
        _id: "89bee1c2-a7c2-4dff-8563-79c6c8f4477a",
        name: "Jane Doe",
    },
    {
        _id: "75b2cef5-1ae4-4004-8efc-c9931ae35039",
        name: "Alice Doe",
    },
]);

db.auctions.insertMany([
    {
        _id: "d535fac9-f89a-4e37-9303-bc2720cdc16c",
        product_name: "Volkswagen Golf 2021",
        category: "Automotive",
        description: "Motor 1.4 TSI",
        condition: 1,
        status: 0,
        timestamp: Math.floor(new Date().getTime() / 1000), // Unix timestamp in seconds (not milliseconds)
    },
    {
        _id: "fd58f3b5-8b4e-401a-b1d9-113f843c02c1",
        product_name: "Ford Mustang 2023",
        category: "Automotive",
        description: "Motor 5.0 V8",
        condition: 1,
        status: 0,
        timestamp: Math.floor(new Date().setMinutes(now.getMinutes() - 10) / 1000),
    },
    {
        _id: "a4f1b6d9-3e3b-4c2d-8b1a-6e1b3d0f1e1e",
        product_name: "Tesla Model S 2022",
        category: "Automotive",
        description: "Motor Electrico",
        condition: 1,
        status: 0,
        timestamp: Math.floor(new Date().setMinutes(now.getMinutes() + 1) / 1000),
    }
]);