package core

import (
	"bufio"
	"fmt"
	"strings"
)

func ReadBlock(node Node, part *bufio.Reader) ([]byte, error) {
	block := make([]byte, node.BlockEnd-node.BlockStart)
	blockSize := node.BlockEnd - node.BlockStart
	_, err := part.Discard(node.BlockStart)
	if err != nil {
		return []byte{}, err
	}
	block, err = part.Peek(blockSize)
	return block, err
}

func VerifyBlock(block []byte, node Node) error {
	calculatedBlockHash, err := CalculateBlockHash(block)
	if err != nil {
		return err
	}
	wantedBlockHash := node.BlockSum
	if strings.Compare(calculatedBlockHash, strings.TrimSpace(wantedBlockHash)) == 0 {
		return nil
	}
	return fmt.Errorf("Error: Node %s ranging from %d to %d does not match block", node.PrevNodeSum, node.BlockStart, node.BlockEnd)
}

func VerifyNode(node Node, nextNode Node) error {
	nodeHash, err := calculateStringHash(fmt.Sprintf("%d%d%s%s", node.BlockStart, node.BlockEnd, node.BlockSum, node.PrevNodeSum))
	if err != nil {
		return err
	}
	if strings.Compare(nodeHash, nextNode.PrevNodeSum) != 0 {
		return fmt.Errorf("Node %s is not valid!", node.PrevNodeSum)
	}
	return nil
}
