package localstore

// DBSchemaCode is the first hop schema identifier.
const DBSchemaCode = "code"

// DBSchemaCurrent represents the DB schema we want to use.
// The actual/current DB schema might differ until migrations are run.
var DBSchemaCurrent = DBSchemaBatchIndex
