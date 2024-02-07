package repository_test

import (
	"encoding/json"
	"fmt"
	"github.com/openinfradev/tks-api/internal/helper"
	"github.com/openinfradev/tks-api/internal/repository"
	"github.com/openinfradev/tks-api/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm_test/config"
	"testing"
)

func TestPermission(t *testing.T) {
	db := db_connection()

	db.AutoMigrate(&repository.Permission{})

	//model := domain.Permission{
	//	Name: "대시보드",
	//}

	repo := repository.NewPermissionRepository(db)

	permissions, err := repo.List()
	if err != nil {
		t.Fatal(err)
	}
	out, err := json.MarshalIndent(permissions, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	t.Log("start")
	t.Logf("permission: %s", string(out))
	t.Log("end")

	//t.Logf("permission: %+v", permissions)

	//
	//for _, permission := range permissions {
	//	// encoding to json
	//	b, err := json.Marshal(permission)
	//	if err != nil {
	//		t.Fatal(err)
	//	}
	//	t.Logf("permission: %s", string(b))
	//}
	//
	////create

	//if err := repo.Create(dashboard); err != nil {
	//	t.Fatal(err)
	//}
	//if err := repo.Create(stack); err != nil {
	//	t.Fatal(err)
	//}
	//if err := repo.Create(security_policy); err != nil {
	//	t.Fatal(err)
	//}
	//if err := repo.Create(projectManagement); err != nil {
	//	t.Fatal(err)
	//}
	//if err := repo.Create(notification); err != nil {
	//	t.Fatal(err)
	//}
	//if err := repo.Create(configuration); err != nil {
	//	t.Fatal(err)
	//}

	// get
	//permission, err := repo.Get(uuid.MustParse("fd4363a7-d1d2-4feb-b976-b87d99a775c4"))
	//if err != nil {
	//	t.Fatal(err)
	//}
	//out, err := json.Marshal(permission)
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("permission: %s", string(out))

	// print json pretty
	//out, err := json.MarshalIndent(permission, "", "  ")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//t.Logf("permission: %s", string(out))

}

func db_connection() *gorm.DB {
	conf := config.NewDefaultConfig()
	dsn := fmt.Sprintf(
		"host=%s dbname=%s user=%s password=%s port=%d sslmode=disable TimeZone=Asia/Seoul",
		conf.Address, conf.Database, conf.AdminId, conf.AdminPassword, conf.Port,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	return db
}

var dashboard = &domain.Permission{
	Name: "대시보드",
	Children: []*domain.Permission{
		{
			Name: "대시보드",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "대시보드 설정",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
	},
}

var stack = &domain.Permission{
	Name: "스택 관리",
	Children: []*domain.Permission{
		{
			Name:      "조회",
			IsAllowed: helper.BoolP(false),
		},
		{
			Name:      "생성",
			IsAllowed: helper.BoolP(false),
		},
		{
			Name:      "수정",
			IsAllowed: helper.BoolP(false),
		},
		{
			Name:      "삭제",
			IsAllowed: helper.BoolP(false),
		},
	},
}

var security_policy = &domain.Permission{
	Name: "보안/정책 관리",
	Children: []*domain.Permission{
		{
			Name: "보안/정책",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
	},
}

var projectManagement = &domain.Permission{
	Name: "프로젝트 관리",
	Children: []*domain.Permission{
		{
			Name: "프로젝트",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "앱 서빙",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "빌드",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "배포",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "설정-일반",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "설정-멤버",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "설정-네임스페이스",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
	},
}

var notification = &domain.Permission{
	Name: "알림",
	Children: []*domain.Permission{
		{
			Name: "시스템 경고",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "보안/정책 감사로그",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
	},
}

var configuration = &domain.Permission{
	Name: "설정",
	Children: []*domain.Permission{
		{
			Name: "일반",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "클라우드 계정",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "스택 템플릿",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "프로젝트 관리",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "사용자",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "사용자 권한 관리",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
		{
			Name: "알림 설정",
			Children: []*domain.Permission{
				{
					Name:      "조회",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "생성",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "수정",
					IsAllowed: helper.BoolP(false),
				},
				{
					Name:      "삭제",
					IsAllowed: helper.BoolP(false),
				},
			},
		},
	},
}
