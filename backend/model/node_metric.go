package model

import "time"

type NodeMetric struct {
	ID          uint64    `gorm:"primaryKey;autoIncrement;comment:统计记录自增ID" json:"id"`
	NodeID      int       `gorm:"not null;uniqueIndex:uk_node_metric,priority:1;comment:关联知识节点ID" json:"nodeId"`
	MetricType  string    `gorm:"not null;type:varchar(50);uniqueIndex:uk_node_metric,priority:2;index:idx_type_value,priority:1;comment:指标类型" json:"metricType"`
	MetricValue int       `gorm:"default:0;index:idx_type_value,priority:2;comment:指标数值" json:"metricValue"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP;comment:更新时间" json:"updatedAt"`
}

func (NodeMetric) TableName() string {
	return "node_metrics"
}
