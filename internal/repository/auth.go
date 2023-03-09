package repository

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/openinfradev/tks-api/pkg/domain"
	"github.com/openinfradev/tks-api/pkg/log"
)

// Interface
type IAuthRepository interface {
	GetUser(userId uuid.UUID) (domain.User, error)
	GetUserByAccountId(accountId string) (domain.User, error)
	Create(accountId string, password string, name string) (domain.User, error)
	AssignRole(accountId string, roleName string) error
}

type AuthRepository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) IAuthRepository {
	return &AuthRepository{
		db: db,
	}
}

// Models
type User struct {
	gorm.Model

	ID            uuid.UUID `gorm:"primarykey;type:uuid;"`
	AccountId     string    `gorm:"uniqueIndex"`
	Name          string
	Password      string
	AuthType      string         `gorm:"authType"`
	Roles         []Role         `gorm:"many2many:user_roles"`
	Organizations []Organization `gorm:"many2many:organization_users"`
	Creator       uuid.UUID
}

func (g *User) BeforeCreate(tx *gorm.DB) (err error) {
	g.ID = uuid.New()
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

// Public members
func (r *AuthRepository) GetUser(userId uuid.UUID) (user domain.User, err error) {
	return domain.User{}, nil
}

func (r *AuthRepository) GetUserByAccountId(accountId string) (domain.User, error) {
	user, err := r.getUserByAccountId(accountId)
	if err != nil {
		return domain.User{}, err
	}

	return r.reflect(user), nil
}

func (r *AuthRepository) Create(accountId string, password string, name string) (domain.User, error) {
	_, err := r.GetUserByAccountId(accountId)
	if err == nil {
		return domain.User{}, fmt.Errorf("Already existed user %s", accountId)
	}

	newUser := User{
		AccountId: accountId,
		Password:  password,
		Name:      name,
	}
	res := r.db.Create(&newUser)
	if res.Error != nil {
		return domain.User{}, res.Error
	}

	return r.reflect(newUser), nil
}

func (r *AuthRepository) AssignRole(accountId string, roleName string) error {
	user, err := r.getUserByAccountId(accountId)
	if err != nil {
		return fmt.Errorf("Failed to get user %s", accountId)
	}

	role, err := r.getRoleByName(roleName)
	if err != nil {
		return fmt.Errorf("Failed to get role %s", roleName)
	}

	newRole := UserRole{
		UserId: user.ID,
		RoleId: role.ID,
	}
	res := r.db.Create(&newRole)
	if res.Error != nil {
		return res.Error
	}

	return nil
}

func (r *AuthRepository) GetRoleByName(roleName string) (out domain.Role, err error) {
	role, err := r.getRoleByName(roleName)
	if err != nil {
		return domain.Role{}, err
	}
	return r.reflectRole(role), nil
}

// private members
func (r *AuthRepository) getUserByAccountId(accountId string) (User, error) {
	user := User{}
	res := r.db.Model(&User{}).Preload("Organizations").Preload("Roles").Find(&user, "account_id = ?", accountId)
	if res.RowsAffected == 0 || res.Error != nil {
		return User{}, fmt.Errorf("Not found user. %s", res.Error)
	}

	return user, nil
}

func (r *AuthRepository) getRoleByName(roleName string) (out Role, err error) {
	role := Role{}
	res := r.db.First(&role, "name = ?", roleName)
	if res.RowsAffected == 0 || res.Error != nil {
		log.Info(res.Error)
		return Role{}, fmt.Errorf("Not found role for %s", roleName)
	}

	return role, nil
}

func (r *AuthRepository) reflect(user User) domain.User {
	resRoles := []domain.Role{}
	for _, role := range user.Roles {
		outRole := domain.Role{
			ID:          role.ID.String(),
			Name:        role.Name,
			Description: role.Description,
			Creator:     role.Creator.String(),
			CreatedAt:   role.CreatedAt,
			UpdatedAt:   role.UpdatedAt,
		}
		resRoles = append(resRoles, outRole)
	}

	resOrganizations := []domain.Organization{}
	for _, organization := range user.Organizations {
		outOrganization := domain.Organization{
			ID:          organization.ID,
			Name:        organization.Name,
			Description: organization.Description,
			Creator:     organization.Creator.String(),
			CreatedAt:   organization.CreatedAt,
			UpdatedAt:   organization.UpdatedAt,
		}
		resOrganizations = append(resOrganizations, outOrganization)
	}

	return domain.User{
		ID:            user.ID.String(),
		AccountId:     user.AccountId,
		Password:      user.Password,
		Name:          user.Name,
		Roles:         resRoles,
		Organizations: resOrganizations,
		Creator:       user.Creator.String(),
		CreatedAt:     user.CreatedAt,
		UpdatedAt:     user.UpdatedAt,
	}
}

func (r *AuthRepository) reflectRole(role Role) domain.Role {
	return domain.Role{
		ID:          role.ID.String(),
		Name:        role.Name,
		Description: role.Description,
		Creator:     role.Creator.String(),
		CreatedAt:   role.CreatedAt,
		UpdatedAt:   role.UpdatedAt,
	}
}
