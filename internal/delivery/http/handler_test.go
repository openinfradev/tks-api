package http

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/openinfradev/tks-api/internal/database"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func init() {
	// log.Disable()
}

func TestMain(m *testing.M) {
	pool, resource, err := database.CreatePostgres()
	if err != nil {
		fmt.Printf("Could not create postgres: %s", err)
		os.Exit(-1)
	}
	testDBHost, testDBPort := database.GetHostAndPort(resource)

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Seoul",
		testDBHost, "postgres", "password", "tks", testDBPort)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf(context.TODO(), "Failed to open gorm: %s", err)
	}

	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	err = database.MigrateSchema(db)
	if err != nil {
		log.Fatalf(context.TODO(), "Could not migrate postgres: %s", err)
	}
	err = database.EnsureDefaultRows(db)
	if err != nil {
		log.Fatalf(context.TODO(), "cannot Initializing Default Rows in Database: %s", err)
	}

	code := m.Run()

	if err := database.RemovePostgres(pool, resource); err != nil {
		log.Fatalf(context.TODO(), "Could not remove postgres: %s", err)
		os.Exit(-1)
	}

	os.Exit(code)
}
