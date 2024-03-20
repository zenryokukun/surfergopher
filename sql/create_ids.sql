/* openId */
CREATE TABLE OPEN_IDS (
    -- 注文ID
    ORDER_ID TEXT PRIMARY KEY,
    -- 取引時間（UNIX）
    U_TIME INTEGER
);

/* closeId */
CREATE TABLE CLOSE_IDS (
    -- 注文ID 複数ある場合はカンマ区切り
    ORDER_ID TEXT PRIMARY KEY,
    -- 取引時間（UNIX）
    U_TIME INTEGER 
);