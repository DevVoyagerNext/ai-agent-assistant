package dao

import "admin/backend/internal/model"

func (d *AdminDao) FindUserByUsername(username string) (model.User, error) {
	var user model.User
	err := d.db.Where("username = ?", username).First(&user).Error
	return user, err
}

func (d *AdminDao) FindUserByEmail(email string) (model.User, error) {
	var user model.User
	err := d.db.Where("email = ?", email).First(&user).Error
	return user, err
}

func (d *AdminDao) FindAdmin(adminID uint) (model.User, error) {
	var user model.User
	err := d.db.Where("id = ? AND role = ? AND status = ?", adminID, "admin", 1).First(&user).Error
	return user, err
}

func (d *AdminDao) CountUsersByUsernameOrEmail(username string, email string) (int64, error) {
	var count int64
	err := d.db.Model(&model.User{}).Where("username = ? OR email = ?", username, email).Count(&count).Error
	return count, err
}

func (d *AdminDao) CountUsersByEmail(email string) (int64, error) {
	var count int64
	err := d.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count, err
}

func (d *AdminDao) CreateUser(user *model.User) error {
	return d.db.Create(user).Error
}

func (d *AdminDao) UpdateUser(userID uint, updates map[string]any) error {
	return d.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (d *AdminDao) UpdateLastLogin(userID uint, value any) error {
	return d.db.Model(&model.User{}).Where("id = ?", userID).Update("last_login_at", value).Error
}
