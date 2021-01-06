#!/bin/sh
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Punc -updateSize=32000,16000,8000 | tee -a results/incremental/boosted.tsv
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Matrix -updateSize=32000,16000,8000 | tee -a results/incremental/matrix.tsv
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=DPF -updateSize=32000,16000,8000 | tee -a results/incremental/dpf.tsv
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Punc -updateSize=4000,2000,1000,500 | tee -a results/incremental/boosted.tsv
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=Matrix -updateSize=4000,2000,1000,500 | tee -a results/incremental/matrix.tsv
go run ./cmd/benchmark_incremental -numRows=3000000 -pirType=DPF -updateSize=4000,2000,1000,500 | tee -a results/incremental/dpf.tsv