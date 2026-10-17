// Runs once on the first startup of the MongoDB container.
// The database selected here matches MONGO_INITDB_DATABASE.
db.createCollection("users");
db.createCollection("auth_sessions");
db.createCollection("submissions");
db.createCollection("share_links");
db.createCollection("share_views");
