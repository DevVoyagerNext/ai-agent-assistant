
<template>
  <div>
    <div class="gva-search-box">
      <el-form ref="elSearchFormRef" :inline="true" :model="searchInfo" class="demo-form-inline" @keyup.enter="onSubmit">
            <el-form-item label="学科ID" prop="id">
  <el-input v-model.number="searchInfo.id" placeholder="搜索条件" />
</el-form-item>
            
            <el-form-item label="作者ID" prop="creatorId">
  <el-input v-model.number="searchInfo.creatorId" placeholder="搜索条件" />
</el-form-item>
            
            <el-form-item label="学科名称" prop="name">
  <el-input v-model="searchInfo.name" placeholder="搜索条件" />
</el-form-item>
            
            <el-form-item label="名称草稿" prop="nameDraft">
  <el-input v-model="searchInfo.nameDraft" placeholder="搜索条件" />
</el-form-item>
            
            <el-form-item label="教材状态" prop="status">
  <el-input v-model="searchInfo.status" placeholder="搜索条件" />
</el-form-item>
            

        <template v-if="showAllQuery">
          <!-- 将需要控制显示状态的查询条件添加到此范围内 -->
        </template>

        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
          <el-button link type="primary" icon="arrow-down" @click="showAllQuery=true" v-if="!showAllQuery">展开</el-button>
          <el-button link type="primary" icon="arrow-up" @click="showAllQuery=false" v-else>收起</el-button>
        </el-form-item>
      </el-form>
    </div>
    <div class="gva-table-box">
        <div class="gva-btn-list">
            <el-button  type="primary" icon="plus" @click="openDialog()">新增</el-button>
            <el-button  icon="delete" style="margin-left: 10px;" :disabled="!multipleSelection.length" @click="onDelete">删除</el-button>
            
        </div>
        <el-table
        ref="multipleTable"
        style="width: 100%"
        tooltip-effect="dark"
        :data="tableData"
        row-key="id"
        @selection-change="handleSelectionChange"
        >
        <el-table-column type="selection" width="55" />
        
            <el-table-column align="left" label="学科ID" prop="id" width="120" />

            <el-table-column align="left" label="作者ID" prop="creatorId" width="120" />

            <el-table-column align="left" label="学科标识" prop="slug" width="120" />

            <el-table-column align="left" label="学科名称" prop="name" width="120" />

            <el-table-column align="left" label="名称草稿" prop="nameDraft" width="120" />

            <el-table-column align="left" label="学科图标" prop="icon" width="120" />

            <el-table-column align="left" label="图标草稿" prop="iconDraft" width="120" />

            <el-table-column align="left" label="学科简介" prop="description" width="120" />

            <el-table-column align="left" label="简介草稿" prop="descriptionDraft" width="120" />

            <el-table-column align="left" label="创建时间" prop="createdAt" width="180">
   <template #default="scope">{{ formatDate(scope.row.createdAt) }}</template>
</el-table-column>
            <el-table-column align="left" label="学科封面ID" prop="coverImageId" width="120" />

            <el-table-column align="left" label="封面ID草稿" prop="coverImageIdDraft" width="120" />

            <el-table-column align="left" label="教材状态" prop="status" width="120" />

            <el-table-column align="left" label="审核状态" prop="auditStatus" width="120" />

            <el-table-column align="left" label="审批流水" prop="lastLogId" width="120" />

            <el-table-column align="left" label="未处理草稿" prop="hasDraft" width="120">
    <template #default="scope">{{ formatBoolean(scope.row.hasDraft) }}</template>
</el-table-column>
        <el-table-column align="left" label="操作" fixed="right" :min-width="appStore.operateMinWith">
            <template #default="scope">
            <el-button  type="primary" link class="table-button" @click="getDetails(scope.row)"><el-icon style="margin-right: 5px"><InfoFilled /></el-icon>查看</el-button>
            <el-button  type="primary" link icon="edit" class="table-button" @click="updateSubjectsFunc(scope.row)">编辑</el-button>
            <el-button   type="primary" link icon="delete" @click="deleteRow(scope.row)">删除</el-button>
            </template>
        </el-table-column>
        </el-table>
        <div class="gva-pagination">
            <el-pagination
            layout="total, sizes, prev, pager, next, jumper"
            :current-page="page"
            :page-size="pageSize"
            :page-sizes="[10, 30, 50, 100]"
            :total="total"
            @current-change="handleCurrentChange"
            @size-change="handleSizeChange"
            />
        </div>
    </div>
    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="dialogFormVisible" :show-close="false" :before-close="closeDialog">
       <template #header>
              <div class="flex justify-between items-center">
                <span class="text-lg">{{type==='create'?'新增':'编辑'}}</span>
                <div>
                  <el-button :loading="btnLoading" type="primary" @click="enterDialog">确 定</el-button>
                  <el-button @click="closeDialog">取 消</el-button>
                </div>
              </div>
            </template>

          <el-form :model="formData" label-position="top" ref="elFormRef" :rules="rule" label-width="80px">
            <el-form-item label="教材状态:" prop="status">
    <el-input v-model="formData.status" :clearable="true" placeholder="请输入教材状态" />
</el-form-item>
          </el-form>
    </el-drawer>

    <el-drawer destroy-on-close :size="appStore.drawerSize" v-model="detailShow" :show-close="true" :before-close="closeDetailShow" title="查看">
            <el-descriptions :column="1" border>
                    <el-descriptions-item label="学科ID">
    {{ detailForm.id }}
</el-descriptions-item>
                    <el-descriptions-item label="作者ID">
    {{ detailForm.creatorId }}
</el-descriptions-item>
                    <el-descriptions-item label="学科标识">
    {{ detailForm.slug }}
</el-descriptions-item>
                    <el-descriptions-item label="学科名称">
    {{ detailForm.name }}
</el-descriptions-item>
                    <el-descriptions-item label="名称草稿">
    {{ detailForm.nameDraft }}
</el-descriptions-item>
                    <el-descriptions-item label="学科图标">
    {{ detailForm.icon }}
</el-descriptions-item>
                    <el-descriptions-item label="图标草稿">
    {{ detailForm.iconDraft }}
</el-descriptions-item>
                    <el-descriptions-item label="学科简介">
    {{ detailForm.description }}
</el-descriptions-item>
                    <el-descriptions-item label="简介草稿">
    {{ detailForm.descriptionDraft }}
</el-descriptions-item>
                    <el-descriptions-item label="创建时间">
    {{ detailForm.createdAt }}
</el-descriptions-item>
                    <el-descriptions-item label="学科封面ID">
    {{ detailForm.coverImageId }}
</el-descriptions-item>
                    <el-descriptions-item label="封面ID草稿">
    {{ detailForm.coverImageIdDraft }}
</el-descriptions-item>
                    <el-descriptions-item label="教材状态">
    {{ detailForm.status }}
</el-descriptions-item>
                    <el-descriptions-item label="审核状态">
    {{ detailForm.auditStatus }}
</el-descriptions-item>
                    <el-descriptions-item label="审批流水">
    {{ detailForm.lastLogId }}
</el-descriptions-item>
                    <el-descriptions-item label="未处理草稿">
    {{ detailForm.hasDraft }}
</el-descriptions-item>
            </el-descriptions>
        </el-drawer>

  </div>
</template>

<script setup>
import {
  createSubjects,
  deleteSubjects,
  deleteSubjectsByIds,
  updateSubjects,
  findSubjects,
  getSubjectsList
} from '@/api/course/subjects'

// 全量引入格式化工具 请按需保留
import { getDictFunc, formatDate, formatBoolean, filterDict ,filterDataSource, returnArrImg, onDownloadFile } from '@/utils/format'
import { ElMessage, ElMessageBox } from 'element-plus'
import { ref, reactive } from 'vue'
import { useAppStore } from "@/pinia"




defineOptions({
    name: 'Subjects'
})

// 提交按钮loading
const btnLoading = ref(false)
const appStore = useAppStore()

// 控制更多查询条件显示/隐藏状态
const showAllQuery = ref(false)

// 自动化生成的字典（可能为空）以及字段
const formData = ref({
            status: '',
        })



// 验证规则
const rule = reactive({
})

const elFormRef = ref()
const elSearchFormRef = ref()

// =========== 表格控制部分 ===========
const page = ref(1)
const total = ref(0)
const pageSize = ref(10)
const tableData = ref([])
const searchInfo = ref({})
// 重置
const onReset = () => {
  searchInfo.value = {}
  getTableData()
}

// 搜索
const onSubmit = () => {
  elSearchFormRef.value?.validate(async(valid) => {
    if (!valid) return
    page.value = 1
    if (searchInfo.value.hasDraft === ""){
        searchInfo.value.hasDraft=null
    }
    getTableData()
  })
}

// 分页
const handleSizeChange = (val) => {
  pageSize.value = val
  getTableData()
}

// 修改页面容量
const handleCurrentChange = (val) => {
  page.value = val
  getTableData()
}

// 查询
const getTableData = async() => {
  const table = await getSubjectsList({ page: page.value, pageSize: pageSize.value, ...searchInfo.value })
  if (table.code === 0) {
    tableData.value = table.data.list
    total.value = table.data.total
    page.value = table.data.page
    pageSize.value = table.data.pageSize
  }
}

getTableData()

// ============== 表格控制部分结束 ===============

// 获取需要的字典 可能为空 按需保留
const setOptions = async () =>{
}

// 获取需要的字典 可能为空 按需保留
setOptions()


// 多选数据
const multipleSelection = ref([])
// 多选
const handleSelectionChange = (val) => {
    multipleSelection.value = val
}

// 删除行
const deleteRow = (row) => {
    ElMessageBox.confirm('确定要删除吗?', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
    }).then(() => {
            deleteSubjectsFunc(row)
        })
    }

// 多选删除
const onDelete = async() => {
  ElMessageBox.confirm('确定要删除吗?', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async() => {
      const ids = []
      if (multipleSelection.value.length === 0) {
        ElMessage({
          type: 'warning',
          message: '请选择要删除的数据'
        })
        return
      }
      multipleSelection.value &&
        multipleSelection.value.map(item => {
          ids.push(item.id)
        })
      const res = await deleteSubjectsByIds({ ids })
      if (res.code === 0) {
        ElMessage({
          type: 'success',
          message: '删除成功'
        })
        if (tableData.value.length === ids.length && page.value > 1) {
          page.value--
        }
        getTableData()
      }
      })
    }

// 行为控制标记（弹窗内部需要增还是改）
const type = ref('')

// 更新行
const updateSubjectsFunc = async(row) => {
    const res = await findSubjects({ id: row.id })
    type.value = 'update'
    if (res.code === 0) {
        formData.value = res.data
        dialogFormVisible.value = true
    }
}


// 删除行
const deleteSubjectsFunc = async (row) => {
    const res = await deleteSubjects({ id: row.id })
    if (res.code === 0) {
        ElMessage({
                type: 'success',
                message: '删除成功'
            })
            if (tableData.value.length === 1 && page.value > 1) {
            page.value--
        }
        getTableData()
    }
}

// 弹窗控制标记
const dialogFormVisible = ref(false)

// 打开弹窗
const openDialog = () => {
    type.value = 'create'
    dialogFormVisible.value = true
}

// 关闭弹窗
const closeDialog = () => {
    dialogFormVisible.value = false
    formData.value = {
        status: '',
        }
}
// 弹窗确定
const enterDialog = async () => {
     btnLoading.value = true
     elFormRef.value?.validate( async (valid) => {
             if (!valid) return btnLoading.value = false
              let res
              switch (type.value) {
                case 'create':
                  res = await createSubjects(formData.value)
                  break
                case 'update':
                  res = await updateSubjects(formData.value)
                  break
                default:
                  res = await createSubjects(formData.value)
                  break
              }
              btnLoading.value = false
              if (res.code === 0) {
                ElMessage({
                  type: 'success',
                  message: '创建/更改成功'
                })
                closeDialog()
                getTableData()
              }
      })
}

const detailForm = ref({})

// 查看详情控制标记
const detailShow = ref(false)


// 打开详情弹窗
const openDetailShow = () => {
  detailShow.value = true
}


// 打开详情
const getDetails = async (row) => {
  // 打开弹窗
  const res = await findSubjects({ id: row.id })
  if (res.code === 0) {
    detailForm.value = res.data
    openDetailShow()
  }
}


// 关闭详情弹窗
const closeDetailShow = () => {
  detailShow.value = false
  detailForm.value = {}
}


</script>

<style>

</style>
