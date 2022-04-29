from cgi import test
import json
import matplotlib.pyplot as plt
from pathlib import Path


DATA_PATH = Path(__file__).parent
CDATA_PATH = DATA_PATH / "testdata4hr.json"
PDATA_PATH = DATA_PATH / "pos.json"
BDATA_PATH = DATA_PATH / "bal.json"

HERE = Path(__file__).parent
OUT_PATH = HERE / "bktest.png"

def chart(tdata,pdata,bdata,opath,save=False):
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
    ax.scatter(opensell_x,opensell_y,label="@openSell")
    ax.scatter(close_x,close_y,label="@close",color="lime")

    ax2 = ax.twinx()
    ax2.set_ylabel("balance")
    ax2.plot(bdata["X"],bdata["Y"],color="orange",label="totalProf")
    plt.grid(True)
    ax.legend(loc=1)
    ax2.legend(loc=2)
    if save:
        plt.savefig(opath)
    else:
        plt.show()


with open(CDATA_PATH) as f:
    cdata = json.load(f)

with open(PDATA_PATH) as f:
    pdata = json.load(f)

with open(BDATA_PATH) as f:
    bdata = json.load(f)

chart(cdata,pdata,bdata,OUT_PATH)
