package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/internal/model"
	"github.com/openinfradev/tks-api/internal/pagination"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"github.com/openinfradev/tks-api/pkg/log"
	"gorm.io/gorm"
)

// Interface
type IUserRepository interface {
	CreateWithUuid(ctx context.Context, uuid uuid.UUID, accountId string, name string, email string,
		department string, description string, organizationId string, roleId string) (model.User, error)
	List(ctx context.Context, filters ...FilterFunc) (out *[]model.User, err error)
	ListWithPagination(ctx context.Context, pg *pagination.Pagination, organizationId string) (out *[]model.User, err error)
	Get(ctx context.Context, accountId string, organizationId string) (model.User, error)
	GetByUuid(ctx context.Context, userId uuid.UUID) (model.User, error)
	UpdateWithUuid(ctx context.Context, uuid uuid.UUID, accountId string, name string, roleId string, email string,
		department string, description string) (model.User, error)
	UpdatePasswordAt(ctx context.Context, userId uuid.UUID, organizationId string, isTemporary bool) error
	DeleteWithUuid(ctx context.Context, uuid uuid.UUID) error
	Flush(ctx context.Context, organizationId string) error

	AccountIdFilter(accountId string) FilterFunc
	OrganizationFilter(organization string) FilterFunc
	EmailFilter(email string) FilterFunc
	NameFilter(name string) FilterFunc
}

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) Flush(ctx context.Context, organizationId string) error {
	res := r.db.WithContext(ctx).Where("organization_id = ?", organizationId).Delete(&model.User{})
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateWithUuid(ctx context.Context, uuid uuid.UUID, accountId string, name string, email string,
	department string, description string, organizationId string, roleId string) (model.User, error) {

	newUser := model.User{
		ID:                uuid,
		AccountId:         accountId,
		Name:              name,
		Email:             email,
		Department:        department,
		Description:       description,
		OrganizationId:    organizationId,
		RoleId:            roleId,
		PasswordUpdatedAt: time.Now(),
	}
	res := r.db.WithContext(ctx).Create(&newUser)
	if res.Error != nil {
		log.Error(ctx, res.Error.Error())
		return model.User{}, res.Error
	}
	user, err := r.getUserByAccountId(ctx, accountId, organizationId)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) List(ctx context.Context, filters ...FilterFunc) (*[]model.User, error) {
	var users []model.User
	var res *gorm.DB

	if filters == nil {
		res = r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Role").Find(&users)
	} else {
		combinedFilter := func(filters ...FilterFunc) FilterFunc {
			return func(user *gorm.DB) *gorm.DB {
				for _, f := range filters {
					user = f(user)
				}
				return user
			}
		}
		cFunc := combinedFilter(filters...)
		res = cFunc(r.db.Model(&model.User{}).Preload("Organization").Preload("Role")).Find(&users)
	}

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}

	var out []model.User
	out = append(out, users...)

	return &out, nil
}

func (r *UserRepository) ListWithPagination(ctx context.Context, pg *pagination.Pagination, organizationId string) (*[]model.User, error) {
	var users []model.User

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload("Organization").Preload("Role").Model(&model.User{}).Where("users.organization_id = ?", organizationId), &users)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}

	var out []model.User
	out = append(out, users...)

	return &out, nil
}

func (r *UserRepository) Get(ctx context.Context, accountId string, organizationId string) (model.User, error) {
	user, err := r.getUserByAccountId(ctx, accountId, organizationId)
	if err != nil {
		return model.User{}, err
	}

	return user, nil
}

func (r *UserRepository) GetByUuid(ctx context.Context, userId uuid.UUID) (respUser model.User, err error) {
	user := model.User{}
	res := r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Role").Find(&user, "id = ?", userId)

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return model.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}

	return user, nil
}
func (r *UserRepository) UpdateWithUuid(ctx context.Context, uuid uuid.UUID, accountId string, name string, roleId string,
	email string, department string, description string) (model.User, error) {
	var user model.User
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", uuid).Updates(model.User{
		AccountId:   accountId,
		Name:        name,
		Email:       email,
		Department:  department,
		Description: description,
		RoleId:      roleId,
	})
	if res.RowsAffected == 0 || res.Error != nil {
		return model.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.User{}, res.Error
	}
	res = r.db.Model(&model.User{}).Preload("Organization").Preload("Role").Where("id = ?", uuid).Find(&user)
	if res.Error != nil {
		return model.User{}, res.Error
	}
	return user, nil
}

func (r *UserRepository) UpdatePasswordAt(ctx context.Context, userId uuid.UUID, organizationId string, isTemporary bool) error {
	var updateUser = model.User{}
	if isTemporary {
		updateUser.PasswordUpdatedAt = time.Time{}
	} else {
		updateUser.PasswordUpdatedAt = time.Now()
	}
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ? AND organization_id = ?", userId, organizationId).
		Select("password_updated_at").Updates(updateUser)

	if res.RowsAffected == 0 || res.Error != nil {
		return httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}

	return nil
}

func (r *UserRepository) DeleteWithUuid(ctx context.Context, uuid uuid.UUID) error {
	res := r.db.WithContext(ctx).Unscoped().Delete(&model.User{}, "id = ?", uuid)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func (r *UserRepository) GetRoleByName(ctx context.Context, roleName string) (model.Role, error) {
	role, err := r.getRoleByName(ctx, roleName)
	if err != nil {
		return model.Role{}, err
	}

	return role, nil
}

//func (r *UserRepository) FetchRoles() (*[]model.Role, error) {
//	var roles []model.Role
//	res := r.db.Find(&roles)
//
//	if res.Error != nil {
//		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
//		return nil, res.Error
//	}
//
//	if res.RowsAffected == 0 {
//		return nil, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
//	}
//
//	var out []model.Role
//	for _, role := range roles {
//		outRole := r.reflectRole(role)
//		out = append(out, outRole)
//	}
//
//	return &out, nil
//}

// private members
func (r *UserRepository) getUserByAccountId(ctx context.Context, accountId string, organizationId string) (model.User, error) {
	user := model.User{}
	res := r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Role").
		Find(&user, "account_id = ? AND organization_id = ?", accountId, organizationId)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return model.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}

	return user, nil
}

func (r *UserRepository) getRoleByName(ctx context.Context, roleName string) (model.Role, error) {
	role := model.Role{}
	res := r.db.WithContext(ctx).First(&role, "name = ?", roleName)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.Role{}, res.Error
	}
	if res.RowsAffected == 0 {
		return model.Role{}, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}

	return role, nil
}

func (r *UserRepository) AccountIdFilter(accountId string) FilterFunc {
	return func(user *gorm.DB) *gorm.DB {
		return user.Where("account_id = ?", accountId)
	}
}

func (r *UserRepository) OrganizationFilter(organization string) FilterFunc {
	return func(user *gorm.DB) *gorm.DB {
		return user.Where("organization_id = ?", organization)
	}
}

func (r *UserRepository) EmailFilter(email string) FilterFunc {
	return func(user *gorm.DB) *gorm.DB {
		return user.Where("email = ?", email)
	}
}

func (r *UserRepository) NameFilter(name string) FilterFunc {
	return func(user *gorm.DB) *gorm.DB {
		return user.Where("name = ?", name)
	}
}
