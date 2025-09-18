import sys
import argparse
import pandas as pd
from pathlib import Path
from typing import List, Tuple

# ---------- CLI ----------

def parse_args():
    p = argparse.ArgumentParser(
        description="Convert CSV to LaTeX tables. Supports raw table or pivot by op/scenario for benchmark results."
    )
    p.add_argument("csv", help="Path to input CSV (e.g., results.csv)")
    p.add_argument("--out", default="-", help="Output .tex file (default: stdout)")
    p.add_argument("--mode", choices=["raw", "pivot"], default="pivot",
                   help="raw: direct CSV->LaTeX; pivot: N×scenario per op with stats (for benchmark results)")
    # RAW options
    p.add_argument("--raw-cols", nargs="*", help="Subset of columns to keep (raw mode)")
    p.add_argument("--index", action="store_true", help="Show DataFrame index in LaTeX")
    # PIVOT options (expecting columns: ts,scenario,op,n,note,ms)
    p.add_argument("--ops", nargs="*", help="Filter operations (default: all present)")
    p.add_argument("--stat", choices=["median","p25","p75","median_iqr"], default="median",
                   help="Which statistic to show in pivot cells")
    p.add_argument("--round", type=int, default=2, help="Round numeric values to N decimals (fixed-point)")
    p.add_argument("--na", default="–", help="NA representation in LaTeX cells")
    # LaTeX cosmetics
    p.add_argument("--caption", default=None, help="Caption (raw mode only; in pivot it’s auto)")
    p.add_argument("--label", default=None, help="Label (raw mode only; in pivot it’s auto)")
    p.add_argument("--longtable", action="store_true", help="Use longtable environment")
    p.add_argument("--column-format", default=None,
                   help="LaTeX column format string, e.g. 'lrrrr' (overrides automatic)")
    return p.parse_args()

# ---------- Helpers ----------

def quantiles(s: pd.Series) -> Tuple[float,float,float]:
    return (s.median(), s.quantile(0.25), s.quantile(0.75))

def format_fixed(x, ndigits: int):
    """Фиксированный формат с сохранением хвостовых нулей, либо None для NaN."""
    if pd.isna(x):
        return None
    try:
        return f"{float(x):.{ndigits}f}"
    except Exception:
        return x

def latex_table(df: pd.DataFrame, caption=None, label=None, index=False,
                longtable=False, na_rep="–", column_format=None) -> str:
    return df.to_latex(index=index,
                       escape=True,
                       na_rep=na_rep,
                       longtable=longtable,
                       caption=caption,
                       label=label,
                       column_format=column_format)

def ensure_numeric(df: pd.DataFrame, cols: List[str]) -> pd.DataFrame:
    for c in cols:
        if c in df.columns:
            df[c] = pd.to_numeric(df[c], errors="coerce")
    return df

# ---------- RAW MODE ----------

def run_raw(df: pd.DataFrame, args) -> str:
    if args.raw_cols:
        missing = [c for c in args.raw_cols if c not in df.columns]
        if missing:
            raise SystemExit(f"[error] columns not found: {missing}")
        df = df[args.raw_cols]

    colfmt = args.column_format
    if colfmt is None:
        # Автоформат: если печатаем индекс — начальный 'l', далее r для числовых и l для остальных
        prefix = 'l' if args.index else ''
        body = ''.join('r' if pd.api.types.is_numeric_dtype(df[c]) else 'l' for c in df.columns)
        colfmt = prefix + body

    # При желании можно округлять и в raw, но обычно pivot-режима достаточно
    return latex_table(df, caption=args.caption, label=args.label,
                       index=args.index, longtable=args.longtable,
                       na_rep=args.na, column_format=colfmt)

# ---------- PIVOT MODE (benchmark-friendly) ----------

def run_pivot(df: pd.DataFrame, args) -> str:
    # Expect columns: ts, scenario, op, n, note, ms
    required = {"scenario","op","n","ms"}
    if not required.issubset(set(df.columns)):
        raise SystemExit(f"[error] pivot mode requires columns {sorted(required)}; got: {list(df.columns)}")

    df = df.copy()
    df["n"] = pd.to_numeric(df["n"], errors="coerce")
    df["ms"] = pd.to_numeric(df["ms"], errors="coerce")
    df = df.dropna(subset=["n","ms"])
    if args.ops:
        df = df[df["op"].isin(args.ops)]

    # group by (scenario, n) to compute stats
    g = (df.groupby(["op","scenario","n"])["ms"]
            .agg(median="median", p25=lambda s: s.quantile(0.25), p75=lambda s: s.quantile(0.75))
            .reset_index())

    # Build LaTeX per operation
    latex_parts = []
    ops = list(g["op"].unique())
    for op in ops:
        sub = g[g["op"] == op].copy()

        # округление и отображение значений
        if args.stat == "median":
            sub["val"] = sub["median"].apply(lambda x: format_fixed(x, args.round))
        elif args.stat == "p25":
            sub["val"] = sub["p25"].apply(lambda x: format_fixed(x, args.round))
        elif args.stat == "p75":
            sub["val"] = sub["p75"].apply(lambda x: format_fixed(x, args.round))
        elif args.stat == "median_iqr":
            m = sub["median"].apply(lambda x: format_fixed(x, args.round))
            q1 = sub["p25"].apply(lambda x: format_fixed(x, args.round))
            q3 = sub["p75"].apply(lambda x: format_fixed(x, args.round))
            sub["val"] = m.combine(q1, lambda a,b: (a,b)).combine(
                q3, lambda ab,c: f"{ab[0]} [{ab[1]}–{c}]")
        else:
            raise ValueError("unknown stat")

        # pivot: rows=n, cols=scenario
        tbl = sub.pivot(index="n", columns="scenario", values="val").sort_index()
        # Optional nicer column order: keep natural sorted order of scenarios
        tbl = tbl[sorted(tbl.columns, key=lambda x: str(x))]

        # Правое выравнивание числовых столбцов: index (n) = 'l', далее все сценарии = 'r'
        colfmt = args.column_format or ('l' + 'r' * len(tbl.columns))

        caption = f"{op}: время выполнения (ms), статистика = {args.stat}"
        label = f"tab:{op}_{args.stat}"
        part = latex_table(tbl, caption=caption, label=label,
                           index=True, longtable=args.longtable,
                           na_rep=args.na, column_format=colfmt)
        latex_parts.append(part)

    # Join tables with a blank line
    return "\n\n".join(latex_parts)

# ---------- MAIN ----------

def main():
    args = parse_args()
    df = pd.read_csv(args.csv)
    if args.mode == "raw":
        out = run_raw(df, args)
    else:
        out = run_pivot(df, args)

    if args.out == "-" or args.out == "":
        sys.stdout.write(out)
    else:
        Path(args.out).write_text(out, encoding="utf-8")
        print(f"[ok] wrote {args.out}")

if __name__ == "__main__":
    main()