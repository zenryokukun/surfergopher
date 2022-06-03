"""
[想定利用関数]
summary_all -> 月別の利益、月最期の総利益を集計
show_all    -> summaryをグラフ表示
--------------------------------
[概要1] Monthly Result Summary
balance.jsonから月次のレポート等をする。
balance.jsonは総損益が出力されているため、
月次に換算するためには「今月の総利益-先月の総利益」と計算する
--------------------------------

--------------------------------
[summaryの補足]
- summary_all()で取得できる。
summary["month"]([]str)    -> unixtimestampを"2022-1"のフォーマットで保存
summary["tprof]([]int)     -> 月末時点の総利益
summary["lossprof"]([]int) -> 月ごとの利益
--------------------------------

"""

import json
import datetime
import matplotlib.pyplot as plt
from pathlib import Path

BOT = "SURFER GOPHER"

# ********************************************************
# Monthly Result Summary
# ********************************************************


def show_result(su):
    """
    月別利益を棒グラフ、総利益を折れ線グラフで表示する

    Args:
        su (summary)
    """
    fig = plt.figure()
    ax = fig.add_subplot(111)
    ax.set_title(f"{BOT}:Monthly Profit/Loss")
    ax.set_xlabel("month")
    ax.set_ylabel("MONTHLY PROFIT(JPY)")
    # 棒グラフで＋と－で色分ける
    colors = ["r" if v < 0 else "b" for v in su["lossprof"]]
    # 棒グラフ
    ax.bar(su["month"], su["lossprof"], width=0.2,
           align="center", color=colors, label="monthly profit")

    # 折れ線グラフ
    ax2 = ax.twinx()
    ax2.set_ylabel("TOTAL PROFIT(JPY)")
    ax2.plot(su["month"], su["tprof"], color="lime", label="total profit")
    # 中心線を引く
    ax.hlines(0, max(su["month"]), min(su["month"]))
    # レジェンド
    ax.legend(loc=2)
    ax2.legend(loc=1)
    # xラベルを傾ける（重なり防止）
    plt.xticks(rotation=30)
    # 見切れ防止
    plt.tight_layout()
    plt.show()


def get_balance():
    fp = Path(__file__).parents[1] / "data" / "balance.json"
    with open(fp) as f:
        data = json.load(f)
        return data


def summary_add(su, data, i):
    """
    data["X"][i]をsu["month]に追加
    data["Y][i]をsu["tprof"]に追加
    Args:
        su (summary): 集約データ
        data ({"X","Y"}): balance.jsonのデータ
        i (int): index
    """
    # datetime as datetime object.
    d = datetime.datetime.fromtimestamp(data["X"][i])
    # format d as "2022-1"
    ym = f"{d.year}-{d.month}"
    # total profit
    tp = data["Y"][i]
    su["month"].append(ym)
    su["tprof"].append(tp)


def summary_monthly(su):
    """
    summaryにlossprofを追加する。
    summaryのtprofは各月末時点の総利益のため、月次の利益に換算する。
    今月のtprof-先月のtrofを今月の利益として計算する。

    Args:
        su (summary): 集約データ
    """
    su["lossprof"].append(su["tprof"][0])
    for i, t in enumerate(su["tprof"][1:], 1):
        diff = t - su["tprof"][i-1]
        su["lossprof"].append(diff)


def summary_all(show_current=False):
    """
    総利益ファイル:balance.jsonから月単位の総利益を集計する
    その月の最期のレコードを最終利益とする
    月単位の利益は、今月-前月とする

    Args:
        show_current (bool): 実行月を集計対象にするか。Falseならしない

    Returns:
        summary
    """
    data = get_balance()
    if data is None:
        return
    if len(data["X"]) == 0:
        return

    summary = {
        "month": [], "tprof": [], "lossprof": []
    }

    prev_month = datetime.datetime.fromtimestamp(data["X"][0]).month
    for i in range(len(data["X"])):
        x = data["X"][i]
        month = datetime.datetime.fromtimestamp(x).month
        if month != prev_month:
            summary_add(summary, data, i-1)
        prev_month = month

    # 最期の月が集計されないのでここで集計
    if show_current:
        summary_add(summary, data, -1)

    # 月単位の利益を追加
    summary_monthly(summary)

    return summary


# ********************************************************
# Recent Indicator Evaluation
# ********************************************************

def get_trade():
    fp = Path(__file__).parents[1] / "data" / "trade.json"
    with open(fp) as f:
        data = json.load(f)
        return data

# ========================
# working...6月分から
# ========================


def evaluate_all():
    # 全期間評価
    trades = get_trade()
    # はじめの取引はエラーなので除外。。。
    EXCLUDE = 1
    times = trades["X"][EXCLUDE:]
    sides = trades["Side"][EXCLUDE:]
    sides = [side if i % 2 == 0 else "-" for i, side in enumerate(sides)]
    actions = trades["Action"][EXCLUDE:]
    btc = trades["Y"][EXCLUDE:]

    tsets = []
    for i in range(0, len(times)-1, 2):
        _side = sides[i]
        op = btc[i]  # open price
        cp = btc[i+1]  # close price
        prof = cp - op
        if _side == "SELL":
            prof = -prof
        print(actions[i], actions[i+1])


if __name__ == "__main__":
    evaluate_all()
    # show_result(summary_all(True))
    """
    su = {
        "month": ["2022-1", "2022-2", "2022-3", "2022-4", "2022-5", "2022-6",
                  "2022-7", "2022-8", "2022-9"],
        "lossprof": [100, 80, 110, -100, 50, 100, 150, 110, -90],
        "tprof": [100, 180, 290, 190, 240, 340, 490, 600, 510]
    }
    show_result(su)
    """
