package repository

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/domain"
	"testing"
	"time"
)

// test case
func TestConvert(t *testing.T) {
	genUuid, _ := uuid.NewRandom()
	uuidStr := genUuid.String()

	type args struct {
		src interface{}
		dst interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test case 1",
			args: args{
				src: User{
					ID:             genUuid,
					AccountId:      "testAccount",
					Name:           "testName",
					Password:       "testPassword",
					RoleId:         uuid.UUID{},
					Role:           Role{},
					OrganizationId: "testOrganizationId",
					Organization:   Organization{},
					Creator:        uuid.UUID{},
					Email:          "testEmail",
					Department:     "testDepartment",
					Description:    "testDescription",
				},
				dst: &domain.User{},
			},
			wantErr: false,
		},
		{
			name: "test case 2",
			args: args{
				src: domain.User{
					ID:           uuidStr,
					AccountId:    "testAccount",
					Password:     "testPassword",
					Name:         "testName",
					Token:        "testToken",
					Role:         domain.Role{},
					Organization: domain.Organization{},
					Creator:      "testCreator",
					CreatedAt:    time.Time{},
					UpdatedAt:    time.Time{},
					Email:        "test	email",
					Department:   "testDepartment",
					Description:  "testDescription",
				},
				dst: &User{},
			},
			wantErr: false,
		},
	}
	_ = tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Map(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("Map() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Println(tt.args.dst)
			}
		})
	}
}
