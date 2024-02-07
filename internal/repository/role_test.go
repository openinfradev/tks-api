package repository_test

import (
	"encoding/json"
	"github.com/openinfradev/tks-api/internal/repository"
	"testing"
)

func TestRole(t *testing.T) {
	db := db_connection()

	db.AutoMigrate(&repository.Role{}, &repository.Organization{}, &repository.Project{}, &repository.User{}, &repository.TksRole{}, &repository.ProjectRole{})

	repo := repository.NewRoleRepository(db)

	roles, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.MarshalIndent(roles, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("start")
	t.Logf("role: %s", string(out))
	t.Log("end")

}
