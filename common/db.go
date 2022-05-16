/**
 * Tencent is pleased to support the open source community by making Polaris available.
 *
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *
 * Licensed under the BSD 3-Clause License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * https://opensource.org/licenses/BSD-3-Clause
 *
 * Unless required by applicable law or agreed to in writing, software distributed
 * under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
 * CONDITIONS OF ANY KIND, either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 */

package common

import (
	"database/sql"
	"errors"
	"github.com/golang/glog"
)

// PolarisDB 操作polaris数据库的工具类
type PolarisDB struct {
	db *sql.DB
}

// Close 关闭数据库连接
func (p *PolarisDB) Close() {
	if p.db != nil {
		_ = p.db.Close()
	}
}

// NewPolarisDB 创建polaris数据库操作类
func NewPolarisDB(source string) (*PolarisDB, error) {
	db, err := NewMysqlDB(source)
	if err != nil {
		return nil, err
	}
	return &PolarisDB{db: db}, nil
}

// NewMysqlDB 创建MySQL操作类
func NewMysqlDB(dbSource string) (*sql.DB, error) {
	if dbSource == "" {
		return nil, errors.New("database source is empty")
	}

	db, err := sql.Open("mysql", dbSource)
	if err != nil {
		glog.Errorf("open db(%s) err: %s", dbSource, err.Error())
		return nil, err
	}

	if err := db.Ping(); err != nil {
		glog.Errorf("db ping err: %s", err.Error())
		return nil, err
	}

	return db, nil
}

// LoadAllInvalidInstances 加载所有失效的实例
func (p *PolarisDB) LoadAllInvalidInstances(limitTime, limitNum int) ([]string, error) {
	str := `select instance.id from instance where mtime <= DATE_SUB(NOW(), INTERVAL ? MINUTE) and flag = 1 limit ?`
	rows, err := p.db.Query(str, limitTime, limitNum)
	if err != nil {
		glog.Errorf("[PolarisDB] load all invalid instances err: %s", err.Error())
		return nil, err
	}
	defer rows.Close()

	progress := 0
	out := make([]string, 0)
	for rows.Next() {
		progress++
		if progress%50000 == 0 {
			glog.Infof("[PolarisDB] instance fetch rows progress: %d", progress)
		}
		var id string
		err := rows.Scan(&id)
		if err != nil {
			glog.Errorf("[PolarisDB] fetch instance rows err: %s", err.Error())
			return nil, err
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		glog.Errorf("[PolarisDB] instance rows catch err: %s", err.Error())
		return nil, err
	}

	glog.Infof("[PolarisDB] get all invalid instances count: %d", len(out))
	return out, nil
}

// CleanInvalidInstanceList 清理失效的实例列表
func (p *PolarisDB) CleanInvalidInstanceList(instanceIds []string) error {
	if len(instanceIds) == 0 {
		return errors.New("missing id")
	}

	ids := make([]interface{}, 0, len(instanceIds))
	var paramStr string
	first := true
	for _, insId := range instanceIds {
		if first {
			paramStr = "(?"
			first = false
		} else {
			paramStr += ", ?"
		}
		ids = append(ids, insId)
	}
	paramStr += ")"

	str := "delete from instance where flag = 1 and id in " + paramStr
	_, err := p.db.Exec(str, ids...)
	if err != nil {
		glog.Errorf("[PolarisDB] clean invalid instance(%s) err: %s", instanceIds, err.Error())
		return err
	}

	glog.Info("[PolarisDB] clean invalid instance: ", instanceIds)

	return nil
}
