
package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	
)

type SubjectsSearch struct{
      Id  *int `json:"id" form:"id"` 
      CreatorId  *int `json:"creatorId" form:"creatorId"` 
      Name  *string `json:"name" form:"name"` 
      NameDraft  *string `json:"nameDraft" form:"nameDraft"` 
      Status  *string `json:"status" form:"status"` 
    request.PageInfo
}
