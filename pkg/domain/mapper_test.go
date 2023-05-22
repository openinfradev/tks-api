package domain

import (
	"fmt"
	"testing"
	"time"
)

// test case
func TestConvert(t *testing.T) {
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
			name: "test case: CreateOrganizationRequest->Organization",
			args: args{
				src: CreateOrganizationRequest{
					Name:        "test",
					Description: "test",
					Phone:       "test",
				},
				dst: &Organization{},
			},
			wantErr: false,
		},
		{
			name: "test case Organization->CreateOrganizationResponse",
			args: args{
				src: Organization{
					ID:          "",
					Name:        "test",
					Description: "test",
					Phone:       "test",
					StatusDesc:  "good",
					Creator:     "",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
				},
				dst: &CreateOrganizationResponse{},
			},
			wantErr: false,
		},
		{
			name: "test case Organization->GetOrganizationResponse",
			args: args{
				src: Organization{
					ID:          "",
					Name:        "test",
					Description: "test",
					Phone:       "test",
					Status:      OrganizationStatus_CREATE,
					StatusDesc:  "good",
					Creator:     "",
					CreatedAt:   time.Time{},
					UpdatedAt:   time.Time{},
				},
				dst: &(&GetOrganizationResponse{}).Organization,
			},
			wantErr: false,
		},
		{
			name: "test case CreateUserRequest->User",
			args: args{
				src: CreateUserRequest{
					AccountId:   "testAccount",
					Password:    "testPassword",
					Name:        "testName",
					Email:       "testEmail",
					Department:  "testDepartment",
					Role:        "testRole",
					Description: "testDescription",
				},
				dst: &User{},
			},
			wantErr: false,
		},
		{
			name: "test case User->GetUserResponse",
			args: args{
				src: User{
					ID:           "",
					AccountId:    "testAccount",
					Password:     "testPassword",
					Name:         "testName",
					Token:        "testToken",
					Role:         Role{},
					Organization: Organization{},
					Creator:      "",
					CreatedAt:    time.Time{},
					UpdatedAt:    time.Time{},
					Email:        "",
					Department:   "",
					Description:  "",
				},
				dst: &GetUserResponse{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Map(tt.args.src, tt.args.dst); (err != nil) != tt.wantErr {
				t.Errorf("Map() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				fmt.Printf("Input: %+v\n", tt.args.src)
				fmt.Printf("Output: %+v\n\n", tt.args.dst)
			}
		})
	}
}
