package miner

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis"
	"github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/rishabht08/template-miner/pkg/miner/utils"
	"github.com/vmihailenco/msgpack/v5"
)

type Miner struct {
	Config  Config
	tree    *PatternTree
	cache   *expirable.LRU[string, *PatternTree]
	cacheMu sync.Mutex

	redis    *redis.Client
	redisKey string
	useRedis bool
	stopChan chan struct{}
	changed  atomic.Bool // Tracks if tree was modified
}

var tree *PatternTree

func NewMiner(redisClient *redis.Client, redisKey string, config Config) (*Miner, error) {

	cache := expirable.NewLRU[string, *PatternTree](20, nil, time.Hour*1) // 1 item with expiration
	miner := &Miner{
		cache:    cache,
		redis:    redisClient,
		redisKey: redisKey,
		useRedis: true,
		stopChan: make(chan struct{}),
	}
	// Load tree from Redis if available
	miner.cacheMu.Lock()
	if redisClient != nil {
		tree = loadTreeFromRedis(redisClient, redisKey)
	} else {
		tree = &PatternTree{Root: &Node{Children: sync.Map{}}}
		miner.tree = tree
		miner.cache.Add(redisKey, miner.tree)
	}

	miner.cacheMu.Unlock()

	// Start background saver
	go miner.savePeriodically(5 * time.Minute)

	return miner, nil
}

func (mn *Miner) Train(logs []string) []MinedTemplate {

	var results []MinedTemplate

	for _, log := range logs {
		tokens := utils.Tokenize(log, mn.Config.RegexMap)
		template, params, templateID := mn.tree.AddOrMatch(tokens)
		mn.changed.Store(true)
		paramID := utils.Sha1Hex(strings.Join(params, "|") + "|" + templateID)

		results = append(results, MinedTemplate{
			OriginalLog: log,
			Template:    strings.Join(template, " "),
			TemplateID:  templateID,
			Parameters:  params,
			ParamID:     paramID,
			Tokens:      tokens,
		})
	}

	mn.cacheMu.Lock()
	mn.cache.Add(mn.redisKey, mn.tree)
	mn.cacheMu.Unlock()

	return results
}

func (mn *Miner) Parse(log string) (MinedTemplate, error) {

	if mn.tree == nil {
		return MinedTemplate{}, fmt.Errorf("pattern tree not initialized")
	}
	tokens := utils.Tokenize(log, mn.Config.RegexMap)
	template, params, templateID := mn.tree.GetPattern(tokens)
	paramID := utils.Sha1Hex(strings.Join(params, "|") + "|" + templateID)

	return MinedTemplate{
		OriginalLog: log,
		Template:    strings.Join(template, " "),
		TemplateID:  templateID,
		Parameters:  params,
		ParamID:     paramID,
		Tokens:      tokens,
	}, nil
}

func loadTreeFromRedis(client *redis.Client, key string) *PatternTree {
	data, err := client.Get(key).Bytes()
	if err != nil {
		return nil
	}

	var tree PatternTree
	err = msgpack.Unmarshal(data, &tree)
	if err != nil {
		log.Println("msgpack unmarshal error:", err)
		return nil
	}
	return &tree
}

func (m *Miner) SaveTreeToRedis() error {
	if m.redis == nil {
		return fmt.Errorf("redis client not initialized")
	}
	data, err := msgpack.Marshal(m.tree)
	if err != nil {
		return err
	}

	return m.redis.Set(m.redisKey, data, 12*time.Hour).Err()
}

func (mn *Miner) savePeriodically(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if mn.changed.Load() {
				if err := mn.SaveTreeToRedis(); err != nil {
					log.Println("Failed to save tree to Redis:", err)
				} else {
					mn.changed.Store(false)
				}
			}
		case <-mn.stopChan:
			// Final save on shutdown
			if mn.changed.Load() {
				if err := mn.SaveTreeToRedis(); err != nil {
					log.Println("Failed to save tree to Redis:", err)
				}
			}
			return
		}
	}
}

func (mn *Miner) Close() {
	close(mn.stopChan)
}

type CacheEntry struct {
	Key   string
	Value interface{}
}

func SaveLRUToFile(cache *expirable.LRU[string, *PatternTree], filename string) error {
	var entries []CacheEntry

	allKeys := cache.Keys()

	for _, key := range allKeys {
		val, ok := cache.Get(key)
		if ok {
			entries = append(entries, CacheEntry{Key: key, Value: val})
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := msgpack.NewEncoder(f)
	return encoder.Encode(entries)
}

func SaveTreeToFileMsgpack(tree *PatternTree, path string) error {
	serial := tree.ToSerializable()
	data, err := msgpack.Marshal(serial)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadTreeFromFileMsgpack(path string) (*PatternTree, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var serial SerializablePatternTree
	if err := msgpack.Unmarshal(data, &serial); err != nil {
		return nil, err
	}
	return FromSerializableTree(&serial), nil
}
