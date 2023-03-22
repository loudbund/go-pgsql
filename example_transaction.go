package main

import (
	"database/sql"
	"fmt"
	"github.com/loudbund/go-pgsql/pgsql_v1"
	"log"
)

func init() {
	pgsql_v1.Init("test.conf")
}

func main() {
	runTransaction()
}

func runTransaction() {
	Db := pgsql_v1.Handle().GetDb()
	var KeyTx *sql.Tx
	if tx, err := Db.Begin(); err != nil {
		log.Panic(err)
	} else {
		KeyTx = tx
	}

	if true {
		Sql, Vals := pgsql_v1.Handle().UtilInsert("demo", map[string]interface{}{
			"status":  1,
			"debug":   "test Insert11",
			"creator": "123",
		})
		if _, err := KeyTx.Exec(pgsql_v1.UtilFormatExec(Sql), Vals...); err != nil {
			fmt.Println(err)
			_ = KeyTx.Rollback()
			return
		}
	}
	if true {
		Sql, vals := pgsql_v1.Handle().UtilUpdate("demo", map[string]interface{}{
			"id":      3,
			"status":  1,
			"debug":   "test Insert update",
			"creator": "123",
		}, map[string]interface{}{
			"id": 3,
		})
		if _, err := KeyTx.Exec(pgsql_v1.UtilFormatExec(Sql), vals...); err != nil {
			fmt.Println(err)
			_ = KeyTx.Rollback()
			return
		}
	}

	if true {
		Sql, vals := pgsql_v1.Handle().UtilDelete("demo", map[string]interface{}{
			"id": 5,
		})
		if _, err := KeyTx.Exec(pgsql_v1.UtilFormatExec(Sql), vals...); err != nil {
			fmt.Println(err)
			_ = KeyTx.Rollback()
			return
		}
	}

	if err := KeyTx.Commit(); err != nil {
		_ = KeyTx.Rollback()
	}
}
