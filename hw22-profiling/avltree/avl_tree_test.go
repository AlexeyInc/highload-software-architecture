package avltree

import (
	"testing"
)

func TestTree(t *testing.T) {
	tree := &AVLTree{}

	// Adding elements
	tree.Add(10, 100)
	tree.Add(20, 200)
	tree.Add(30, 300)
	tree.Add(15, 150)
	tree.Add(5, 50)
	tree.Add(25, 250)
	tree.Add(35, 350)

	// Test searching
	searchKey := 20
	foundNode := tree.Search(searchKey)
	if foundNode == nil || foundNode.Value != 200 {
		t.Errorf("Expected to find key %d with value 200, but got %+v", searchKey, foundNode)
	}

	// Test updating existing key (same key, different value)
	tree.Add(20, 999)
	updatedNode := tree.Search(20)
	if updatedNode == nil || updatedNode.Value != 999 {
		t.Errorf("Expected to find key 20 with updated value 999, but got %+v", updatedNode)
	}

	// Test updating with new key
	tree.Update(20, 40, 400)
	updatedNode = tree.Search(40)
	if updatedNode == nil || updatedNode.Value != 400 {
		t.Errorf("Expected to find updated key 40 with value 400, but got %+v", updatedNode)
	}

	// Test removing node with only right child
	tree.Add(50, 500)
	tree.Remove(40)
	if tree.Search(40) != nil {
		t.Errorf("Expected key 40 to be removed, but it still exists")
	}

	// Test AVL tree right-heavy balancing
	tree.Add(60, 600)
	tree.Add(55, 550)
	if tree.root.left != nil && tree.root.left.right != nil && tree.root.left.right.key == 55 {
		t.Errorf("Tree is not correctly rebalancing when right-heavy rotation is needed")
	}

	// Test removing node with no children
	tree.Remove(15)
	if tree.Search(15) != nil {
		t.Errorf("Expected key 15 to be removed, but it still exists")
	}

	// Test removing node with left and right children
	tree.Remove(10)
	if tree.Search(10) != nil {
		t.Errorf("Expected key 10 to be removed, but it still exists")
	}

	// Test left-heavy AVL balancing (case where child is right-heavy)
	tree = &AVLTree{}
	tree.Add(50, 500)
	tree.Add(30, 300)
	tree.Add(70, 700)
	tree.Add(20, 200)
	tree.Add(40, 400)
	tree.Add(35, 350)

	// This should trigger a left-right rotation on node 30
	if tree.root.left != nil && tree.root.left.key == 35 {
		t.Errorf("Tree is not correctly rebalancing when left-heavy rotation is needed")
	}
}
