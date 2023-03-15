package gorm_tool

import (
	"context"
	"gorm_tool/dbm"
	"gorm_tool/utils"
	"io/ioutil"
	"path/filepath"
	"testing"
)

/*
读取文件
*/
func TestReadIpFile(t *testing.T) {
	utils.InitGlobalLogger()
	logger := utils.GlobalLogger
	ctx := context.Background()
	filePath := "C:\\Users\\sai\\Documents\\银联\\公网安全组"
	fileName := "高危攻击IP0309.xlsx"

	list, err := utils.ReadUnionXlsxFile(ctx, filePath, fileName)
	if err != nil {
		logger.Errorc(ctx, "ReadUnionXlsxFile err : %s", err)
		return
	}
	logger.Debugf("list[0] %v", list[0])
}

/*
入库银联公网安全组
*/
func TestInsetIp(t *testing.T) {
	utils.InitGlobalLogger()
	dbm.InitDB()
	defer dbm.CloseDB()

	logger := utils.GlobalLogger
	ctx := context.Background()
	pathname := "C:\\Users\\sai\\Documents\\银联\\公网安全组"

	logger.Infoc(ctx, "开始测试....")

	fileList := make([]string, 0)
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		logger.Errorf("read dir fail:", err)
		return
	}
	for _, fi := range rd {
		if fi.IsDir() {
			logger.Infof("子文件[%s]是目录跳过", fi.Name())
		} else {
			logger.Infof("子文件[%s]", fi.Name())
			fileList = append(fileList, fi.Name())
		}
	}

	err = utils.CreateIpInfo(ctx, pathname, fileList)
	if err != nil {
		logger.Errorc(ctx, "CreateIpInfo err: %s", err)
		return
	}
}

/*
查询是否存在高危IP
*/
func TestQueryHighRiskIp(t *testing.T) {
	utils.InitGlobalLogger()
	dbm.InitDB()
	defer dbm.CloseDB()

	logger := utils.GlobalLogger
	ctx := context.Background()
	pathname := "C:\\Users\\sai\\Documents\\银联\\高危ip\\now"

	logger.Infoc(ctx, "开始测试....")

	fileList := make([]string, 0)
	rd, err := ioutil.ReadDir(pathname)
	if err != nil {
		logger.Errorf("read dir fail:", err)
		return
	}
	for _, fi := range rd {
		if fi.IsDir() {
			logger.Infof("子文件[%s]是目录跳过", fi.Name())
		} else {
			logger.Infof("子文件[%s]", fi.Name())
			fileName := filepath.Join(pathname, fi.Name())
			fileList = append(fileList, fileName)
		}
	}

	err = utils.QueryHighRiskIp(ctx, fileList)
	if err != nil {
		logger.Errorf("QueryHighRiskIp err: %s", err)
		return
	}
}
