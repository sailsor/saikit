package mysql

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"code.jshyjdtech.com/godev/hykit/config"
	"code.jshyjdtech.com/godev/hykit/log"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	test1Config = DbConfig{
		Db:      "bat_test_db",
		Dsn:     "root:root@tcp(localhost:3306)/bat_test_db?charset=utf8&parseTime=True&loc=Local",
		MaxIdle: 10,
		MaxOpen: 100}

	test2Config = DbConfig{
		Db:      "onl_test_db",
		Dsn:     "root:root@tcp(localhost:3306)/onl_test_db?charset=utf8&parseTime=True&loc=Local",
		MaxIdle: 10,
		MaxOpen: 100}
)

type TestStruct struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
}

type UserStruct struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

var db *sql.DB
var logger log.Logger

/*func TestMain(m *testing.M) {
	logger = log.NewLogger()

	pool, err := dockertest.NewPool("")
	if err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	opt := &dockertest.RunOptions{
		Repository: "mysql/mysql-server:8.0",
		Tag:        "latest",
		Env:        []string{"MYSQL_ROOT_PASSWORD=root"},
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opt, func(hostConfig *dc.HostConfig) {
		hostConfig.PortBindings = map[dc.Port][]dc.PortBinding{
			"3306/tcp": {{HostIP: "", HostPort: "3306"}},
		}
	})
	if err != nil {
		logger.Fatalf("Could not start resource: %s", err.Error())
	}

	err = resource.Expire(50)
	if err != nil {
		logger.Fatalf(err.Error())
	}

	if err := pool.Retry(func() error {
		var err error
		db, err = sql.Open("mysql",
			"root:root@tcp(localhost:3306)/mysql?charset=utf8&parseTime=True&loc=Local")
		if err != nil {
			return err
		}
		db.SetMaxOpenConns(100)

		return db.Ping()
	}); err != nil {
		logger.Fatalf("Could not connect to docker: %s", err)
	}

	sqls := []string{
		`create database bat_test_db;`,
		`CREATE TABLE IF NOT EXISTS bat_test_db.test(
		  id int not NULL auto_increment,
		  title VARCHAR(10) not NULL DEFAULT '',
		  PRIMARY KEY (id)
		)engine=innodb;`,
		`create database onl_test_db;`,
		`CREATE TABLE IF NOT EXISTS onl_test_db.user(
		  id int not NULL auto_increment,
		  username VARCHAR(10) not NULL DEFAULT '',
			PRIMARY KEY (id)
		)engine=innodb;`}

	for _, execSQL := range sqls {
		res, err := db.Exec(execSQL)
		if err != nil {
			logger.Errorf(err.Error())
		}
		_, err = res.RowsAffected()
		if err != nil {
			logger.Errorf(err.Error())
		}
	}
	code := m.Run()

	db.Close()
	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logger.Fatalf("Could not purge resource: %s", err)
	}
	os.Exit(code)
}*/

func TestInitAndSingleInstance(t *testing.T) {
	clientOptions := ClientOptions{}

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	_, ok := client.gdbs["bat_test_db"]
	assert.True(t, ok)

	assert.Equal(t, client, NewClient())

	client.Close()
}

func TestProxyPatternWithTwoInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()
	memConfig.Set("debug", false)

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithConf(memConfig),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(false),
					log.WithLogLevel(zapcore.FatalLevel),
				)),
			),
		}),
	)

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)
	assert.Nil(t, db1.Error)

	t.Logf("test1:%+v", ts)

	db2 := client.GetCtxDb(ctx, "onl_test_db")
	db2.Exec("use onl_test_db;")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)
	assert.Nil(t, db2.Error)

	t.Logf("test2:%+v,%v", us, db2.Error)

	client.Close()
}

func TestMulProxyPatternWithOneInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
		clientOptions.WithConf(memConfig),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")

	t.Logf("db1.ConnPool %p", client.gdbs)

	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)
	assert.Nil(t, db1.Error)

	client.Close()
}

func TestMulProxyPatternWithTwoInstance(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithConf(memConfig),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)

	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	ts := &TestStruct{}
	db1.Table("test").First(ts)

	assert.Nil(t, db1.Error)

	db2 := client.GetCtxDb(ctx, "onl_test_db")
	db2.Exec("use onl_test_db;")
	assert.NotNil(t, db2)

	us := &UserStruct{}
	db2.Table("user").First(us)

	assert.Nil(t, db1.Error)

	client.Close()
}

func BenchmarkParallelGetDB(b *testing.B) {
	clientOnce = sync.Once{}

	b.ReportAllocs()
	b.ResetTimer()

	clientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithConf(memConfig),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ctx := context.Background()
			client.GetCtxDb(ctx, "bat_test_db")

			db2 := client.GetCtxDb(ctx, "onl_test_db")
			db2.Exec("use onl_test_db;")
			assert.NotNil(b, db2)
		}
	})

	//	//client.Close()

	b.StopTimer()
}

func TestDummyProxy_Exec(t *testing.T) {
	clientOnce = sync.Once{}

	clientOptions := ClientOptions{}
	memConfig := config.NewMemConfig()

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config}),
		clientOptions.WithConf(memConfig),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	db1.Table("test").Create(&TestStruct{})

	assert.Equal(t, db1.RowsAffected, int64(0))

	client.Close()
}

func TestClient_GetStats(t *testing.T) {
	clientOnce = sync.Once{}
	clientOptions := ClientOptions{}

	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithStateTicker(10*time.Millisecond),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	time.Sleep(100 * time.Millisecond)

	lab := prometheus.Labels{"db": "bat_test_db", "stats": "max_open_conn"}
	c, _ := mysqlStats.GetMetricWith(lab)
	metric := &io_prometheus_client.Metric{}
	err := c.Write(metric)
	assert.Nil(t, err)

	assert.Equal(t, float64(100), metric.Gauge.GetValue())

	labIdle := prometheus.Labels{"db": "bat_test_db", "stats": "idle"}
	c, _ = mysqlStats.GetMetricWith(labIdle)
	metric = &io_prometheus_client.Metric{}
	err = c.Write(metric)
	assert.Nil(t, err)

	assert.Equal(t, float64(1), metric.Gauge.GetValue())

	client.Close()
}

//nolint:dupl
func TestClient_TxCommit(t *testing.T) {
	clientOnce = sync.Once{}
	clientOptions := ClientOptions{}
	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)
	ctx := context.Background()
	ctx = client.SetCtxSession(ctx)

	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	db2 := client.GetCtxDb(ctx, "bat_test_db")
	assert.NotNil(t, db2)
	db2.Exec("use bat_test_db;")

	tx := db1.Begin()
	assert.Nil(t, tx.Error)
	db1.Exec("insert into test values (?, ?)", 10, "1000")
	db1.Exec("insert into test values (?, ?)", 11, "1010")
	db1.Exec("insert into test values (?, ?)", 12, "1010")
	db1.Exec("insert into test values (?, ?)", 13, "1010")
	assert.Nil(t, tx.Error)

	test := &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	test = &TestStruct{}

	tx.Table("test").First(test)
	assert.Equal(t, 10, test.ID)
	tx.Commit()

	client.Close()
}

//nolint:dupl
func TestClient_TxRollBack(t *testing.T) {
	clientOnce = sync.Once{}
	clientOptions := ClientOptions{}
	client := NewClient(
		clientOptions.WithDbConfig([]DbConfig{test1Config, test2Config}),
		clientOptions.WithGormConfig(&gorm.Config{
			Logger: log.NewGormLogger(
				log.WithGLogEsimZap(log.NewEsimZap(
					log.WithEsimZapDebug(true),
				)),
			),
		}),
	)
	ctx := context.Background()
	db1 := client.GetCtxDb(ctx, "bat_test_db")
	db1.Exec("use bat_test_db;")
	assert.NotNil(t, db1)

	tx := db1.Begin()
	assert.Nil(t, tx.Error)
	tx.Exec("insert into test values (100, 'test')")
	tx.Rollback()

	assert.Nil(t, tx.Error)

	ts := TestStruct{}
	db1.Table("test").Where("id = 100").First(&ts)
	assert.Equal(t, 0, ts.ID)

	//client.Close()
}
