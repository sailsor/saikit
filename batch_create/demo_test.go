package main

import (
	"batch_create/dbm"
	"batch_create/utils"
	"context"
	"database/sql"
	"encoding/json"
	"testing"
)

func TestCreate(t *testing.T) {
	logger := utils.GlobalLogger
	c := dbm.GetDBClient()
	order := dbm.NewOrder()
	ctx := context.Background()
	tx := c.GetCtxDb(ctx, "app_db").Table(order.TableName()).Begin(&sql.TxOptions{})
	defer tx.Rollback()

	err := order.Create(tx)
	if err != nil {
		logger.Errorf("Create err:%s", err.Error())
		return
	}

	tx.Commit()
}

func TestCreateDup(t *testing.T) {
	logger := utils.GlobalLogger
	c := dbm.GetDBClient()
	order := dbm.NewOrder()
	ctx := context.Background()
	tx := c.GetCtxDb(ctx, "app_db").Table(order.TableName()).Begin(&sql.TxOptions{})
	defer tx.Rollback()

	order.TradeID = "20250914293139931709047054"
	err := order.Create(tx)
	if err != nil {
		logger.Errorf("Create err:%s", err.Error())
		return
	}

	tx.Commit()
}

func TestFind(t *testing.T) {
	logger := utils.GlobalLogger
	c := dbm.GetDBClient()
	order := new(dbm.TblProdOrderLog)
	ctx := context.Background()
	tx := c.GetCtxDb(ctx, "app_db").Table(order.TableName()).Begin(&sql.TxOptions{})
	defer tx.Rollback()

	order.TradeID = "20250914293139931709047054"
	order.OrderID = "20250914293139931709047054"
	err := order.GetLk(tx)
	if err != nil {
		logger.Errorf("GetLk err:%s", err.Error())
		return
	}

	res, _ := json.MarshalIndent(order, " ", " ")
	logger.Infof("%s", res)

	tx.Commit()
}
