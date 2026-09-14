package ds

type Engineer struct {
	ID       uint   `gorm:"primaryKey"`
	FullName string `gorm:"column:full_name;type:varchar(150);not null"`
}

func (Engineer) TableName() string {
	return "engineers"
}
