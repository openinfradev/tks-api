package usecase_test

import (
	"fmt"
	"github.com/google/uuid"
	myRepository "github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/internal/usecase"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm_test/config"
	"testing"
)

func TestNewRoleUsecase(t *testing.T) {
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

	if err = db.AutoMigrate(&myRepository.Organization{}); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&myRepository.Role{}); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&myRepository.TksRole{}); err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(&myRepository.ProjectRole{}); err != nil {
		panic(err)
	}

	orgaizationRepository := myRepository.NewOrganizationRepository(db)
	roleDbClient := myRepository.NewRoleRepository(db)

	roleUsecase := usecase.NewRoleUsecase(roleDbClient)

	// create organization
	// organizationId = "test"
	if _, err := orgaizationRepository.Get("test"); err != nil {
		if _, err := orgaizationRepository.Create("test", "test", uuid.New(), "test", "test"); err != nil {
			panic(err)
		}
	}

	// create project
	// projectName = "testProject"
	projectRepository := myRepository.NewProjectRepository(db)
	projectFound := false
	if projects, err := projectRepository.List(); err != nil {
		panic(err)
	} else {
		for _, project := range projects {
			if project.Name == "testProject" {
				projectFound = true
				break
			}
		}
	}

	if !projectFound {
		project := &domain.Project{
			Name:           "testProject",
			OrganizationID: "test",
			Description:    "testDesc",
		}
		if err := projectRepository.Create(project); err != nil {
			panic(err)
		}
	}

	// get id of project
	projects, err := projectRepository.List()
	if err != nil {
		panic(err)
	}
	var projectId uuid.UUID
	for _, project := range projects {
		if project.Name == "testProject" {
			projectId = project.ID
			break
		}
	}

	// create tks role
	tksRoles, err := roleUsecase.ListTksRoles()
	if err != nil {
		panic(err)
	}
	tksRoleFound := false
	for _, role := range tksRoles {
		if role.Name == "testTksRole" {
			tksRoleFound = true
			break
		}
	}
	if !tksRoleFound {
		role := &domain.TksRole{
			Role: domain.Role{
				Name:           "testTksRole",
				OrganizationID: "test",
				Description:    "testDesc",
			},
		}
		if err := roleUsecase.CreateTksRole(role); err != nil {
			panic(err)
		}
	}

	// create project role
	projectRoles, err := roleUsecase.ListProjectRoles()
	if err != nil {
		panic(err)
	}

	projectRoleFound := false
	for _, role := range projectRoles {
		if role.Role.Name == "testProjectRole" {
			projectRoleFound = true
			break
		}
	}

	if !projectRoleFound {
		role := &domain.ProjectRole{
			Role: domain.Role{
				Name:           "testProjectRole",
				OrganizationID: "test",
				Description:    "testDesc",
			},
			ProjectID: projectId,
		}
		if err := roleUsecase.CreateProjectRole(role); err != nil {
			panic(err)
		}
	}

	// list tks role
	tksRoles, err = roleUsecase.ListTksRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range tksRoles {
		t.Logf("index: %d, tksRole: %+v", i, role)
	}

	// list project role
	projectRoles, err = roleUsecase.ListProjectRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range projectRoles {
		t.Logf("index: %d, projectRole: %+v", i, role)
	}

	// list all role
	roles, err := roleUsecase.ListRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range roles {
		t.Logf("index: %d, role: %+v", i, role)
	}

	t.Logf("now delete all roles")

	// delete tks role
	for _, role := range tksRoles {
		if err := roleDbClient.Delete(role.RoleID); err != nil {
			panic(err)
		}
	}

	// delete project role
	for _, role := range projectRoles {
		if err := roleDbClient.Delete(role.RoleID); err != nil {
			panic(err)
		}
	}

	// print tks role
	tksRoles, err = roleUsecase.ListTksRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range tksRoles {
		t.Logf("index: %d, tksRole: %+v", i, role)
	}

	// print project role
	projectRoles, err = roleUsecase.ListProjectRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range projectRoles {
		t.Logf("index: %d, projectRole: %+v", i, role)
	}

	// print all role
	roles, err = roleUsecase.ListRoles()
	if err != nil {
		panic(err)
	}
	for i, role := range roles {
		t.Logf("index: %d, role: %+v", i, role)
	}

}
