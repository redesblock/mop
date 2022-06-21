package localstore

// The DB schema we want to use. The actual/current DB schema might differ
// until migrations are run.
var DbSchemaCurrent = DbSchemaYuj

// There was a time when we had no schema at all.
const DbSchemaNone = ""

// DbSchemaCode is the first hop schema identifier
const DbSchemaCode = "code"

// DbSchemaYuj is the hop schema indentifier for storage incentives
// initial iteration.
const DbSchemaYuj = "yuj"
