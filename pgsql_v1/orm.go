package pgsql_v1

import (
	"database/sql"
	"errors"
	"github.com/larspensjo/config"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

// 全局变量 -------------------------------------------------------------------------
var (
	handles    = make(map[string]*ormPgsql) // 实例句柄
	handleLock = sync.RWMutex{}             // 实例句柄操作锁
	pathConfig = ""                         // 配置文件地址
	dbSections = map[string]string{}        // 名称=> 数据库
)

// @Title 初始化配置文件路径
func Init(cfgPath string) {

	// 配置文件路径赋值
	pathConfig = cfgPath

	// 读取配置文件
	cfg, err := config.ReadDefault(pathConfig)
	if err != nil {
		log.Panic("读取配置文件出错" + err.Error())
		return
	}

	// 初始化所有数据库配置项的连接
	for _, v := range cfg.Sections() {
		if v[:3] == "pg_" {
			dbSections[v[3:]], _ = cfg.String(v, "pg")
			_, err := getConnectedHandle(v[3:])
			if err != nil {
				log.Panic("数据库连接初始化失败", v)
			}
		}
	}
}

// 初始化
func getConnectedHandle(dbCfgName string, varDbName ...string) (*ormPgsql, error) {
	// 判断配置文件是否已赋值
	if pathConfig == "" {
		log.Panic("请先初始化设置数据库配置文件")
		return nil, errors.New("请先初始化设置数据库配置文件")
	}

	// 没有数据库配置项
	if _, ok := dbSections[dbCfgName]; !ok {
		log.Error(pathConfig + " 没有数据库 " + dbCfgName + " 这个配置项 ")
		return nil, errors.New(pathConfig + " 没有数据库 " + dbCfgName + " 这个配置项 ")
	}

	// 数据库名称
	dbName := dbSections[dbCfgName]
	if len(varDbName) > 0 {
		dbName = varDbName[0]
	}

	// 定义实例map键值名称
	var dbInstance = "dbInstance|" + dbCfgName + "|" + dbName

	// 句柄已存在，直接返回
	if true {
		handleLock.Lock()
		_, ok := handles[dbInstance]
		handleLock.Unlock()

		if ok {
			return handles[dbInstance], nil
		}
	}

	// 读取配置文件
	host, port, db, username, password, _, maxIdle, maxConn, interpolateParams, maxLifetime, err := getDbConfig("pg_" + dbCfgName)
	if err != nil {
		log.Error("读取配置文件出错:" + err.Error())
		return nil, err
	}

	// 参数传了数据库名称，则使用传入的数据库名称
	if dbName != "" {
		db = dbName
	}

	// 连接数据库
	dataSourceName := "host=" + host + " port=" + port + " user=" + username + " password=" + password + " dbname=" + db + " sslmode=disable"
	if interpolateParams {
		dataSourceName += "&interpolateParams=true"
	}
	dbHandle, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Error(err)
		log.Error("数据库:"+dbCfgName+" . "+dbName, " 连接失败")
		return nil, err
	}
	dbHandle.SetMaxOpenConns(maxConn)
	dbHandle.SetMaxIdleConns(maxIdle)
	dbHandle.SetConnMaxLifetime(time.Second * time.Duration(maxLifetime))

	oneHandle := &ormPgsql{
		o:          dbHandle,
		dbInstance: dbInstance,
		dbCfgName:  dbCfgName,
		dbName:     dbName,
		initErr:    false,
	}

	handleLock.Lock()
	handles[dbInstance] = oneHandle
	handleLock.Unlock()

	return oneHandle, nil
}

// @Title 获取数据库句柄，所有配置信息从配置文件读取
func Handle(Name ...string) *ormPgsql {
	// 有参数使用传入的参数，否则使用default
	if len(Name) == 0 {
		Name = append(Name, "default")
	}

	// 获取指定配置和库名的句柄
	handle, err := getConnectedHandle(Name[0], Name[1:]...)
	if err != nil {
		return &ormPgsql{initErr: true}
	}

	// 返回连接句柄
	return handle
}

// @Title 获取配置文件
func getDbConfig(name string) (host string, port string, db string, username string, password string, charset string, maxIdle int, maxConn int, interpolateParams bool, maxLifetime int, err error) {

	// 读取配置文件
	cfg, err := config.ReadDefault(pathConfig)
	if err != nil {
		log.Error("读取配置文件出错" + err.Error())
		return "", "", "", "", "", "", 0, 0, false, 0, err
	}

	// 取出配置项
	host, hostErr := cfg.String(name, "host")
	port, _ = cfg.String(name, "port")
	db, _ = cfg.String(name, "db")
	username, usernameErr := cfg.String(name, "username")
	password, passwordErr := cfg.String(name, "password")
	charset, charsetErr := cfg.String(name, "charset")
	maxIdle, maxIdleErr := cfg.Int(name, "maxIdle")
	maxConn, maxConnErr := cfg.Int(name, "maxConn")
	interpolateParams, _ = cfg.Bool(name, "interpolateParams")
	maxLifetime, _ = cfg.Int(name, "maxLifetime")

	// 主配置项出错
	if hostErr != nil || usernameErr != nil || passwordErr != nil {
		log.Error("出错")
		return "", "", "", "", "", "", 0, 0, false, 0, errors.New(name + "数据库主配置项为空")
	}

	// 可设置默认值配置项
	if charsetErr != nil {
		charset = "utf8mb4"
	}
	if maxIdleErr != nil {
		maxIdle = 8
	}
	if maxConnErr != nil {
		maxConn = 20
	}
	if maxLifetime == 0 { // 默认4个小时过期
		maxLifetime = 4 * 60 * 60
	}

	// 返回
	return host, port, db, username, password, charset, maxIdle, maxConn, interpolateParams, maxLifetime, nil
}
