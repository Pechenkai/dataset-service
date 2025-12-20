package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	rand "math/rand/v2"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type cfgT struct {
	DSN     string
	Out     string
	Warmups int
	Runs    int
	NStart  int
	NEnd    int
	NStep   int
	Seed    uint64
}

func parse() cfgT {
	var c cfgT
	flag.StringVar(&c.DSN, "dsn", "postgres://postgres:pass@localhost:5433/ppo?sslmode=disable", "Postgres DSN")
	flag.StringVar(&c.Out, "out", "results.csv", "CSV output")
	flag.IntVar(&c.Warmups, "warmups", 3, "warmups per case")
	flag.IntVar(&c.Runs, "runs", 10, "measured runs per case")
	flag.IntVar(&c.NStart, "n_start", 10_000, "начальное N")
	flag.IntVar(&c.NEnd, "n_end", 200_000, "конечное N")
	flag.IntVar(&c.NStep, "n_step", 30_000, "шаг N")
	flag.Uint64Var(&c.Seed, "seed", 42, "random seed (reproducible)")
	flag.Parse()
	return c
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func mustExec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) {
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		log.Fatal(err)
	}
}

func genNs(start, end, step int) []int {
	if step <= 0 {
		step = 1
	}
	if end < start {
		end = start
	}
	var out []int
	for n := start; n <= end; n += step {
		out = append(out, n)
	}
	return out
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func timeMs(f func() error) (float64, error) {
	t0 := time.Now()
	err := f()
	return float64(time.Since(t0).Microseconds()) / 1000.0, err
}

func perm(n int, rng *rand.Rand) []int {
	p := make([]int, n)
	for i := 0; i < n; i++ {
		p[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := int(rng.UintN(uint(i + 1)))
		p[i], p[j] = p[j], p[i]
	}
	return p
}

func reloadData(ctx context.Context, pool *pgxpool.Pool, keys []int, base time.Time) error {
	mustExec(ctx, pool, "TRUNCATE bench_items RESTART IDENTITY;")
	return bulkInsert(ctx, pool, keys, base)
}

func bulkInsert(ctx context.Context, pool *pgxpool.Pool, keys []int, base time.Time) error {
	const chunk = 1000
	sb := strings.Builder{}
	args := make([]any, 0, chunk*3)
	place := 1

	for i, k := range keys {
		if i%chunk == 0 {
			if sb.Len() > 0 {
				if _, err := pool.Exec(ctx, sb.String(), args...); err != nil {
					return err
				}
				sb.Reset()
				args = args[:0]
				place = 1
			}
			sb.WriteString("INSERT INTO bench_items (k, payload, created_at) VALUES ")
		}
		if i%chunk != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("($%d,$%d,$%d)", place, place+1, place+2))
		place += 3

		args = append(args, k, "x", base.Add(time.Duration(i)*time.Millisecond).UTC())
	}
	if sb.Len() > 0 {
		if _, err := pool.Exec(ctx, sb.String(), args...); err != nil {
			return err
		}
	}
	return nil
}

func bench(w *csv.Writer, scenario, op string, n int, note string, warmups, runs int, f func() (float64, error)) {
	for i := 0; i < warmups; i++ {
		_, _ = f()
	}
	for i := 0; i < runs; i++ {
		ms, err := f()
		if err != nil {
			log.Fatal(err)
		}
		_ = w.Write([]string{time.Now().Format(time.RFC3339Nano), scenario, op, fmt.Sprint(n), note, fmt.Sprintf("%.3f", ms)})
	}
	w.Flush()
}

type scenario interface {
	Name() string
	Prepare(ctx context.Context)
	Reset(ctx context.Context)
	PostInsertLayout(ctx context.Context)
	Cleanup(ctx context.Context)
}

type noIndex struct{ pool *pgxpool.Pool }

func (s noIndex) Name() string { return "heap_no_index" }
func (s noIndex) Prepare(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}
func (s noIndex) Reset(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}
func (s noIndex) PostInsertLayout(ctx context.Context) {
	// ничего
}
func (s noIndex) Cleanup(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}

type heapBtree struct{ pool *pgxpool.Pool }

func (s heapBtree) Name() string { return "heap_btree" }
func (s heapBtree) Prepare(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}
func (s heapBtree) Reset(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
	mustExec(ctx, s.pool, `CREATE INDEX idx_bench_k ON bench_items(k);`)
}
func (s heapBtree) PostInsertLayout(ctx context.Context) {
	// просто ANALYZE, чтобы планировщик видел статистики
	mustExec(ctx, s.pool, `ANALYZE bench_items;`)
}
func (s heapBtree) Cleanup(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}

type heapUniqueBtree struct{ pool *pgxpool.Pool }

func (s heapUniqueBtree) Name() string { return "heap_unique_btree" }
func (s heapUniqueBtree) Prepare(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}
func (s heapUniqueBtree) Reset(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
	mustExec(ctx, s.pool, `CREATE UNIQUE INDEX idx_bench_k ON bench_items(k);`)
}
func (s heapUniqueBtree) PostInsertLayout(ctx context.Context) {
	mustExec(ctx, s.pool, `ANALYZE bench_items;`)
}
func (s heapUniqueBtree) Cleanup(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}

type clusteredBtree struct{ pool *pgxpool.Pool }

func (s clusteredBtree) Name() string { return "clustered_btree" }
func (s clusteredBtree) Prepare(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}
func (s clusteredBtree) Reset(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
	mustExec(ctx, s.pool, `CREATE INDEX idx_bench_k ON bench_items(k);`)
}
func (s clusteredBtree) PostInsertLayout(ctx context.Context) {
	mustExec(ctx, s.pool, `CLUSTER bench_items USING idx_bench_k;`)
	mustExec(ctx, s.pool, `ANALYZE bench_items;`)
}
func (s clusteredBtree) Cleanup(ctx context.Context) {
	mustExec(ctx, s.pool, `DROP INDEX IF EXISTS idx_bench_k;`)
}

//func main() {
//	cfg := parse()
//	ctx := context.Background()
//	pool, err := pgxpool.New(ctx, cfg.DSN)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer pool.Close()
//
//	mustExec(ctx, pool, `
//CREATE TABLE IF NOT EXISTS bench_items (
//    id BIGSERIAL PRIMARY KEY,
//    k  INT NOT NULL,
//    payload TEXT NOT NULL,
//    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
//);`)
//
//	////out := must(os.Create(cfg.Out))
//	////defer out.Close()
//	////w := csv.NewWriter(out)
//	////defer w.Flush()
//	////_ = w.Write([]string{"ts", "scenario", "op", "n", "note", "ms"})
//	//
//	//ns := genNs(cfg.NStart, cfg.NEnd, cfg.NStep)
//	//
//	//scenarios := []scenario{
//	//	noIndex{pool: pool},
//	//	heapBtree{pool: pool},
//	//	heapUniqueBtree{pool: pool},
//	//	clusteredBtree{pool: pool},
//	//}
//	//
//	//for _, s := range scenarios {
//	//	log.Printf("== Scenario: %s ==", s.Name())
//	//	s.Prepare(ctx)
//	//	mustExec(ctx, pool, "TRUNCATE bench_items RESTART IDENTITY;")
//	//	s.Reset(ctx)
//	//
//	//	for _, N := range ns {
//	//		rng := rand.New(rand.NewPCG(cfg.Seed, uint64(N)))
//	//		keys := perm(N, rng)
//	//		baseTS := time.Now().Add(-1 * time.Hour)
//	//
//	//		bench(w, s.Name(), "insert", N, fmt.Sprintf("bulk=%d", N),
//	//			cfg.Warmups, cfg.Runs, func() (float64, error) {
//	//				return timeMs(func() error {
//	//					mustExec(ctx, pool, "TRUNCATE bench_items RESTART IDENTITY;")
//	//					return bulkInsert(ctx, pool, keys, baseTS)
//	//				})
//	//			})
//	//
//	//		mustExec(ctx, pool, "TRUNCATE bench_items RESTART IDENTITY;")
//	//		if err := bulkInsert(ctx, pool, keys, baseTS); err != nil {
//	//			log.Fatal(err)
//	//		}
//	//		s.PostInsertLayout(ctx)
//	//
//	//		bench(w, s.Name(), "select", N, "k in [a,b] (~N/50, cap 1000)",
//	//			cfg.Warmups, cfg.Runs, func() (float64, error) {
//	//				width := min(1000, max(1, N/50))
//	//				a := int(rng.UintN(uint(max(1, N-width))))
//	//				b := min(N-1, a+width-1)
//	//				return timeMs(func() error {
//	//					_, err := pool.Exec(ctx, `SELECT COUNT(*) FROM bench_items WHERE k BETWEEN $1 AND $2;`, a, b)
//	//					return err
//	//				})
//	//			})
//	//
//	//		bench(w, s.Name(), "delete", N, "k%7=0",
//	//			cfg.Warmups, cfg.Runs, func() (float64, error) {
//	//				return timeMs(func() error {
//	//					mustExec(ctx, pool, "TRUNCATE bench_items RESTART IDENTITY;")
//	//					if err := bulkInsert(ctx, pool, keys, baseTS); err != nil {
//	//						return err
//	//					}
//	//					_, err := pool.Exec(ctx, `DELETE FROM bench_items WHERE k % 7 = 0;`)
//	//					return err
//	//				})
//	//			})
//	//	}
//
//		//s.Cleanup(ctx)
//	}
//
//	log.Printf("Done. Results -> %s", cfg.Out)
//}
