package utils

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/tealeg/xlsx"
	"gorm.io/gorm"
	"gorm_tool/dbm"
	"path/filepath"
	"strings"
	"time"
)

func CreateIpInfo(ctx context.Context, dir string, ipFiles []string) error {
	logger := GlobalLogger
	dbClient := dbm.GetDBClient()
	tx := dbClient.GetCtxDb(ctx, "test_admin_mgm_db").Begin(&sql.TxOptions{})
	defer tx.Rollback()

	for _, ipFile := range ipFiles {
		list, err := ReadUnionXlsxFile(ctx, dir, ipFile)
		if err != nil {
			logger.Errorc(ctx, "ReadUnionXlsxFile err: %s", err)
			return err
		}
		for _, unionGroupIp := range list {
			err = tx.Create(&unionGroupIp).Error
			if err != nil {
				logger.Errorc(ctx, "安全组插入失败[%s][%s]", unionGroupIp.GroupName, unionGroupIp.Ip)
				return err
			}
		}
	}
	tx.Commit()

	return nil
}

/*
查询db中是否配置的有文件中的高危的ip
*/
func QueryHighRiskIp(ctx context.Context, ipFiles []string) error {
	logger := GlobalLogger
	dbClient := dbm.GetDBClient()
	tx := dbClient.GetCtxDb(ctx, "test_admin_mgm_db").Begin(&sql.TxOptions{})
	defer tx.Rollback()

	risked := make([]string, 0)
	for _, ipFile := range ipFiles {
		list, err := ReadHighRiskFile(ctx, ipFile)
		if err != nil {
			logger.Errorf("ReadHighRiskFile err: %s", err)
			return err
		}

		for _, ip := range list {
			ipInfo := new(dbm.TblUnionGroupIp)
			err = tx.Table("tbl_union_group_ip").Where("ip = ?", ip).Select("*").First(ipInfo).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					logger.Infoc(ctx, "检查高危ip[%s]不存在", ip)
					continue
				} else {
					logger.Errorc(ctx, "db 查询异常:%s", err)
					return err
				}
			}
			if ipInfo.Id != 0 {
				logger.Errorc(ctx, "存在高危ip入库:文件[%s] ip[%s] 安全组[%s]", ipFile, ipInfo.Ip, ipInfo.GroupName)
				risked = append(risked, ipInfo.Ip)
			}
		}
	}

	if len(risked) > 0 {
		logger.Infof("开始写入文件")
		err := WriteHighRiskIpToFile(ctx, risked)
		if err != nil {
			logger.Errorf("WriteHighRiskIpToFile err: %s", err)
			return err
		}
	}

	return nil
}

/*
读取银联的ip文件，以[]dbm.TblUnionGroupIp返回
*/
func ReadUnionXlsxFile(ctx context.Context, fileDir, ipFile string) (list []dbm.TblUnionGroupIp, err error) {
	logger := GlobalLogger

	list = make([]dbm.TblUnionGroupIp, 0)

	xlsxF := filepath.Join(fileDir, ipFile)

	xlsxFile, err := xlsx.OpenFile(xlsxF)
	if err != nil {
		logger.Errorc(ctx, "err:", err.Error())
		return nil, err
	}
	//读取每一个sheet
	for _, oneSheet := range xlsxFile.Sheets {
		//读取每个sheet下面的行数据
		for index, row := range oneSheet.Rows {
			if index == 0 {
				logger.Infof("打印标题信息：%v", row.Cells)
				continue
			}
			logger.Infof("第[%v]行body信息：%v", index+1, row.Cells)

			// 创建TblUnionGroupIp
			//if len(row.Cells) < 5 {
			//	logger.Errorc(ctx, "row非法的长度[%d]", len(row.Cells))
			//	return nil, errors.Errorf("row非法的长度[%d]", len(row.Cells))
			//}
			unionGroupIp := dbm.TblUnionGroupIp{
				Port:      row.Cells[1].String(),
				Ip:        row.Cells[2].String(),
				Policy:    row.Cells[3].String(),
				Reserve:   row.Cells[4].String(),
				GroupName: strings.Trim(ipFile, ".xlsx"),
			}
			list = append(list, unionGroupIp)
		}
	}
	return
}

/*
读取xlsx文件中的高危ip，以[]string返回
第一行为标题进行忽略，其他行取值第一个为ip
*/
func ReadHighRiskFile(ctx context.Context, ipFile string) ([]string, error) {
	logger := GlobalLogger

	xlsxFile, err := xlsx.OpenFile(ipFile)
	if err != nil {
		logger.Errorc(ctx, "err:", err.Error())
		return nil, err
	}

	list := make([]string, 0)

	//读取每一个sheet
	for _, oneSheet := range xlsxFile.Sheets {
		//读取每个sheet下面的行数据
		for index, row := range oneSheet.Rows {
			if index == 0 {
				logger.Infof("打印标题信息：%v", row.Cells)
				continue
			}
			logger.Infof("第[%v]行body信息：%v", index+1, row.Cells)
			list = append(list, row.Cells[0].String())
		}
	}

	return list, nil
}

func WriteHighRiskIpToFile(ctx context.Context, ipList []string) error {
	logger := GlobalLogger
	fileName := fmt.Sprintf("HighRiskIp_%s.xlsx", time.Now().Format("2006-01-02"))
	fileName = filepath.Join("tmp_file", fileName)

	var file *xlsx.File
	var sheet *xlsx.Sheet
	var row *xlsx.Row
	var cell *xlsx.Cell
	var err error

	file = xlsx.NewFile()
	sheet, err = file.AddSheet("Sheet1")
	if err != nil {
		logger.Infof("AddSheet err: %s", err.Error())
	}
	row = sheet.AddRow()
	cell = row.AddCell()
	cell.Value = "IP"

	for _, ip := range ipList {
		row2 := sheet.AddRow()
		cell := row2.AddCell()
		cell.Value = ip
	}

	err = file.Save(fileName)
	if err != nil {
		logger.Infof("save err: %s", err.Error())
	}

	return nil
}
