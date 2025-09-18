go run . \
  -dsn "postgres://postgres:pass@localhost:5433/ppo?sslmode=disable" \
  -out results.csv -warmups 3 -runs 100 \
  -n_start 10000 -n_end 200000 -n_step 30000 \
  -seed 42