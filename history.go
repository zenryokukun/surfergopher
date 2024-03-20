package main

/*DBのHISTORYテーブルを操作するための機能*/

import (
	"database/sql"
	"fmt"
)

type (
	History struct {
		OpenTime   uint64
		OpenSide   string
		OpenPrice  uint64
		OpenAmt    float32
		CloseTime  uint64
		CloseSide  string
		ClosePrice uint64
		CloseAmt   float32
		Profit     int
		Balance    int
	}
)

// HISTORYテーブルの最大IDを取得する
// 実行されるSQL：
//
//	Select MAX(ID) FROM HISTORY WHERE CLOSE_TIME IS NULL OR CLOSE_TIME = 0
//
// 決済されるまでCLOSE_TIMEはNULLだが、移行データは0になっているため上記の条件になっている。
func selectMaxId(db *sql.DB) int {
	var id int
	row := db.QueryRow("Select MAX(ID) FROM HISTORY WHERE CLOSE_TIME IS NULL OR CLOSE_TIME = 0")
	err := row.Scan(&id)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("selectMaxId failed. id:%v\n", id)
	}
	return id
}

// HISTORYテーブルから直近のBALANCE（累計損益）を取得する
// 直近の決済済みのレコードを特定するSQL:
//
//	SELECT MAX(ID) FROM HISTORY WHERE CLOSE_TIME IS NOT NULL AND CLOSE_TIME != 0
//
// 決済されるまでCLOSE_TIMEはNULLだが、移行データは0になっているため上記の条件になっている。
func selectRecentBalance(db *sql.DB) int {
	row := db.QueryRow(`SELECT MAX(ID) FROM HISTORY WHERE CLOSE_TIME IS NOT NULL AND CLOSE_TIME != 0`)

	var id int
	err := row.Scan(&id)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("First row.Scan failed at selectRecentBalance. id:%v\n", id)
	}

	row = db.QueryRow(`
		SELECT BALANCE FROM HISTORY WHERE ID=(?)
	`, id)

	var balance int
	err = row.Scan(&balance)
	if err != nil {
		fmt.Println(err)
		fmt.Printf("Second row.Scan failed at selectRecentBalance. balance:%v\n", balance)
	}

	return balance
}

// HISTORYテーブルにレコードを挿入する。
//
//	INSERT INTO HISTORY (OPEN_TIME, OPEN_SIDE,OPEN_PRICE,OPEN_AMT) VALUES (?,?,?,?)
//
// (?,?,?,?) -> OPEN_TIME, OPEN_SIDE, OPEN_PRICE, OPEN_AMT
func insertHistory(
	db *sql.DB,
	ht *History,
) {
	_, err := db.Exec(`
		INSERT INTO HISTORY (OPEN_TIME, OPEN_SIDE,OPEN_PRICE,OPEN_AMT) VALUES (?,?,?,?)
	`, ht.OpenTime, ht.OpenSide, ht.OpenPrice, ht.OpenAmt)

	if err != nil {
		fmt.Println(err)
		fmt.Println("insertHistory failed.")
	}
}

// HISTORYテーブルのレコードを更新する
//
//	UPDATE HISTORY SET CLOSE_TIME = ?, CLOSE_SIDE = ?, CLOSE_PRICE = ?, CLOSE_AMT = ?, PROFIT = ?,BALANCE = ? WHERE ID = ?
func updateHistory(
	db *sql.DB, ht *History,
) {
	// selectでヒットが無い場合はゼロになっている。その場合何もしない。
	id := selectMaxId(db)
	if id == 0 {
		return
	}

	_, err := db.Exec(`
        UPDATE HISTORY SET
		  CLOSE_TIME = ?,
		  CLOSE_SIDE = ?,
		  CLOSE_PRICE = ?,
		  CLOSE_AMT = ?,
		  PROFIT = ?,
		  BALANCE = ?
		WHERE
		  ID = ? 
	`,
		ht.CloseTime, ht.CloseSide, ht.ClosePrice, ht.CloseAmt,
		ht.Profit, ht.Balance, id,
	)

	if err != nil {
		fmt.Println(err)
		fmt.Println("updateHistory failed.")
	}
}
