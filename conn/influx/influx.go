package influx

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"time"

	_ "github.com/influxdata/influxdb1-client" // this is important because of the bug in go mod
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var Client client.Client
var UDPClient client.Client
var DB = "tsdb"
var Mod = "http"

// New 创建客户端
func Init() {
	var (
		host     = viper.GetString("influx.host")
		port     = viper.GetInt("influx.port")
		udpPort  = viper.GetInt("influx.udpPort")
		username = viper.GetString("influx.username")
		password = viper.GetString("influx.password")
	)
	Mod = viper.GetString("influx.mode")
	DB = viper.GetString("influx.db")
	var err error
	Client, err = client.NewHTTPClient(client.HTTPConfig{
		Addr:     fmt.Sprintf("http://%s:%d", host, port),
		Username: username,
		Password: password,
	})
	if err != nil {
		logrus.Panic(err)
	}
	if Mod == "udp" {
		UDPClient, err = client.NewUDPClient(client.UDPConfig{
			Addr: net.JoinHostPort(host, strconv.Itoa(udpPort)),
		})
		if err != nil {
			logrus.Panic(err)
		}
	}
}

func Close() {
	if Client != nil {
		err := Client.Close()
		if err != nil {
			logrus.Errorln("Influx关闭失败", err.Error())
		}
	}

	if UDPClient != nil {
		err := UDPClient.Close()
		if err != nil {
			logrus.Errorln("Influx UDP关闭失败", err.Error())
		}
	}
}

// CreateDB is 创建数据库
func CreateDB(c client.Client, database ...string) (err error) {
	if database != nil && len(database) > 0 {
		for _, db := range database {
			_, err = QueryDB(c, fmt.Sprintf("CREATE DATABASE %s", db), db)
			if err != nil {
				return fmt.Errorf("创建数据库失败.%v", err)
			}
		}
	}
	return nil
}

// QueryDB convenience function to query the database
func QueryDB(c client.Client, cmd, database string) ([]client.Result, error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	response, err := c.Query(q)
	if err != nil {
		return nil, err
	} else if response.Error() != nil {
		return nil, response.Error()
	}
	return response.Results, nil
}

// QueryList convenience function to query the database
func QueryList(c client.Client, cmd, database string) ([]map[string]interface{}, error) {
	q := client.Query{
		Command:  cmd,
		Database: database,
	}
	response, err := c.Query(q)
	if err != nil {
		return nil, err
	} else if response.Error() != nil {
		return nil, response.Error()
	}
	results := make([]map[string]interface{}, 0)
	res := response.Results
	if res != nil && len(res) > 0 && len(res[0].Series) > 0 && len(res[0].Series[0].Values) > 0 {
		results = make([]map[string]interface{}, len(res[0].Series[0].Values))
		//tags :=  res[0].Series[0].Tags
		columns := res[0].Series[0].Columns
		values := res[0].Series[0].Values
		for h, value := range values {
			if len(columns) != len(value) {
				return nil, errors.New("查询错误,列与数据长度不匹配")
			}
			data := make(map[string]interface{})
			for i := 0; i < len(columns); i++ {
				data[columns[i]] = values[i]
			}
			results[h] = data
		}
	}
	return results, nil
}

// CreateRetentionPolicy is 创建保留策略
func CreateRetentionPolicy(c client.Client, database, policyName, duration string, replication int) error {
	_, err := QueryDB(c, fmt.Sprintf("CREATE RETENTION POLICY \"%s\" ON %s DURATION %s REPLICATION %d", policyName, database, duration, replication), database)
	if err != nil {
		return fmt.Errorf("创建保留策略失败.%v", err)
	}
	return nil
}

// WritePointsRetention is 添加数据
func WritePointsRetention(c client.Client, database, precision, policyName, measurement string, tags map[string]string, fields map[string]interface{}, timestamp time.Time) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        database,
		Precision:       precision,
		RetentionPolicy: policyName,
	})
	if err != nil {
		return fmt.Errorf("新建批量添加操作失败.%s", err.Error())
	}

	pt, err := client.NewPoint(
		measurement,
		tags,
		fields,
		timestamp,
	)
	if err != nil {
		return fmt.Errorf("新建数据操作失败.%s", err.Error())
	}
	bp.AddPoint(pt)

	if err := c.Write(bp); err != nil {
		return fmt.Errorf("存储失败.%s", err.Error())
	}
	return nil
}

// WritePointsRetentionBatch is 新建批量添加操作
func WritePointsRetentionBatch(c client.Client, database, precision, policyName, measurement string, tags []map[string]string, fields []map[string]interface{}, timestamps []time.Time) error {
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:        database,
		Precision:       precision,
		RetentionPolicy: policyName,
	})
	if err != nil {
		return fmt.Errorf("新建批量添加操作失败.%s", err.Error())
	}

	for i := 0; i < len(tags); i++ {
		pt, err := client.NewPoint(
			measurement,
			tags[i],
			fields[i],
			timestamps[i],
		)
		if err != nil {
			return fmt.Errorf("新建数据操作失败.%s", err.Error())
		}
		bp.AddPoint(pt)
	}

	if err := c.Write(bp); err != nil {
		return fmt.Errorf("存储Point失败(Retention).%v", err)
	}
	return nil
}
