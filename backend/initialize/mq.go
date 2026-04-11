package initialize

import (
	"context"

	"backend/global"
	"backend/pkg/mq"
	"backend/pkg/mq/tasks"
)

func MQ() {
	// 1. 初始化全局 MQ 实例
	global.GVA_MQ = mq.NewRedisMQ(global.GVA_REDIS, global.GVA_LOG)

	// 2. 注册任务处理器 (此处根据业务按需注册)
	global.GVA_MQ.Register("subject_progress", tasks.HandleSubjectProgress)

	// 3. 启动消费者协程
	ctx := context.Background()
	global.GVA_MQ.Start(ctx)
}
