
// 自动生成模板Subjects
package course
import (
	"time"
)

// subjects表 结构体  Subjects
type Subjects struct {
  Id  *int32 `json:"id" form:"id" gorm:"primarykey;comment:学科主键ID;column:id;"`  //学科主键ID
  CreatorId  *int32 `json:"creatorId" form:"creatorId" gorm:"comment:所属作者/所有者ID;column:creator_id;"`  //所属作者/所有者ID
  Slug  *string `json:"slug" form:"slug" gorm:"comment:学科唯一标识，如data_structure;column:slug;size:50;"`  //学科唯一标识，如data_structure
  Name  *string `json:"name" form:"name" gorm:"primarykey;comment:学科显示名称，如数据结构;column:name;size:100;"`  //学科显示名称，如数据结构
  NameDraft  *string `json:"nameDraft" form:"nameDraft" gorm:"primarykey;comment:教材名称草稿;column:name_draft;size:100;"`  //教材名称草稿
  Icon  *string `json:"icon" form:"icon" gorm:"comment:学科图标CSS类名/URL地址;column:icon;size:255;"`  //学科图标CSS类名/URL地址
  IconDraft  *string `json:"iconDraft" form:"iconDraft" gorm:"primarykey;comment:学科图标草稿（CSS类名或URL）;column:icon_draft;size:255;"`  //学科图标草稿（CSS类名或URL）
  Description  *string `json:"description" form:"description" gorm:"comment:学科简介描述;column:description;"`  //学科简介描述
  DescriptionDraft  *string `json:"descriptionDraft" form:"descriptionDraft" gorm:"comment:教材简介草稿;column:description_draft;"`  //教材简介草稿
  CreatedAt  *time.Time `json:"createdAt" form:"createdAt" gorm:"comment:学科创建时间;column:created_at;"`  //学科创建时间
  CoverImageId  *int32 `json:"coverImageId" form:"coverImageId" gorm:"comment:学科封面图片ID，关联images表;column:cover_image_id;"`  //学科封面图片ID，关联images表
  CoverImageIdDraft  *int32 `json:"coverImageIdDraft" form:"coverImageIdDraft" gorm:"primarykey;comment:教材封面ID草稿;column:cover_image_id_draft;"`  //教材封面ID草稿
  Status  *bool `json:"status" form:"status" gorm:"primarykey;comment:教材整体状态;column:status;"`  //教材整体状态
  AuditStatus  *bool `json:"auditStatus" form:"auditStatus" gorm:"comment:审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回;column:audit_status;"`  //审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回
  LastLogId  *int64 `json:"lastLogId" form:"lastLogId" gorm:"comment:关联最新一条审批流水ID;column:last_log_id;"`  //关联最新一条审批流水ID
  HasDraft  *bool `json:"hasDraft" form:"hasDraft" gorm:"comment:是否有未处理的草稿：1=是, 0=否;column:has_draft;"`  //是否有未处理的草稿：1=是, 0=否
}


// TableName subjects表 Subjects自定义表名 subjects
func (Subjects) TableName() string {
    return "subjects"
}





