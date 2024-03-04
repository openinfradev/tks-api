package repository_test

import (
	"fmt"
	myRepository "github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm_test/config"
	"testing"
)

func TestNewEndpointRepository(t *testing.T) {
	conf := config.NewDefaultConfig()
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=disable TimeZone=Asia/Seoul",
		conf.Address, conf.Database, conf.AdminId, conf.AdminPassword, conf.Port,
	)

	t.Logf("dsn: %s", dsn)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	if err = db.AutoMigrate(&myRepository.Endpoint{}); err != nil {
		panic(err)
	}

	dbClient := myRepository.NewEndpointRepository(db)

	// create
	endpoint := &domain.Endpoint{
		Name:  "test",
		Group: "test",
	}

	if err := dbClient.Create(endpoint); err != nil {
		panic(err)
	}

	t.Log("list")
	// list
	endpoints, err := dbClient.List(nil)
	if err != nil {
		panic(err)
	}

	for _, endpoint := range endpoints {
		t.Logf("endpoint: %+v", endpoint)
	}

	t.Log("get")
	// get
	for _, endpoint := range endpoints {
		endpoint, err := dbClient.Get(endpoint.ID)
		if err != nil {
			panic(err)
		}
		t.Logf("endpoint: %+v", endpoint)
	}

	t.Log("update")
	// update
	for _, endpoint := range endpoints {
		endpoint.Name = "test2"
		t.Logf("BeforeUpdate: %+v", endpoint)
		if err := dbClient.Update(endpoint); err != nil {
			panic(err)
		} else {
			t.Logf("AfterUpdate: %+v", endpoint)
		}
	}
}
