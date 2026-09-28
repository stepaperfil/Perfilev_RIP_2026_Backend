package ds

type Engineer struct {
	ID           uint   `gorm:"primaryKey"`
	FullName     string `gorm:"column:full_name;type:varchar(150);not null"`
	Login        string `gorm:"column:login;type:varchar(100)"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255)"`
}

func (Engineer) TableName() string {
	return "engineers"
}
