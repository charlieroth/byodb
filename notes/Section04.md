# B-Tree: The Practice (Part 01)

## B-Tree Node Format

A node consists of:

- A fixed-sized header containing the type of node (leaf or internal) and the number of keys
- A list of pointers to the child nodes (Used by internal nodes)
- A list of offsets pointing to each key-value pair
- Packed KV pairs

| type | nkeys | pointers      | offsets       | key-values |
| ---- | ----- | ------------- | ------------- | ---------- |
| 2B   | 2B    | `nkeys` \* 8B | `nkeys` \* 2B | ...        |

The format of the KV pair. Lengths followed by data:

| klen | vlen | key | val |
| ---- | ---- | --- | --- |
| 2B   | 2B   | ... | ... |

For simplicity, leaf nodes and internal nodes use the same format

### Offset List Details

- The offset is relative to the position
- The offset of the first KV pair is always zero, so it is not store in the list
- The offset is stored to the end of the last KV pair in the offset list, which is
  used to determine the size of the node

The offset list is used to locate the nth KV pair quickly
