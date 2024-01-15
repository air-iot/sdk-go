package sql

import (
	"fmt"

	"github.com/air-iot/logger"
	"github.com/jmoiron/sqlx"
)

type DBConn struct {
	*sqlx.DB
}

func NewDB(driverName, url string, maxIdleConn, maxOpenConn int) (*DBConn, error) {
	db, err := sqlx.Open(driverName, url)
	if err != nil {
		return nil, fmt.Errorf("DB连接错误,%s", err.Error())
	}
	db.SetMaxIdleConns(maxIdleConn)
	db.SetMaxOpenConns(maxOpenConn)
	return &DBConn{DB: db}, nil
}

func (p *DBConn) Close() {
	if err := p.DB.Close(); err != nil {
		logger.Errorln("cli错误:", err.Error())
	}
}
