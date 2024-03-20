import sqlite3
import datetime
import json
from pathlib import Path
import matplotlib.pyplot as plt

BASE_PATH = Path(__file__).parents[1]
CDATA_PATH = BASE_PATH / "data" / "candle.json"
DB_PATH = BASE_PATH / "data.db"
OUT_PATH = Path(__file__).parent / "tweet.png"

I_OPEN_TIME = 1
I_OPEN_SIDE = 2
I_OPEN_PRICE = 3
I_CLOSE_TIME = 5
I_CLOSE_PRICE = 7
I_BALANCE = 10

def select(n):
  """HISTORYテーブルのデータを取得
  HISTORYのレコード:
    ID,
    OPEN_TIME,OPEN_SIDE,OPEN_PRICE,OPEN_AMT,
    CLOSE_TIME,CLOSE_SIDE,CLOSE_PRICE,CLOSE_AMT,
    PROFIT,BALANCE
  Params:
      n: Number of rows to select.
  Returns:
      [(col1,col2..),...]: List of table records as tuple.
  """
  con = sqlite3.connect(DB_PATH)
  cur = con.cursor()
  cur.execute("select max(id) as start from history")
  
  _max = cur.fetchone()
  _start = _max[0] - n
  start = 0 if _start < 0 else _start + 1
  
  cur.execute("select * from history where id >= (?) ",(start,))
  data = cur.fetchall() 
  return data


def slice_after(after,histories):
  for i,rec in enumerate(histories):
    opentime = rec[I_OPEN_TIME]
    if opentime > after:
      break

  return histories[i:]


def chart(data,candles):
  fig = plt.figure()
  ax = fig.add_subplot(111)
  ax.set_title("SURFER GOPHER RESULTS")
  ax.set_ylabel("BTC_JPY")

  # 左グラフ：ロウソク足をプロット
  candle_x = [unixlist_to_datelist(v) for v in candles["OpenTime"]]
  candle_y = candles["Close"]
  ax.plot(candle_x,candle_y,label="close")

  # 新規注文（buy)
  openbuy_x = []
  openbuy_y = []
  # 新規中温（sell）
  opensell_x = []
  opensell_y = []
  # 決済注文
  close_x = []
  close_y = []
  # 残高推移用
  balance_x = []
  balance_y = []

  for rec in data:
    open_side = rec[I_OPEN_SIDE]
    # 新規取引用
    _ox = unixlist_to_datelist(rec[I_OPEN_TIME])
    _oy = rec[I_OPEN_PRICE]
    # 決済取引用
    _cx = None if rec[I_CLOSE_TIME] is None or rec[I_CLOSE_TIME] == 0 else unixlist_to_datelist(rec[I_CLOSE_TIME])
    _cy = rec[I_CLOSE_PRICE]
    
    if open_side == "BUY":
        openbuy_x.append(_ox)
        openbuy_y.append(_oy)
    else:
      opensell_x.append(_ox)
      opensell_y.append(_oy)
    
    # ポジション保有中は決済情報がDBにセットされていないため、Noneを除外
    if _cx is not None:
      # 決済注文をセット
      close_x.append(_cx)
      close_y.append(_cy)
      # 残高推移をセット
      balance_x.append(_cx)
      balance_y.append(rec[I_BALANCE])
  
  # 左グラフ: 取引情報をプロット
  ax.scatter(openbuy_x,openbuy_y, label="@openBuy",color="red")
  ax.scatter(opensell_x,opensell_y,label="@openSell",color="lime")
  ax.scatter(close_x,close_y,label="@close",facecolors="none",edgecolors="black",s=80)

  # *************************************************
  # 右グラフ
  ax2 = ax.twinx()
  ax2.set_ylabel("balance")
  ax2.plot(balance_x,balance_y,color="orange",label="totalProf")
  # *************************************************
  
  ax.ticklabel_format(style="plain",axis="y")
  ax.grid(True)
  ax.legend(loc=0)
  ax2.legend(loc=4)
  plt.gcf().autofmt_xdate()
  plt.tight_layout()
  plt.savefig(OUT_PATH)


def unixlist_to_datelist(uts):
  """unix timestamp ->date string

  Args:
    uts (float): unix timestamp

  Returns:
    string:date string
  """
  return datetime.datetime.fromtimestamp(uts)


if __name__ == "__main__":
  # 取引データ
  data = select(500)
  # ロウソク足データ
  with open(CDATA_PATH) as f:
    candles = json.load(f)
    # ファイルには500足分入っている。多いので調整
    candles = {k:v[-400:] for k,v in candles.items()}
  
  first_candle_x = candles["OpenTime"][0]
  data = slice_after(first_candle_x,data)

  chart(data,candles)
