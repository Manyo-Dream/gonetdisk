package repository

import (
	"errors"
	"fmt"

	"github.com/manyodream/gonetdisk/internal/model"
	"gorm.io/gorm"
)

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

// PhysicalFile
func (r *FileRepo) CreatePhyFile(phyfile *model.PhysicalFile) error {
	return r.db.Create(phyfile).Error
}

func (r *FileRepo) GetPhyFileByFileName(filename string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := r.db.Model(&model.PhysicalFile{}).Where("file_name = ?", filename).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, nil
}

func (r *FileRepo) GetPhyFileByID(id uint64) (*model.PhysicalFile, error) {
	var phyFile model.PhysicalFile
	err := r.db.Where("id = ?", id).First(&phyFile).Error
	return &phyFile, err
}

func (r *FileRepo) HashDeduplication(fileHash string) (*model.PhysicalFile, error) {
	var phyfile model.PhysicalFile

	err := r.db.Model(&model.PhysicalFile{}).Where("file_hash = ?", fileHash).First(&phyfile).Error
	if err != nil {
		return nil, err
	}

	return &phyfile, err
}

// UserFile
func (r *FileRepo) CreateUserFile(userfile *model.UserFile) error {
	return r.db.Create(userfile).Error
}

func (r *FileRepo) UpdateUserFile(userfile *model.UserFile) error {
	return r.db.Model(&model.UserFile{}).Save(userfile).Error
}

func (r *FileRepo) GetUserFileByFolderName(userID, parentID uint64, folderName string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.db.Where("user_id = ? AND parent_id = ? AND file_name = ? AND is_dir = ?", userID, parentID, folderName, true).
		First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetParentFolderByParentID(userID, parentID uint64) (*model.UserFile, error) {
	var userFolder model.UserFile

	err := r.db.Where("user_id = ? AND id = ? AND is_dir = ?", userID, parentID, true).First(&userFolder).Error
	if err != nil {
		return nil, err
	}

	return &userFolder, nil
}

func (r *FileRepo) GetUserFolderByID(userID, parentID uint64) (*model.UserFile, error) {
	var userFolder model.UserFile

	err := r.db.Where("user_id = ? AND id = ? AND is_dir = ?", userID, parentID, true).First(&userFolder).Error
	if err != nil {
		return nil, err
	}
	return &userFolder, nil
}

func (r *FileRepo) GetUserFileByID(userID, userFileID uint64) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.db.Model(&model.UserFile{}).Where("user_id = ? AND id = ? AND is_dir = ?", userID, userFileID, false).
		First(&userfile).Error
	if err != nil {
		return nil, err
	}
	return &userfile, nil
}

func (r *FileRepo) GetParentIDyFolderName(userID uint64, folderName string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.db.Where("user_id = ? AND file_name = ? AND is_dir = ?", userID, folderName, true).First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetUserFileByFileName(userID, parentID uint64, filename string) (*model.UserFile, error) {
	var userfile model.UserFile

	err := r.db.Model(&model.UserFile{}).
		Where("user_id = ? AND parent_id = ? AND file_name = ? AND is_dir = ?", userID, parentID, filename, false).
		First(&userfile).Error
	if err != nil {
		return nil, err
	}

	return &userfile, nil
}

func (r *FileRepo) GetUserFileList(userID, parentID uint64, page, pageSize int, sortBy, orderBy string) ([]model.UserFile, int64, error) {
	// 构建查询
	query := r.db.Model(&model.UserFile{}).
		Where("user_id = ? AND parent_id = ?", userID, parentID)

	// 统计总数
	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("获取文件总数失败: %w", err)
	}

	if total == 0 {
		return []model.UserFile{}, 0, nil
	}

	// 分页查询
	var items []model.UserFile

	offset := (page - 1) * pageSize
	orderClause := fmt.Sprintf("is_dir DESC, %s %s", sortBy, orderBy)

	err := query.Order(orderClause).
		Offset(offset).
		Limit(pageSize).
		Find(&items).Error
	if err != nil {
		return []model.UserFile{}, 0, fmt.Errorf("分页查询用户文件失败: %w", err)
	}

	return items, total, nil
}

func (r *FileRepo) GetChildrenFiles(userID, parentID uint64) ([]model.UserFile, error) {
	var childrenFiles []model.UserFile

	err := r.db.Model(&model.UserFile{}).Where("user_id = ? AND parent_id = ? ", userID, parentID).Find(&childrenFiles).Error
	if err != nil {
		return nil, fmt.Errorf("查询子文件失败: %w", err)
	}

	return childrenFiles, nil
}

func (r *FileRepo) GetTrashChildrenFiles(userID, parentID uint64) ([]model.UserFile, error) {
	var childrenFiles []model.UserFile

	err := r.db.Unscoped().Model(&model.UserFile{}).Where("user_id = ? AND parent_id = ?", userID, parentID).
		Find(&childrenFiles).Error
	if err != nil {
		return nil, fmt.Errorf("查询子文件失败: %w", err)
	}

	return childrenFiles, nil
}

func (r *FileRepo) GetTrashFileList(userID uint64, page, pageSize int) ([]model.UserFile, int64, error) {
	var total int64
	var userFiles []model.UserFile

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := r.db.Unscoped().Model(&model.UserFile{}).Where("user_id = ? AND deleted_at IS NOT NULL", userID)

	if err := query.Count(&total).Error; err != nil {
		return []model.UserFile{}, 0, fmt.Errorf("获取回收站文件总数失败: %w", err)
	}
	if total == 0 {
		return []model.UserFile{}, 0, nil
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("deleted_at DESC").Find(&userFiles).Error
	if err != nil {
		return []model.UserFile{}, 0, fmt.Errorf("分页查询回收站文件失败: %w", err)
	}

	return userFiles, total, nil
}

func (r *FileRepo) GetUserFileByIDAny(userID, userFileID uint64) (*model.UserFile, error) {
	var userFile model.UserFile
	err := r.db.Unscoped().Model(&model.UserFile{}).
		Where("user_id = ? AND id = ?", userID, userFileID).First(&userFile).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("文件不存在")
		}
		return nil, fmt.Errorf("查询文件失败: %w", err)
	}

	return &userFile, nil
}

func (r *FileRepo) SoftDeleteUserItem(userID, userFileID uint64) error {
	result := r.db.Model(&model.UserFile{}).
		Where("user_id = ? AND id = ?", userID, userFileID).
		Delete(&model.UserFile{})

	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到符合条件的文件")
	}

	return nil
}

func (r *FileRepo) RestoreUserFile(userID, userFileID uint64) error {
	result := r.db.Unscoped().Model(&model.UserFile{}).
		Where("user_id = ? AND id = ? AND deleted_at IS NOT NULL", userID, userFileID).
		Update("deleted_at", nil)
	if result.Error != nil {
		return fmt.Errorf("恢复文件失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("文件不存在或不在回收站中")
	}

	return nil
}

// Others
func (r *FileRepo) GetSpace(userid string) (uint64, error) {
	var user model.User
	err := r.db.Where("user_id = ?", userid).First(&user).Error
	return user.Total_Space, err
}

func (r *FileRepo) IncrPhyFileRefCount(id uint64, delta int) error {
	return r.db.Model(&model.PhysicalFile{}).
		Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("COALESCE(ref_count, 0) + ?", delta)).
		Error
}

func (r *FileRepo) DecrPhyFileRefCount(id uint64, delta int) error {
	return r.db.Model(&model.PhysicalFile{}).
		Where("id = ?", id).
		UpdateColumn("ref_count", gorm.Expr("COALESCE(ref_count, 0) - ?", delta)).
		Error
}

func (r *FileRepo) GetPhyFileByUserFile(userID, userFileID uint64) (*uint64, error) {
	var result struct {
		PhysicalID *uint64 `gorm:"column:physical_id"`
	}

	err := r.db.Model(&model.UserFile{}).
		Select("physical_id").
		Where("id = ? AND user_id = ? AND is_dir = ?", userFileID, userID, false).
		Take(&result).Error
	if err != nil {
		return nil, err
	}

	if result.PhysicalID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	return result.PhysicalID, nil
}

func (r *FileRepo) GetFileByDownloadReq(userID, userFileID uint64) (*model.UserFile, *model.PhysicalFile, error) {
	var userFile model.UserFile

	err := r.db.
		Preload("PhysicalFile"). //预载PhysicalFile表，通过相连外键进行查询
		Where("id = ? AND user_id = ? AND is_dir = ?", userFileID, userID, false).
		Take(&userFile).Error
	if err != nil {
		return nil, nil, err
	}

	if userFile.PhysicalID == 0 || userFile.PhysicalFile == nil {
		return nil, nil, gorm.ErrRecordNotFound
	}

	return &userFile, userFile.PhysicalFile, nil
}

func (r *FileRepo) UpdateUserFilePath(id uint64, pathStack string) error {
	return r.db.Model(&model.UserFile{}).Where("id = ?", id).Update("path_stack", pathStack).Error
}
