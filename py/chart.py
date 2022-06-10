'''
Generates tweet image from
./candle.json,./trade.json,./balance.json
'''
import json
import datetime
from pathlib import Path
import matplotlib.pyplot as plt


DATA_PATH = Path(__file__).parents[1] / "data"
CDATA_PATH = DATA_PATH / "candle.json"
PDATA_PATH = DATA_PATH / "trade.json"
BDATA_PATH = DATA_PATH / "balance.json"

HERE = Path(__file__).parent
OUT_PATH = HERE / "tweet.png"


def chart(tdata, pdata, bdata, opath):
    """グラフ表示

    Args:
        tdata (dict): gmo.CandlesData
        pdata (dict): open sell point data
        bdata (dict): balance data
    """

    fig = plt.figure()
    ax = fig.add_subplot(111)
    ax.set_title("SURFER GOPHER RESULTS")
    ax.set_ylabel("BTC_JPY")

    # unix timestamp -> jst datestring
    tdata_x_date = unixlist_to_datelist(tdata["OpenTime"])

    ax.plot(tdata_x_date, tdata["Close"], label="close")

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

    # unix timestamp -> jst datestring
    openbuy_x_date = unixlist_to_datelist(openbuy_x)
    opensell_x_date = unixlist_to_datelist(opensell_x)
    close_x_date = unixlist_to_datelist(close_x)

    ax.scatter(openbuy_x_date, openbuy_y, label="@openBuy", color="red")
    ax.scatter(opensell_x_date, opensell_y, label="@openSell", color="lime")
    ax.scatter(close_x_date, close_y, label="@close",
               facecolors="none", edgecolors="black", s=80)

    ###########################################
    # 右グラフ
    ax2 = ax.twinx()
    ax2.set_ylabel("balance")

    # unix timestamp -> jst datestring
    bdata_x_date = unixlist_to_datelist(bdata["X"])

    ax2.plot(bdata_x_date, bdata["Y"], color="orange", label="totalProf")

    ax.grid(True)
    ax.legend(loc=1)
    ax2.legend(loc=2)

    '''日付を間引きして表示。データが増えたらコメントはずす。
    for i, lbl in enumerate(ax.get_xticklabels()):
        if i % 2 != 0:
            lbl.set_visible(False)
    '''
    ax.ticklabel_format(style="plain", axis="y")
    plt.gcf().autofmt_xdate()  # 日付を縦表記にする
    plt.tight_layout()  # ラベルが見切れるの防止するために必要
    plt.savefig(opath)  # 画像として保存


def unixlist_to_datelist(uts):
    """[]unix timestamp ->[]date string

    Args:
        uts (float): unix timestamp

    Returns:
        []string:date string
    """
    ret = []
    for u in uts:
        ret.append(datetime.datetime.fromtimestamp(u))
    return ret


if __name__ == "__main__":

    with open(CDATA_PATH) as f:
        cdata = json.load(f)

    with open(PDATA_PATH) as f:
        pdata = json.load(f)

    with open(BDATA_PATH) as f:
        bdata = json.load(f)

    # candlesのほうが長い場合、長さをblenに合わせる
    blen = len(bdata["X"])
    clen = len(cdata["Open"])
    if clen > blen:
        for key in cdata.keys():
            cdata[key] = cdata[key][-blen:]

    chart(cdata, pdata, bdata, OUT_PATH)
