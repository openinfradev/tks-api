package repository

import (
	"github.com/google/uuid"
	"github.com/openinfradev/tks-api/pkg/httpErrors"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interface
type IUserRepository interface {
	Create(accountId string, organizationId string, paasword string, name string) (domain.User, error)
	CreateWithUuid(uuid uuid.UUID, accountId string, name string, password string, email string,
		department string, description string, orgainzationId string, roleId uuid.UUID) (domain.User, error)
	List(...FilterFunc) (out *[]domain.User, err error)
	Get(accountId string, organizationId string) (domain.User, error)
	GetByUuid(userId uuid.UUID) (domain.User, error)
	UpdateWithUuid(uuid uuid.UUID, accountId string, name string, password string, roleId uuid.UUID, email string,
		department string, description string) (domain.User, error)
	DeleteWithUuid(uuid uuid.UUID) error
	Flush(organizationId string) error

	FetchRoles() (out *[]domain.Role, err error)
	AssignRole(accountId string, organizationId string, roleName string) error
	AccountIdFilter(accountId string) FilterFunc
	OrganizationFilter(organization string) FilterFunc
	AssignRoleWithUuid(uuid uuid.UUID, roleName string) error
}

type UserRepository struct {
	db *gorm.DB
}

func (r *UserRepository) Flush(organizationId string) error {
	res := r.db.Where("organization_id = ?", organizationId).Delete(&User{})
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}

// Models
type User struct {
	gorm.Model

	ID             uuid.UUID `gorm:"primarykey;type:uuid"`
	AccountId      string
	Name           string
	Password       string
	AuthType       string `gorm:"authType"`
	RoleId         uuid.UUID
	Role           Role `gorm:"foreignKey:RoleId;references:ID"`
	OrganizationId string
	Organization   Organization `gorm:"foreignKey:OrganizationId;references:ID"`
	Creator        uuid.UUID
	Email          string
	Department     string
	Description    string
}

//func (g *User) BeforeCreate(tx *gorm.DB) (err error) {
//	g.ID = uuid.New()
//	return nil
//}

func (r *UserRepository) Create(accountId string, organizationId string, password string, name string) (domain.User, error) {
	newUser := User{
		AccountId:      accountId,
		Password:       password,
		OrganizationId: organizationId,
		Name:           name,
	}
	res := r.db.Create(&newUser)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return domain.User{}, res.Error
	}

	return r.reflect(newUser), nil
}
func (r *UserRepository) CreateWithUuid(uuid uuid.UUID, accountId string, name string, password string, email string,
	department string, description string, organizationId string, roleId uuid.UUID) (domain.User, error) {

	newUser := User{
		ID:             uuid,
		AccountId:      accountId,
		Password:       password,
		Name:           name,
		Email:          email,
		Department:     department,
		Description:    description,
		OrganizationId: organizationId,
		RoleId:         roleId,
	}
	res := r.db.Create(&newUser)
	if res.Error != nil {
		log.Error(res.Error.Error())
		return domain.User{}, res.Error
	}
	user, err := r.getUserByAccountId(accountId, organizationId)
	if err != nil {
		return domain.User{}, err
	}

	return r.reflect(user), nil
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
func (r *UserRepository) List(filters ...FilterFunc) (*[]domain.User, error) {
	var users []User
	var res *gorm.DB
	if filters == nil {
		res = r.db.Model(&User{}).Preload("Organization").Preload("Role").Find(&users)
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
		res = cFunc(r.db.Model(&User{}).Preload("Organization").Preload("Role")).Find(&users)
	}
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}

	var out []domain.User
	for _, user := range users {
		out = append(out, r.reflect(user))
	}

	return &out, nil
}
func (r *UserRepository) Get(accountId string, organizationId string) (domain.User, error) {
	user, err := r.getUserByAccountId(accountId, organizationId)
	if err != nil {
		return domain.User{}, err
	}

	return r.reflect(user), nil
}
func (r *UserRepository) GetByUuid(userId uuid.UUID) (respUser domain.User, err error) {
	user := User{}
	res := r.db.Model(&User{}).Preload("Organization").Preload("Role").Find(&user, "id = ?", userId)

	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return domain.User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}

	return r.reflect(user), nil
}
func (r *UserRepository) UpdateWithUuid(uuid uuid.UUID, accountId string, name string, password string, roleId uuid.UUID,
	email string, department string, description string) (domain.User, error) {
	var user User
	res := r.db.Model(&User{}).Where("id = ?", uuid).Updates(User{
		AccountId:   accountId,
		Name:        name,
		Password:    password,
		Email:       email,
		Department:  department,
		Description: description,
		RoleId:      roleId,
	})
	if res.RowsAffected == 0 || res.Error != nil {
		return domain.User{}, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return domain.User{}, res.Error
	}
	res = r.db.Model(&User{}).Preload("Organization").Preload("Role").Where("id = ?", uuid).Find(&user)
	if res.Error != nil {
		return domain.User{}, res.Error
	}
	return r.reflect(user), nil
}
func (r *UserRepository) DeleteWithUuid(uuid uuid.UUID) error {
	res := r.db.Unscoped().Delete(&User{}, "id = ?", uuid)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}
	return nil
}

type Role struct {
	gorm.Model

	ID          uuid.UUID `gorm:"primarykey;type:uuid;"`
	Name        string
	Description string
	Creator     uuid.UUID
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	r.ID = uuid.New()
	return nil
}

type Policy struct {
	gorm.Model

	//ID               uuid.UUID `gorm:"primarykey;type:uuid;"`
	RoleId           uuid.UUID
	Role             Role `gorm:"references:ID"`
	Name             string
	Description      string
	Create           bool `gorm:"column:c"`
	CreatePriviledge string
	Update           bool `gorm:"column:u"`
	UpdatePriviledge string
	Read             bool `gorm:"column:r"`
	ReadPriviledge   string
	Delete           bool `gorm:"column:d"`
	DeletePriviledge string
	Creator          uuid.UUID
}

type UserRole struct {
	UserId uuid.UUID
	User   User
	RoleId uuid.UUID
	Role   Role
}

func (r *UserRepository) AssignRoleWithUuid(uuid uuid.UUID, roleName string) error {
	_, err := r.GetByUuid(uuid)
	if err != nil {
		return err
	}

	role, err := r.getRoleByName(roleName)
	if err != nil {
		return err
	}

	newRole := UserRole{
		UserId: uuid,
		RoleId: role.ID,
	}
	res := r.db.Create(&newRole)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}

	return nil
}

func (r *UserRepository) AssignRole(accountId string, organizationId string, roleName string) error {
	user, err := r.getUserByAccountId(accountId, organizationId)
	if err != nil {
		return err
	}

	role, err := r.getRoleByName(roleName)
	if err != nil {
		return err
	}

	newRole := UserRole{
		UserId: user.ID,
		RoleId: role.ID,
	}
	res := r.db.Create(&newRole)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return res.Error
	}

	return nil
}

func (r *UserRepository) GetRoleByName(roleName string) (domain.Role, error) {
	role, err := r.getRoleByName(roleName)
	if err != nil {
		return domain.Role{}, err
	}

	return r.reflectRole(role), nil
}

func (r *UserRepository) FetchRoles() (*[]domain.Role, error) {
	var roles []Role
	res := r.db.Find(&roles)

	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return nil, res.Error
	}

	if res.RowsAffected == 0 {
		return nil, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}

	var out []domain.Role
	for _, role := range roles {
		outRole := r.reflectRole(role)
		out = append(out, outRole)
	}

	return &out, nil
}

// private members
func (r *UserRepository) getUserByAccountId(accountId string, organizationId string) (User, error) {
	user := User{}
	res := r.db.Model(&User{}).Preload("Organization").Preload("Role").
		Find(&user, "account_id = ? AND organization_id = ?", accountId, organizationId)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return User{}, res.Error
	}
	if res.RowsAffected == 0 {
		return User{}, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}

	return user, nil
}

func (r *UserRepository) getRoleByName(roleName string) (Role, error) {
	role := Role{}
	res := r.db.First(&role, "name = ?", roleName)
	if res.Error != nil {
		log.Errorf("error is :%s(%T)", res.Error.Error(), res.Error)
		return Role{}, res.Error
	}
	if res.RowsAffected == 0 {
		return Role{}, httpErrors.NewNotFoundError(httpErrors.NotFound)
	}

	//if res.RowsAffected == 0 {
	//	return Role{}, nil
	//}

	return role, nil
}

func (r *UserRepository) reflect(user User) domain.User {
	role := domain.Role{
		ID:          user.Role.ID.String(),
		Name:        user.Role.Name,
		Description: user.Role.Description,
		Creator:     user.Role.Creator.String(),
		CreatedAt:   user.Role.CreatedAt,
		UpdatedAt:   user.Role.UpdatedAt,
	}
	//for _, role := range user.Roles {
	//	outRole := domain.Role{
	//		ID:          role.ID.String(),
	//		Name:        role.Name,
	//		Description: role.Description,
	//		Creator:     role.Creator.String(),
	//		CreatedAt:   role.CreatedAt,
	//		UpdatedAt:   role.UpdatedAt,
	//	}
	//	resRoles = append(resRoles, outRole)
	//}

	organization := domain.Organization{
		ID:                user.Organization.ID,
		Name:              user.Organization.Name,
		Description:       user.Organization.Description,
		Phone:             user.Organization.Phone,
		Status:            user.Organization.Status,
		StatusDescription: user.Organization.StatusDesc,
		Creator:           user.Organization.Creator.String(),
		CreatedAt:         user.Organization.CreatedAt,
		UpdatedAt:         user.Organization.UpdatedAt,
	}
	//for _, organization := range user.Organizations {
	//	outOrganization := domain.Organization{
	//		ID:          organization.ID,
	//		Name:        organization.Name,
	//		Description: organization.Description,
	//		Creator:     organization.Creator.String(),
	//		CreatedAt:   organization.CreatedAt,
	//		UpdatedAt:   organization.UpdatedAt,
	//	}
	//	resOrganizations = append(resOrganizations, outOrganization)
	//}

	return domain.User{
		ID:           user.ID.String(),
		AccountId:    user.AccountId,
		Password:     user.Password,
		Name:         user.Name,
		Role:         role,
		Organization: organization,
		Creator:      user.Creator.String(),
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Department:   user.Department,
		Description:  user.Description,
	}
}

func (r *UserRepository) reflectRole(role Role) domain.Role {
	return domain.Role{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		Creator:     role.Creator.String(),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
