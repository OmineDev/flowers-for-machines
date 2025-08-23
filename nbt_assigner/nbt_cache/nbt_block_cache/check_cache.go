package nbt_block_cache

import (
	nbt_hash "github.com/OmineDev/flowers-for-machines/nbt_parser/hash"
)

// CheckCache 检索整个缓存命中系统，查询 hashNumber 是否存在。
// 返回的 structure 指示命中的结果；
// 返回的 isSetHashHit 指示命中的缓存是否是集合哈希校验和
func (n *NBTBlockCache) CheckCache(hashNumber nbt_hash.CompletelyHashNumber) (
	structure StructureNBTBlock,
	hit bool,
	isSetHashHit bool,
) {
	cache, ok := n.completelyCache[hashNumber.HashNumber]
	if ok {
		return *cache, true, false
	}

	if hashNumber.SetHashNumber == nbt_hash.SetHashNumberNotExist {
		return StructureNBTBlock{}, false, false
	}

	cache, ok = n.setHashCache[hashNumber.SetHashNumber]
	if ok {
		return *cache, true, true
	}

	return StructureNBTBlock{}, false, false
}
