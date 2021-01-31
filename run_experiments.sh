#/bin/sh

# Static DB 

go run ./cmd/benchmark_initial -numRows=3000000 -updatable=false -pirType=Punc | tee results/initial/boosted.txt
go run ./cmd/benchmark_initial -numRows=3000000 -updatable=false -pirType=DPF | tee results/initial/dpf.txt
go run ./cmd/benchmark_initial -numRows=3000000 -updatable=false -pirType=Matrix | tee results/initial/matrix.txt

# Updates cost pattern
go run ./cmd/benchmark_updates -numRows=3000000 -updateSize=30000 -numUpdates=202 -pirType=Punc | tee results/updates/boosted.txt

# Amortized costs
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Punc -updateSize=32000,4000,500 | tee results/incremental/boosted.txt
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=DPF -updateSize=32000,4000,500 | tee results/incremental/dpf.txt
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Matrix -updateSize=32000,4000,500 | tee results/incremental/matrix.txt


# Trace replay
go run ./cmd/benchmark_trace -pirType=Punc -trace=safebrowsing/trace.txt | tee results/log/boosted.txt
go run ./cmd/benchmark_trace -pirType=DPF -trace=safebrowsing/trace.txt | tee results/log/dpf.txt
go run ./cmd/benchmark_trace -pirType=NonPrivate -trace=safebrowsing/trace.txt | tee results/log/nonprivate.txt
