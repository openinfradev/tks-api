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
	"gorm.io/gorm/clause"
)

// Interface
type IUserRepository interface {
	Create(ctx context.Context, user *model.User) (*model.User, error)
	List(ctx context.Context, filters ...FilterFunc) (out *[]model.User, err error)
	ListWithPagination(ctx context.Context, pg *pagination.Pagination, organizationId string) (out *[]model.User, err error)
	Get(ctx context.Context, accountId string, organizationId string) (model.User, error)
	GetByUuid(ctx context.Context, userId uuid.UUID) (model.User, error)
	Update(ctx context.Context, user *model.User) (*model.User, error)
	UpdatePasswordAt(ctx context.Context, userId uuid.UUID, organizationId string, isTemporary bool) error
	DeleteWithUuid(ctx context.Context, uuid uuid.UUID) error
	Flush(ctx context.Context, organizationId string) error

	ListUsersByRole(ctx context.Context, organizationId string, roleId string, pg *pagination.Pagination) (*[]model.User, error)

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

// func (r *UserRepository) CreateWithUuid(ctx context.Context, uuid uuid.UUID, accountId string, name string, email string,
//
//	department string, description string, organizationId string, roleId string) (model.User, error) {
func (r *UserRepository) Create(ctx context.Context, user *model.User) (*model.User, error) {
	user.PasswordUpdatedAt = time.Now()
	//newUser := model.User{
	//	ID:                uuid,
	//	AccountId:         accountId,
	//	Name:              name,
	//	Email:             email,
	//	Department:        department,
	//	Description:       description,
	//	OrganizationId:    organizationId,
	//	RoleId:            roleId,
	//	PasswordUpdatedAt: time.Now(),
	//}
	res := r.db.WithContext(ctx).Create(user)
	if res.Error != nil {
		log.Error(ctx, res.Error.Error())
		return nil, res.Error
	}
	resp, err := r.getUserByAccountId(ctx, user.AccountId, user.Organization.ID)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}

func (r *UserRepository) List(ctx context.Context, filters ...FilterFunc) (*[]model.User, error) {
	var users []model.User
	var res *gorm.DB

	if filters == nil {
		res = r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Roles").Find(&users)
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
		res = cFunc(r.db.Model(&model.User{}).Preload("Organization").Preload("Roles")).Find(&users)
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

	db := r.db.WithContext(ctx).Preload(clause.Associations).Model(&model.User{}).
		Where("users.organization_id = ?", organizationId)

	// [TODO] more pretty!
	for _, filter := range pg.Filters {
		if filter.Relation == "Roles" {
			db = db.Joins("join user_roles on user_roles.user_id = users.id").
				Joins("join roles on roles.id = user_roles.role_id").
				Where("roles.name ilike ?", "%"+filter.Values[0]+"%")
			break
		}
	}

	_, res := pg.Fetch(db, &users)
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
	res := r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Roles").Find(&user, "id = ?", userId)

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return model.User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return model.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound, "", "")
	}

	return user, nil
}

func (r *UserRepository) ListUsersByRole(ctx context.Context, organizationId string, roleId string, pg *pagination.Pagination) (*[]model.User, error) {
	var users []model.User

	if pg == nil {
		pg = pagination.NewPagination(nil)
	}

	_, res := pg.Fetch(r.db.WithContext(ctx).Preload("Organization").Preload("Roles").Model(&model.User{}).
		Where("users.organization_id = ? AND roles.id = ?", organizationId, roleId).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Joins("JOIN roles ON user_roles.role_id = roles.id"), &users)
	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}

	var out []model.User
	out = append(out, users...)

	return &out, nil
}

func (r *UserRepository) Update(ctx context.Context, user *model.User) (*model.User, error) {
	res := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", user.ID).
		Select("Name", "Email", "Department", "Description").Updates(model.User{
		Name:        user.Name,
		Email:       user.Email,
		Department:  user.Department,
		Description: user.Description,
	})

	if res.Error != nil {
		log.Errorf(ctx, "error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}

	err := r.db.WithContext(ctx).Model(&user).Association("Roles").Replace(user.Roles)
	if err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		return nil, err
	}

	outUser := model.User{}
	res = r.db.WithContext(ctx).Preload("Organization").Preload("Roles").Model(&model.User{}).Where("users.id = ?", user.ID).Find(&outUser)
	if res.Error != nil {
		return nil, res.Error
	}
	return &outUser, nil
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
	var user model.User
	if err := r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Roles").Find(&user, "id = ?", uuid).Error; err != nil {
		log.Errorf(ctx, "error is :%s(%T)", err.Error(), err)
		return err
	}

	res := r.db.WithContext(ctx).Delete(&user)
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

// private members
func (r *UserRepository) getUserByAccountId(ctx context.Context, accountId string, organizationId string) (model.User, error) {
	user := model.User{}
	res := r.db.WithContext(ctx).Model(&model.User{}).Preload("Organization").Preload("Roles").
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
