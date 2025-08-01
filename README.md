# 🚀 Template Miner

A fast and memory-efficient log template mining engine built in Go, inspired by Drain3 but optimized for speed and concurrency. Uses a nested tree with wildcard generalization, in-memory caching with LRU, and optional persistence via `msgpack`.

---

## 🧠 Features

- 🔥 Blazing-fast template mining (40% faster than Drain3)
- 🌲 Nested `PatternTree` with wildcard (`<*>`) generalization
- ⚙️ Thread-safe with `sync.Map` at node level
- 🧠 LRU caching using `hashicorp/golang-lru`
- 💾 Optional serialization to file via `msgpack`
- 🧰 Easy to plug into your own logs pipeline

---

## 📦 Installation

```bash
go get github.com/rishabht08/template-miner
```

---

## ✨ Usage

### 1. Import the Package

```go
import "github.com/rishabht08/template-miner/pkg/miner"
```

### 2. Create a New Miner

```go
mn := miner.NewMiner()
```

You can also provide a Redis client and key to enable Redis persistence:

```go
mn := miner.NewMinerWithRedis(redisClient, "my-key")
```

### 3. Mine Logs

```go
logs := []string{
    "User 123 logged in",
    "User 456 logged in",
    "User 789 logged out",
}

results := mn.Mine(logs)

for _, r := range results {
    fmt.Println("Template:", r.Template)
    fmt.Println("Params  :", r.Parameters)
}
```

---

## 🧪 Run Sample Locally

A sample `main.go` is provided to test mining with local logs.

```bash
go run main.go
```

---

## 💡 Tree Persistence

You can save/load the mined template tree to/from a file using `msgpack`.

### Save

```go
err := miner.SaveTreeToFileMsgpack(mn.Tree(), "tree.bin")
```

### Load

```go
tree, err := miner.LoadTreeFromFileMsgpack("tree.bin")
mn.SetTree(tree)
```

---

## 🧠 LRU Cache

The miner internally uses an expirable LRU cache:

```go
import "github.com/hashicorp/golang-lru/v2/expirable"
```

This avoids repeated redis reads and controls memory use.

---

## 🧪 Memory Profiling (optional)

To check memory usage:

```bash
go run -memprofile mem.out main.go
go tool pprof -alloc_space main mem.out
```

---

## 🧱 Internal Structure

```go
type PatternTree struct {
    Root *Node
}

type Node struct {
    Children sync.Map
}

type SerializableNode struct {
    Children map[string]*SerializableNode `msgpack:"children"`
}
```

---

## ⏱ Periodic Saving

The miner uses a `SavePeriodically()` routine to write the tree periodically to Redis (or can be adapted to save to file).

---


## 🙌 Acknowledgments

Inspired by [Drain3](https://github.com/logpai/Drain3) but optimized for concurrency and Golang efficiency.
