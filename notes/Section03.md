# Section 03: B-Tree - The Ideas

## Intuitions of the B-Tree and BST

Binary tress are popular data structures for sorted data. Keeping a tree in good shape after inserting or removing keys is what "balancing" is about. N-ary tress should be used instead of binary trees to make use of the "page" (minimum unit of I/O).

B-Tress can be generalized from BSTs. Each node contains multiple keys and multiple links to its children.For a key lookup, all keys are used to decide the next child node

```
        [1,        4,        9]
        /          |          \
       v           v           v
      [1,2,3]      [4,6]       [9,11,12]
```

The balancing of a B-Tree is different from a BST. BSTs like RB trees or AVL trees are balance on the height of sub-tress (by rotation). The height of all B-Tree leaf nodes is the same, a B-Tree is balanced by the size of the nodes:

- If a node is too large to fit on one page, split into two nodes. Increases the size of the parent node and possibly increase the height of the tree if the root node as split
- If a node is too small, try merging it with a sibling
