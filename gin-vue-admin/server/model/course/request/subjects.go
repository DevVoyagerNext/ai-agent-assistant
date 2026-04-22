
package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	
)

type SubjectsSearch struct{
      Id  *int `json:"id" form:"id"` 
      Name  *string `json:"name" form:"name"` 
      NameDraft  *string `json:"nameDraft" form:"nameDraft"` 
      Status  *bool `json:"status" form:"status"` 
    request.PageInfo
}
