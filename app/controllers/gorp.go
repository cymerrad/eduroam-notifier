package controllers

import (
	"database/sql"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"

	sq "gopkg.in/Masterminds/squirrel.v1"
)

var (
	Dbm *gorp.DbMap
)

type GorpController struct {
	*revel.Controller
	SqlStatementBuilder sq.StatementBuilderType
	Txn                 *gorp.Transaction
}

func (c *GorpController) Begin() revel.Result {
	txn, err := Dbm.Begin()
	if err != nil {
		panic(err)
	}
	c.Txn = txn
	return nil
}

func (c *GorpController) Commit() revel.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Commit(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}

func (c *GorpController) Rollback() revel.Result {
	if c.Txn == nil {
		return nil
	}
	if err := c.Txn.Rollback(); err != nil && err != sql.ErrTxDone {
		panic(err)
	}
	c.Txn = nil
	return nil
}
