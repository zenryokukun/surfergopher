package main

import (
	"database/sql"
	"fmt"
)

// OPEN_IDテーブルにidをinsert。
//
//	INSERT INTO OPEN_IDS (U_TIME, ORDER_ID) VALUES (?,?)`
//
// (?,?) -> id, ut
func insertOpenId(db *sql.DB, id string, ut uint64) {
	_, err := db.Exec(`INSERT INTO OPEN_IDS (U_TIME, ORDER_ID) VALUES (?,?)`, id, ut)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("insertOpenId failed. ut:%v id:%v\n", ut, id)
	}
}

// CLOSE_IDテーブルにidをinsert。
//
//	INSERT INTO OPEN_IDS (U_TIME, ORDER_ID) VALUES (?,?)`
//
// (?,?) -> conIds: idsをカンマ区切りで文字列化した値, ut: パラメタの時間
func insertCloseIds(db *sql.DB, ids []string, ut uint64) {
	conIds := ""
	for i, v := range ids {
		conIds += v
		if i < len(ids)-1 {
			conIds += ","
		}
	}

	_, err := db.Exec(`INSERT INTO CLOSE_IDS (U_TIME, ORDER_ID) VALUES (?,?)`, conIds, ut)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("insertCloseIds failed. ut:%v conIds:%v\n", ut, conIds)
	}
}
