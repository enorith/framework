package database_test

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/enorith/framework/database"
	"github.com/enorith/gormdb"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var testDSN = "root:632258@tcp(192.168.8.21:3306)/cdut_role"

type User struct {
	ID        int    `gorm:"column:id"`
	Nickname  string `gorm:"column:nickname"`
	CreatedAt string `gorm:"column:created_at"`
}

func (tu User) TableName() string {
	return "users"
}

func Test_Builder(t *testing.T) {
	tx, e := gormdb.DefaultManager.GetConnection()

	if e != nil {
		t.Fatal(e)
	}

	b := database.NewBuilder(tx, database.NewPaginator[User](TestPageParams{
		P:  1,
		Ps: 10,
	}))

	u, _ := b.First(12)
	fmt.Println(u)

	us, e := b.Query(func(d *gorm.DB) *gorm.DB {
		return d.Where("id > ?", 1)
	}).Get()

	if e != nil {
		t.Fatal(e)
	}

	for _, u := range us {
		fmt.Println(u)
	}
}

func init() {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer（日志输出的目标，前缀和日志包含的内容——译者注）
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略ErrRecordNotFound（记录未找到）错误
		},
	)

	gormdb.DefaultManager.RegisterDefault(func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(testDSN), &gorm.Config{
			Logger: newLogger,
		})
	})

}

type TestPageParams struct {
	P  int
	Ps int
}

func (t TestPageParams) Page() int {
	return t.P
}

func (t TestPageParams) PerPage() int {
	return t.Ps
}
