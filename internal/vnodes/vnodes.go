// Package vnodes 提供虚拟节点管理功能。
// 虚拟节点（Virtual Nodes）用于在一致性哈希环上实现更均匀的数据分布。
package vnodes

import (
	"encoding/binary"
	"fmt"
)

// VNode 表示一个虚拟节点。
type VNode struct {
	// ID 是虚拟节点的唯一标识
	ID string
	// PhysicalNode 是该虚拟节点所属的物理节点名称
	PhysicalNode string
	// Position 是该虚拟节点在哈希环上的位置
	Position uint32
}

// Distribute 根据物理节点列表和副本因子生成虚拟节点集合。
// physicalNodes 是物理节点名称列表，replicaFactor 是每个物理节点的虚拟节点数量，
// hashFn 是用于计算位置的哈希函数。
func Distribute(physicalNodes []string, replicaFactor int, hashFn func([]byte) uint32) []VNode {
	if len(physicalNodes) == 0 || replicaFactor <= 0 {
		return nil
	}
	vnodes := make([]VNode, 0, len(physicalNodes)*replicaFactor)
	for _, node := range physicalNodes {
		for i := 0; i < replicaFactor; i++ {
			id := fmt.Sprintf("%s#%d", node, i)
			buf := make([]byte, len(id)+4)
			copy(buf, id)
			binary.LittleEndian.PutUint32(buf[len(id):], uint32(i))
			pos := hashFn([]byte(id))
			vnodes = append(vnodes, VNode{
				ID:           id,
				PhysicalNode: node,
				Position:     pos,
			})
		}
	}
	return vnodes
}

// AssignedVNodes 返回属于指定物理节点的所有虚拟节点。
func AssignedVNodes(vnodes []VNode, physical string) []VNode {
	var result []VNode
	for _, vn := range vnodes {
		if vn.PhysicalNode == physical {
			result = append(result, vn)
		}
	}
	return result
}

// RemovePhysical 从虚拟节点列表中移除指定物理节点的所有虚拟节点。
func RemovePhysical(vnodes []VNode, physical string) []VNode {
	result := make([]VNode, 0, len(vnodes))
	for _, vn := range vnodes {
		if vn.PhysicalNode != physical {
			result = append(result, vn)
		}
	}
	return result
}

// Count 返回属于指定物理节点的虚拟节点数量。
func Count(vnodes []VNode, physical string) int {
	count := 0
	for _, vn := range vnodes {
		if vn.PhysicalNode == physical {
			count++
		}
	}
	return count
}

// Positions 返回所有虚拟节点的位置列表。
func Positions(vnodes []VNode) []uint32 {
	positions := make([]uint32, len(vnodes))
	for i, vn := range vnodes {
		positions[i] = vn.Position
	}
	return positions
}
