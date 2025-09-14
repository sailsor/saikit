package main

import (
	"batch_create/dbm"
	"batch_create/utils"
	"code.jshyjdtech.com/godev/hykit/pkg/taskpool"
	"context"
	"database/sql"
	"time"
)

func main() {
	BatchOrder()
	time.Sleep(10 * time.Second)
}

func BatchOrder() {
	tp := taskpool.NewTaskPool()
	tp.Run()

	for i := 0; i < 10; i++ {
		tp.AddFunc(func() {
			order, err := OrderCreate()
			if err == nil && order != nil {
				OrderUpdate(order)
			}
		})
	}
}

func OrderCreate() (*dbm.TblProdOrderLog, error) {
	logger := utils.GlobalLogger
	c := dbm.GetDBClient()
	order := dbm.NewOrder()
	ctx := context.Background()
	tx := c.GetCtxDb(ctx, "app_db").Table(order.TableName()).Begin(&sql.TxOptions{})
	defer tx.Rollback()

	order.TradeID = "20250914289380037314540789"
	err := order.Create(tx)
	if err != nil {
		logger.Errorf("Create [%s] err", order.OrderID)
		return nil, err
	}

	tx.Commit()
	return order, nil
}

func OrderUpdate(order *dbm.TblProdOrderLog) {
	logger := utils.GlobalLogger
	c := dbm.GetDBClient()
	ctx := context.Background()
	tx := c.GetCtxDb(ctx, "app_db").Table(order.TableName()).Begin(&sql.TxOptions{})
	defer tx.Rollback()

	order.TradeStatus = "TRADE_SUCCESS"
	err := order.Update(tx)
	if err != nil {
		logger.Errorf("Update [%s] err", order.OrderID)
		return
	}

	tx.Commit()
	return
}
