import sys
import argparse
import pandas as pd
import matplotlib.pyplot as plt
from pathlib import Path

# === CLI ===
def parse_args():
    p = argparse.ArgumentParser(
        description="Plot median and IQR (P25–P75) per scenario vs N for DB index benchmarks."
    )
    p.add_argument("csv", nargs="?", default="results.csv", help="Path to results.csv")
    p.add_argument("--outdir", default="plots", help="Directory to save plots")
    p.add_argument("--logy", action="store_true", help="Use logarithmic Y scale")
    return p.parse_args()

# === Aggregation ===
def agg_by_op(df: pd.DataFrame, op: str) -> pd.DataFrame:
    """
    Возвращает таблицу с колонками:
    scenario, n, median, p25, p75
    """
    g = (df[df["op"] == op]
         .groupby(["scenario", "n"])["ms"]
         .agg(median="median",
              p25=lambda s: s.quantile(0.25),
              p75=lambda s: s.quantile(0.75))
         .reset_index())
    return g.sort_values(["scenario", "n"])

# === Plotting ===
def plot_op(df: pd.DataFrame, op: str, title: str, outpath: Path, logy: bool):
    """
    Рисует по одной фигуре на операцию:
    - по оси X: n
    - для каждой series (scenario): линия медианы + заштрихованный IQR
    """
    if df.empty:
        print(f"[warn] no rows for op={op}, skip")
        return

    scenario_labels = {
        "heap_no_index": "Без индекса",
        "heap_btree": "Некластеризованный",
        "heap_unique_btree": "Уникальный некластеризованный",
        "clustered_btree": "Кластерный",
    }

    scenarios = sorted(df["scenario"].unique())
    fig, ax = plt.subplots(figsize=(8, 5))

    for sc in scenarios:
        sub = df[df["scenario"] == sc].sort_values("n")
        label = scenario_labels.get(sc, sc)  # если нет в словаре, оставляем как есть
        ax.plot(sub["n"], sub["median"], marker="o", label=label)
        # IQR-заливка
        # ax.fill_between(sub["n"], sub["p25"], sub["p75"], alpha=0.2)

    ax.set_title(title)
    ax.set_xlabel("Объём N (число строк)")
    ax.set_ylabel("Время, мс")
    if logy:
        ax.set_yscale("log")
    ax.grid(True, which="both", linestyle="--", alpha=0.4)
    ax.legend(title="Сценарий", loc="best")
    fig.tight_layout()
    fig.savefig(outpath, dpi=160)
    plt.close(fig)

def main():
    args = parse_args()
    outdir = Path(args.outdir)
    outdir.mkdir(parents=True, exist_ok=True)

    df = pd.read_csv(args.csv)
    # Ожидаемые столбцы: ts, scenario, op, n, note, ms
    # На всякий случай приводим типы
    df["n"] = pd.to_numeric(df["n"], errors="coerce")
    df["ms"] = pd.to_numeric(df["ms"], errors="coerce")
    df = df.dropna(subset=["n", "ms"])

    # Карта операций → (человекочитаемый заголовок, имя файла)
    cases = {
        "insert": ("Зависимость времени выполнения вставки от числа операций", "insert_line.png"),
        "select": ("Зависимость времени выполнения чтения от числа операций", "select_line.png"),
        "delete": ("Зависимость времени выполнения удаления от числа операций", "delete_line.png"),
    }

    for op, (title, fname) in cases.items():
        g = agg_by_op(df, op)
        plot_op(g, op, title, outdir / fname, logy=args.logy)

    print(f"OK. Файлы: {', '.join(str((outdir / f).resolve()) for _, (_, f) in cases.items())}")

if __name__ == "__main__":
    main()