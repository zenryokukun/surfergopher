'''
Generates tweet image from
./candle.json,./trade.json,./balance.json
'''
import json
from pathlib import Path
import matplotlib.pyplot as plt


DATA_PATH = Path(__file__).parents[1] / "data"
CDATA_PATH = DATA_PATH / "candle.json"
PDATA_PATH = DATA_PATH / "trade.json"
BDATA_PATH = DATA_PATH / "balance.json"

HERE = Path(__file__).parent
OUT_PATH = HERE / "tweet.png"

def chart(tdata,pdata,bdata,opath):
    """グラフ表示

    Args:
        tdata (dict): gmo.CandlesData
        pdata (dict): open sell point data
        bdata (dict): balance data
    """
    fig = plt.figure()
    ax = fig.add_subplot(111)
    ax.set_ylabel("BTC_JPY")
    ax.plot(tdata["OpenTime"],tdata["Close"],label="close")
    #ax.plot(tdata["OpenTime"],tdata["High"],label="high")
    #ax.plot(tdata["OpenTime"],tdata["Low"],label="low")
    
    openbuy_x = []
    openbuy_y = []
    opensell_x = []
    opensell_y = []
    close_x = []
    close_y = []
    for i in range(len(pdata["X"])):
        _x = pdata["X"][i]
        _y = pdata["Y"][i]
        if pdata["Action"][i] == "OPEN":
            if pdata["Side"][i] == "BUY":
                openbuy_x.append(_x)
                openbuy_y.append(_y)
            else:
                opensell_x.append(_x)
                opensell_y.append(_y)
        else:
            close_x.append(_x)
            close_y.append(_y)

    ax.scatter(openbuy_x,openbuy_y,label="@openBuy",color="red")
    ax.scatter(opensell_x,opensell_y,label="@openSell",color="lime")
    ax.scatter(close_x,close_y,label="@close",facecolors="none",edgecolors="black",s=80)

    ax2 = ax.twinx()
    ax2.set_ylabel("balance")
    ax2.plot(bdata["X"],bdata["Y"],color="orange",label="totalProf")
    plt.grid(True)
    ax.legend(loc=1)
    ax2.legend(loc=2)
    #plt.show()
    plt.savefig(opath)


with open(CDATA_PATH) as f:
    cdata = json.load(f)

with open(PDATA_PATH) as f:
    pdata = json.load(f)

with open(BDATA_PATH) as f:
    bdata = json.load(f)

chart(cdata,pdata,bdata,OUT_PATH)
