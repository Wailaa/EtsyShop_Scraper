package repository

import (
	"EtsyScraper/models"
	"EtsyScraper/utils"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DataBase struct {
	DB *gorm.DB
}

type UserRepository interface {
	GetAccountByID(ID uuid.UUID) (account *models.Account, err error)
	GetAccountByEmail(email string) *models.Account
	UpdateLastTimeLoggedIn(Account *models.Account) error
	JoinShopFollowing(Account *models.Account) error
	UpdateLastTimeLoggedOut(UserID uuid.UUID) error
	UpdateAccountAfterVerify(Account *models.Account) error
	UpdateAccountNewPass(Account *models.Account, passwardHashed string) error
	UpdateAccountAfterResetPass(Account *models.Account, newPasswardHashed string) error
	SaveAccount(Account *models.Account) error
	CreateAccount(newAccount *models.Account) (*models.Account, error)
	InsertTokenForAccount(column, token string, VerifyUser *models.Account) (*models.Account, error)
	GetAccountWithShops(account *models.Account) (*models.Account, error)
}

func (d *DataBase) GetAccountByID(ID uuid.UUID) (account *models.Account, err error) {
	if err := d.DB.Where("ID = ?", ID).First(&account).Error; err != nil {
		return nil, utils.HandleError(err, "no account was Found ")
	}
	return
}

func (s *DataBase) GetAccountByEmail(email string) *models.Account {
	account := &models.Account{}

	if err := s.DB.Where("email = ?", email).First(&account).Error; err != nil {
		return account
	}

	return account
}

func (s *DataBase) UpdateLastTimeLoggedIn(Account *models.Account) error {
	now := time.Now()
	if err := s.DB.Model(Account).Where("id = ?", Account.ID).Update("last_time_logged_in", now).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) JoinShopFollowing(Account *models.Account) error {

	if err := s.DB.Preload("ShopsFollowing").First(Account, Account.ID).Error; err != nil {
		return utils.HandleError(err)
	}

	for i := range Account.ShopsFollowing {
		if err := s.DB.Preload("ShopMenu").Preload("Reviews").Preload("Member").First(&Account.ShopsFollowing[i]).Error; err != nil {
			return utils.HandleError(err)
		}
	}

	return nil
}

func (s *DataBase) UpdateLastTimeLoggedOut(UserID uuid.UUID) error {
	now := time.Now()
	if err := s.DB.Model(&models.Account{}).Where("id = ?", UserID).Update("last_time_logged_out", now).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) UpdateAccountAfterVerify(Account *models.Account) error {

	err := s.DB.Model(Account).Updates(map[string]interface{}{"email_verified": true, "email_verification_token": ""}).Error
	if err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) UpdateAccountNewPass(Account *models.Account, passwardHashed string) error {

	err := s.DB.Model(Account).Update("password_hashed", passwardHashed).Error
	if err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) UpdateAccountAfterResetPass(Account *models.Account, newPasswardHashed string) error {

	err := s.DB.Model(Account).Updates(map[string]interface{}{"request_change_pass": false, "account_pass_reset_token": "", "password_hashed": newPasswardHashed}).Error
	if err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) SaveAccount(Account *models.Account) error {
	if err := s.DB.Save(Account).Error; err != nil {
		return utils.HandleError(err)
	}
	return nil
}

func (s *DataBase) CreateAccount(newAccount *models.Account) (*models.Account, error) {
	if err := s.DB.Create(newAccount).Error; err != nil {
		if utils.StringContains(err.Error(), "email") {
			message := errors.New("this email is already in use")
			return newAccount, utils.HandleError(message)
		}
		return newAccount, utils.HandleError(err)
	}
	return newAccount, nil
}

func (s *DataBase) InsertTokenForAccount(column, token string, VerifyUser *models.Account) (*models.Account, error) {
	if err := s.DB.Where(column, token).Find(VerifyUser).Limit(1).Error; err != nil {
		return nil, utils.HandleError(err, "something went wrong while resetting password")
	}
	return VerifyUser, nil
}

func (s *DataBase) GetAccountWithShops(account *models.Account) (*models.Account, error) {

	if err := s.DB.Preload("ShopsFollowing").First(account, account.ID).Error; err != nil {
		return nil, utils.HandleError(err)
	}
	return account, nil
}
