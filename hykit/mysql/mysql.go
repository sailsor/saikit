package mysql

import (
	"context"
	"database/sql"
	"strings"
	"sync"
	"time"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/prometheus/client_golang/prometheus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var clientOnce sync.Once

var onceClient *Client

type Client struct {
	gdbs map[string]*gorm.DB

	proxy []func() interface{}

	conf config.Config

	logger log.Logger

	dbConfigs []DbConfig

	closeChan chan bool

	stateTicker time.Duration

	gormConfig *gorm.Config
}

type Option func(c *Client)

type ClientOptions struct{}

type DbConfig struct {
	Db          string `json:"db" yaml:"db"`
	Dsn         string `json:"dsn" yaml:"dsn"`
	MaxIdle     int    `json:"max_idle" yaml:"maxidle"`
	MaxOpen     int    `json:"max_open" yaml:"maxopen"`
	MaxLifetime int    `json:"max_lifetime" yaml:"maxlifetime"`
}

func NewClient(options ...Option) *Client {
	clientOnce.Do(func() {
		onceClient = &Client{
			gdbs:        make(map[string]*gorm.DB),
			proxy:       make([]func() interface{}, 0),
			stateTicker: 10 * time.Second,
			closeChan:   make(chan bool, 1),
		}

		for _, option := range options {
			option(onceClient)
		}

		if onceClient.conf == nil {
			onceClient.conf = config.NewNullConfig()
		}

		if onceClient.logger == nil {
			onceClient.logger = log.NewLogger()
		}

		onceClient.init()
	})

	return onceClient
}

func (ClientOptions) WithConf(conf config.Config) Option {
	return func(m *Client) {
		m.conf = conf
	}
}

func (ClientOptions) WithLogger(logger log.Logger) Option {
	return func(m *Client) {
		m.logger = logger
	}
}

func (ClientOptions) WithDbConfig(dbConfigs []DbConfig) Option {
	return func(m *Client) {
		m.dbConfigs = dbConfigs
	}
}

func (ClientOptions) WithProxy(proxys ...func() interface{}) Option {
	return func(m *Client) {
		m.proxy = append(m.proxy, proxys...)
	}
}

func (ClientOptions) WithStateTicker(stateTicker time.Duration) Option {
	return func(m *Client) {
		m.stateTicker = stateTicker
	}
}

func (ClientOptions) WithGormConfig(gormConfig *gorm.Config) Option {
	/*增加gorm配置*/
	return func(m *Client) {
		m.gormConfig = gormConfig
		m.gormConfig.PrepareStmt = true
		m.gormConfig.AllowGlobalUpdate = false
	}
}

// initializes Client.
func (c *Client) init() {
	dbConfigs := make([]DbConfig, 0)
	err := c.conf.UnmarshalKey("dbs", &dbConfigs)
	if err != nil {
		c.logger.Panicf("Fatal error config file: %s \n", err.Error())
	}

	if len(c.dbConfigs) > 0 {
		dbConfigs = append(dbConfigs, c.dbConfigs...)
	}

	for _, dbConfig := range dbConfigs {
		var DB *gorm.DB
		var dbc *sql.DB
		DB, err = gorm.Open(mysql.Open(dbConfig.Dsn), c.gormConfig)
		if err != nil {
			c.logger.Panicf("[db] %s open error : %s", dbConfig.Db, err.Error())
			return
		}
		dbc, err = DB.DB()
		if err != nil {
			c.logger.Panicf("[db] %s 获取sql.DB error : %s", dbConfig.Db, err.Error())
			return
		}

		dbc.SetMaxOpenConns(dbConfig.MaxOpen)
		dbc.SetMaxIdleConns(dbConfig.MaxIdle)
		dbc.SetConnMaxLifetime(time.Duration(dbConfig.MaxLifetime) * time.Minute)

		if c.conf.GetBool("debug") {
			DB = DB.Debug()
		}

		/*if len(c.proxy) > 0 {
			firstProxy := proxy.NewProxyFactory().GetFirstInstance("db_"+dbConfig.Db,
				DB.ConnPool, c.proxy...)
			DB.ConnPool = firstProxy.(gorm.ConnPool)
		}*/

		c.setDb(dbConfig.Db, DB)

		go c.Stats()
		c.logger.Infof("[mysql] %s init success", dbConfig.Db)
	}
}

func (c *Client) setDb(dbName string, gdb *gorm.DB) {
	dbName = strings.ToLower(dbName)
	c.gdbs[dbName] = gdb
}

func (c *Client) GetDb(dbName string) *gorm.DB {
	return c.getDb(context.Background(), dbName)
}

func (c *Client) getDb(ctx context.Context, dbName string) *gorm.DB {
	dbName = strings.ToLower(dbName)
	if db, ok := c.gdbs[dbName]; ok {
		return db.WithContext(ctx)
	}

	c.logger.Errorf("[db] %s not found", dbName)

	return nil
}

func (c *Client) GetCtxDb(ctx context.Context, dbName string) *gorm.DB {
	if db, ok := ctx.Value(prefix + dbName).(*gorm.DB); ok {
		return db
	}
	return c.getDb(ctx, dbName)
}

const prefix = "CTX_DB_"

// SetCtxSession 初始化DB 会话
func (c *Client) SetCtxSession(ctx context.Context) context.Context {
	for dbName, db := range c.gdbs {
		ctx = context.WithValue(ctx, prefix+dbName, db.WithContext(ctx))
	}
	return ctx
}

// CloseCtxSession 关闭缓存stmt
func (c *Client) CloseCtxSession(ctx context.Context) {
	for dbName, _ := range c.gdbs {
		if db, ok := ctx.Value(prefix + dbName).(*gorm.DB); ok {
			if stmtManger, ok := db.ConnPool.(*gorm.PreparedStmtDB); ok {
				stmtManger.Close()
			}
		}
	}
	return
}

func (c *Client) Ping() []error {
	var errs []error
	for name, db := range c.gdbs {
		dbc, err := db.DB()
		if err != nil {
			c.logger.Panicf("获取dbc [%s] err:[%s]", name, err)
		}
		err = dbc.Ping()
		if err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func (c *Client) Close() {
	for name, db := range c.gdbs {
		dbc, err := db.DB()
		if err != nil {
			c.logger.Panicf("获取dbc [%s] err:[%s]", name, err)
		}
		err = dbc.Close()
		if err != nil {
			c.logger.Errorf(err.Error())
		}
	}
}

func (c *Client) Stats() {
	defer func() {
		if err := recover(); err != nil {
			c.logger.Infof(err.(error).Error())
		}
	}()

	ticker := time.NewTicker(c.stateTicker)
	var stats sql.DBStats

	for {
		select {
		case <-ticker.C:
			for dbName, db := range c.gdbs {
				sqldb, _ := db.DB()
				stats = sqldb.Stats()

				maxOpenConnLab := prometheus.Labels{"db": dbName, "stats": "max_open_conn"}
				mysqlStats.With(maxOpenConnLab).Set(float64(stats.MaxOpenConnections))

				openConnLab := prometheus.Labels{"db": dbName, "stats": "open_conn"}
				mysqlStats.With(openConnLab).Set(float64(stats.OpenConnections))

				inUseLab := prometheus.Labels{"db": dbName, "stats": "in_use"}
				mysqlStats.With(inUseLab).Set(float64(stats.InUse))

				idleLab := prometheus.Labels{"db": dbName, "stats": "idle"}
				mysqlStats.With(idleLab).Set(float64(stats.Idle))

				waitCountLab := prometheus.Labels{"db": dbName, "stats": "wait_count"}
				mysqlStats.With(waitCountLab).Set(float64(stats.WaitCount))
			}
		case <-c.closeChan:
			c.logger.Infof("stop stats")
			goto Stop
		}
	}

Stop:
	ticker.Stop()
}
