package helper

import (
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	"github.com/jinzhu/configor"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

type Config struct {
	Exportfile string `mapstructure:"exportfile" json:"exportfile" yaml:"exportfile" default:"./upload/"`
	Ormcrm     struct {
		Driver       string `mapstructure:"driver" json:"driver" yaml:"driver" default:"mysql"`
		Path         string `mapstructure:"path" json:"path" yaml:"path"`                             // 服务器地址
		Port         string `mapstructure:"port" json:"port" yaml:"port"`                             // 端口
		Config       string `mapstructure:"config" json:"config" yaml:"config"`                       // 高级配置
		Dbname       string `mapstructure:"db-name" json:"dbname" yaml:"db-name"`                     // 数据库名
		Username     string `mapstructure:"username" json:"username" yaml:"username"`                 // 数据库用户名
		Password     string `mapstructure:"password" json:"password" yaml:"password"`                 // 数据库密码
		MaxIdleConns int    `mapstructure:"max-idle-conns" json:"maxIdleConns" yaml:"max-idle-conns"` // 空闲中的最大连接数
		MaxOpenConns int    `mapstructure:"max-open-conns" json:"maxOpenConns" yaml:"max-open-conns"` // 打开到数据库的最大连接数
		LogMode      string `mapstructure:"log-mode" json:"logMode" yaml:"log-mode"`                  // 是否开启Gorm全局日志
		LogZap       bool   `mapstructure:"log-zap" json:"logZap" yaml:"log-zap"`                     // 是否通过zap写入日志文件
	}
}

//dsn := "sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
//db, err := gorm.Open(sqlserver.Open(dsn), &gorm.Config{})

//orm连接crm库
func GormCrm() *gorm.DB {
	var config = Config{}
	err := configor.Load(&config, "config.yml")
	if err != nil {
		panic(err)
	}
	uname := config.Ormcrm.Username
	upass := config.Ormcrm.Password
	upath := config.Ormcrm.Path
	uport := config.Ormcrm.Port
	udbname := config.Ormcrm.Dbname
	uconfig := config.Ormcrm.Config
	umic := config.Ormcrm.MaxIdleConns
	umoc := config.Ormcrm.MaxOpenConns

	driver := config.Ormcrm.Driver

	if driver == "mysql" {
		dsnStr := uname + ":" + upass + "@tcp(" + upath + ":" + uport + ")/" + udbname + "?" + uconfig
		mysqlConfig := mysql.Config{
			DSN:                       dsnStr, // DSN data source name
			DefaultStringSize:         191,    // string 类型字段的默认长度
			SkipInitializeWithVersion: false,  // 根据版本自动配置
		}
		if db, err := gorm.Open(mysql.New(mysqlConfig), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, //使用单数表明，就是表明后面不加s
			},
		}); err != nil {
			return nil
		} else {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxIdleConns(umic)
			sqlDB.SetMaxOpenConns(umoc)
			sqlDB.SetConnMaxLifetime(time.Hour)
			return db
		}
	} else if driver == "mssql" {
		//  DSN: "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local",
		//"sqlserver://gorm:LoremIpsum86@localhost:9930?database=gorm"
		dsnStr := fmt.Sprintf("server=%s;port=%s;database=%s;user id=%s;password=%s;encrypt=disable", upath, uport, udbname, uname, upass)
		fmt.Println("连接串为>>>", dsnStr)
		sqlserverConfig := sqlserver.Config{
			DSN:               dsnStr, // DSN data source name
			DefaultStringSize: 191,    // string 类型字段的默认长度
		}
		if db, err := gorm.Open(sqlserver.New(sqlserverConfig), &gorm.Config{
			NamingStrategy: schema.NamingStrategy{
				SingularTable: true, //使用单数表明，就是表明后面不加s
			},
		}); err != nil {
			return nil
		} else {
			sqlDB, _ := db.DB()
			sqlDB.SetMaxIdleConns(umic)
			sqlDB.SetMaxOpenConns(umoc)
			sqlDB.SetConnMaxLifetime(time.Hour)
			return db
		}
	} else {
		return nil
	}
}
