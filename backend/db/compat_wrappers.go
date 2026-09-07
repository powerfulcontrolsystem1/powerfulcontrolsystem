package db

import (
    "database/sql"
    "sync"
)

var (
    defaultDB   *sql.DB
    defaultDBMu sync.RWMutex
)

// SetDefaultDB registra una conexión DB que será retornada por GetDB.
func SetDefaultDB(dbConn *sql.DB) {
    defaultDBMu.Lock()
    defer defaultDBMu.Unlock()
    defaultDB = dbConn
}

// GetDB devuelve la conexión DB registrada previamente con SetDefaultDB.
func GetDB() *sql.DB {
    defaultDBMu.RLock()
    if defaultDB != nil {
        defer defaultDBMu.RUnlock()
        return defaultDB
    }
    defaultDBMu.RUnlock()
    return nil
}

// GetDatabaseType expone el dialecto SQL actual.
func GetDatabaseType() string {
    return currentSQLDialect()
}

// ExecCompat ejecuta una sentencia SQL respetando las adaptaciones de compatibilidad
// definidas en sql_compat.go.
func ExecCompat(dbConn *sql.DB, query string, args ...interface{}) (sql.Result, error) {
    if dbConn == nil {
        dbConn = GetDB()
    }
    return execSQLCompat(dbConn, query, args...)
}

// ExecQueryCompat ejecuta una consulta que devuelve filas usando la capa de compatibilidad vigente.
func ExecQueryCompat(dbConn *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
    if dbConn == nil {
        dbConn = GetDB()
    }
    return querySQLCompat(dbConn, query, args...)
}
