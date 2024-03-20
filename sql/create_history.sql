/*取引履歴*/
CREATE TABLE HISTORY (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    -- オープン時間(UNIX時間）EX:1651651200
    OPEN_TIME INTEGER,
    -- オープン時の'BUY' | 'SELL'
    OPEN_SIDE TEXT,
    -- オープン時の価格
    OPEN_PRICE INTEGER,
    -- オープン時の取引量
    OPEN_AMT REAL,
    -- クローズ時間(UNIX時間）
    CLOSE_TIME INTEGER,
    -- クローズ時のBUY' | 'SELL'
    CLOSE_SIDE TEXT,
    -- クローズ時の価格
    CLOSE_PRICE INTEGER,
    -- クローズ時の取引量
    CLOSE_AMT REAL,
    -- 取引の利益。約定価格がずれるので見做し
    PROFIT INTEGER,
    -- トータル利益（残高推移）
    BALANCE INTEGER
);