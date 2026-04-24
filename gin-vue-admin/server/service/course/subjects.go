
package course

import (
	"context"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/course"
    courseReq "github.com/flipped-aurora/gin-vue-admin/server/model/course/request"
)

type SubjectsService struct {}
// CreateSubjects 创建教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService) CreateSubjects(ctx context.Context, subjects *course.Subjects) (err error) {
	err = global.GVA_DB.Create(subjects).Error
	return err
}

// DeleteSubjects 删除教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService)DeleteSubjects(ctx context.Context, id string) (err error) {
	err = global.GVA_DB.Delete(&course.Subjects{},"id = ?",id).Error
	return err
}

// DeleteSubjectsByIds 批量删除教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService)DeleteSubjectsByIds(ctx context.Context, ids []string) (err error) {
	err = global.GVA_DB.Delete(&[]course.Subjects{},"id in ?",ids).Error
	return err
}

// UpdateSubjects 更新教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService)UpdateSubjects(ctx context.Context, subjects course.Subjects) (err error) {
	err = global.GVA_DB.Model(&course.Subjects{}).Where("id = ?",subjects.Id).Updates(&subjects).Error
	return err
}

// GetSubjects 根据id获取教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService)GetSubjects(ctx context.Context, id string) (subjects course.Subjects, err error) {
	err = global.GVA_DB.Where("id = ?", id).First(&subjects).Error
	return
}
// GetSubjectsInfoList 分页获取教材审批记录
// Author [yourname](https://github.com/yourname)
func (subjectsService *SubjectsService)GetSubjectsInfoList(ctx context.Context, info courseReq.SubjectsSearch) (list []course.Subjects, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)
    // 创建db
	db := global.GVA_DB.Model(&course.Subjects{})
    var subjectss []course.Subjects
    // 如果有条件搜索 下方会自动创建搜索语句
    
    if info.Id != nil {
        db = db.Where("id = ?", *info.Id)
    }
    if info.CreatorId != nil {
        db = db.Where("creator_id = ?", *info.CreatorId)
    }
    if info.Name != nil && *info.Name != "" {
        db = db.Where("name LIKE ?", "%"+ *info.Name+"%")
    }
    if info.NameDraft != nil && *info.NameDraft != "" {
        db = db.Where("name_draft LIKE ?", "%"+ *info.NameDraft+"%")
    }
    if info.Status != nil && *info.Status != "" {
        db = db.Where("status = ?", *info.Status)
    }
	err = db.Count(&total).Error
	if err!=nil {
    	return
    }

	if limit != 0 {
       db = db.Limit(limit).Offset(offset)
    }

	err = db.Find(&subjectss).Error
	return  subjectss, total, err
}
func (subjectsService *SubjectsService)GetSubjectsPublic(ctx context.Context) {
    // 此方法为获取数据源定义的数据
    // 请自行实现
}
