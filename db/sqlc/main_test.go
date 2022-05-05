package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/akshay-dobariya/simple-bank/util"
	_ "github.com/lib/pq"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	var err error
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("cannot load configurations:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatalf("Cannot connect to the DB error: '%s'", err)
	}
	testQueries = New(testDB)
	os.Exit(m.Run())
}
