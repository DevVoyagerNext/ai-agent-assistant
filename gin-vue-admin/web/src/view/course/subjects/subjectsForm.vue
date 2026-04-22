
<template>
  <div>
    <div class="gva-form-box">
      <el-form :model="formData" ref="elFormRef" label-position="right" :rules="rule" label-width="80px">
        <el-form-item label="所属作者/所有者ID:" prop="creatorId">
    <el-input v-model.number="formData.creatorId" :clearable="true" placeholder="请输入所属作者/所有者ID" />
</el-form-item>
        <el-form-item label="学科唯一标识，如data_structure:" prop="slug">
    <el-input v-model="formData.slug" :clearable="true" placeholder="请输入学科唯一标识，如data_structure" />
</el-form-item>
        <el-form-item label="学科图标CSS类名/URL地址:" prop="icon">
    <el-input v-model="formData.icon" :clearable="true" placeholder="请输入学科图标CSS类名/URL地址" />
</el-form-item>
        <el-form-item label="学科简介描述:" prop="description">
    <el-input v-model="formData.description" :clearable="true" placeholder="请输入学科简介描述" />
</el-form-item>
        <el-form-item label="教材简介草稿:" prop="descriptionDraft">
    <el-input v-model="formData.descriptionDraft" :clearable="true" placeholder="请输入教材简介草稿" />
</el-form-item>
        <el-form-item label="学科创建时间:" prop="createdAt">
    <el-date-picker v-model="formData.createdAt" type="date" style="width:100%" placeholder="选择日期" :clearable="true" />
</el-form-item>
        <el-form-item label="学科封面图片ID，关联images表:" prop="coverImageId">
    <el-input v-model.number="formData.coverImageId" :clearable="true" placeholder="请输入学科封面图片ID，关联images表" />
</el-form-item>
        <el-form-item label="教材整体状态:" prop="status">
    <el-switch v-model="formData.status" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="审核状态：0=编辑中, 1=待审核, 2=已通过, 3=被驳回:" prop="auditStatus">
    <el-switch v-model="formData.auditStatus" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item label="关联最新一条审批流水ID:" prop="lastLogId">
    <el-input v-model.number="formData.lastLogId" :clearable="true" placeholder="请输入关联最新一条审批流水ID" />
</el-form-item>
        <el-form-item label="是否有未处理的草稿：1=是, 0=否:" prop="hasDraft">
    <el-switch v-model="formData.hasDraft" active-color="#13ce66" inactive-color="#ff4949" active-text="是" inactive-text="否" clearable ></el-switch>
</el-form-item>
        <el-form-item>
          <el-button :loading="btnLoading" type="primary" @click="save">保存</el-button>
          <el-button type="primary" @click="back">返回</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script setup>
import {
  createSubjects,
  updateSubjects,
  findSubjects
} from '@/api/course/subjects'

defineOptions({
    name: 'SubjectsForm'
})

// 自动获取字典
import { getDictFunc } from '@/utils/format'
import { useRoute, useRouter } from "vue-router"
import { ElMessage } from 'element-plus'
import { ref, reactive } from 'vue'


const route = useRoute()
const router = useRouter()

// 提交按钮loading
const btnLoading = ref(false)

const type = ref('')
const formData = ref({
            creatorId: undefined,
            slug: '',
            icon: '',
            description: '',
            descriptionDraft: '',
            createdAt: new Date(),
            coverImageId: undefined,
            status: false,
            auditStatus: false,
            lastLogId: undefined,
            hasDraft: false,
        })
// 验证规则
const rule = reactive({
})

const elFormRef = ref()

// 初始化方法
const init = async () => {
 // 建议通过url传参获取目标数据ID 调用 find方法进行查询数据操作 从而决定本页面是create还是update 以下为id作为url参数示例
    if (route.query.id) {
      const res = await findSubjects({ ID: route.query.id })
      if (res.code === 0) {
        formData.value = res.data
        type.value = 'update'
      }
    } else {
      type.value = 'create'
    }
}

init()
// 保存按钮
const save = async() => {
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
           }
       })
}

// 返回按钮
const back = () => {
    router.go(-1)
}

</script>

<style>
</style>
