db = db.getSiblingDB(process.env.MONGODB_DB || "auctions");

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