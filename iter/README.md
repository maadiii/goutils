# iter_pipeline — Chainable iterator helpers on iter.Seq

This package provides a chainable pipeline API built on Go's `iter.Seq[T]`, inspired by JavaScript iterator helpers.

## Build/run

- Run tests: `go test ./...`
- Demo binary is behind build tag `example`:
  - `go run -tags example ./problems/iter_pipeline`

## Constructors

- `FromSlice[T]([]T) Pipe[T]`
- `Range(start,end int) Pipe[int]` — [start,end)

## Core ops

- `Map`, `Filter`, `Take`, `Drop`, `Slice(start,end)`, `Concat`
- `TakeWhile`, `DropWhile`
- `MapTo[T,U](Pipe[T], func(T)U) Pipe[U]` (free function)

## Query/collect

- `Some`, `Every`, `Find`
- `ForEach`, `ToSlice`
- `Reduce`, `ReduceTo`
- `Count`, `First`, `Last`, `Nth`
- `Includes`, `IndexOf`, `IndexOfBy`

## Indexing/flattening

- `AsIndexedPairs(Pipe[T]) Pipe[Pair[T]]`
- `FlatMapTo`, `FlatMapSliceTo`, `Flatten`, `FlattenSlice`

## Distinct/group

- `Distinct`, `DistinctBy`
- `GroupBy`

## Windows/chunking

- `Chunk(size)`, `Windows(size)`

## Tail/head and ordering

- `TakeLast(n)`, `DropLast(n)`
- `Reverse`, `SortBy`

## Aggregates

- `SumInts`, `MinInt`, `MaxInt`

See tests for idiomatic usage: `pipeline_test.go`, `pipeline_more_test.go`, `pipeline_extras_test.go`, `pipeline_comprehensive_test.go`.
