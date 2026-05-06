package model

import "time"

// AIAgentConfig AI Agent 配置表
type AIAgentConfig struct {
	ID           int64     `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	AgentKey     string    `gorm:"type:varchar(64);not null;uniqueIndex:uk_agent_key;comment:唯一键，如 problem_solve_agent" json:"agentKey"`
	SystemPrompt string    `gorm:"type:text;not null;comment:系统提示词" json:"systemPrompt"`
	ModelName    string    `gorm:"type:varchar(64);not null;default:'doubao-pro-32k';comment:选用的模型" json:"modelName"`
	Temperature  float64   `gorm:"type:decimal(3,2);not null;default:0.70;comment:温度参数" json:"temperature"`
	IsActive     int       `gorm:"type:tinyint(1);not null;default:1;comment:是否启用" json:"isActive"`
	Description  string    `gorm:"type:varchar(255);comment:备注" json:"description"`
	CreatedAt    time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"type:datetime;not null;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

// TableName 表名
func (AIAgentConfig) TableName() string {
	return "ai_agent_config"
}
