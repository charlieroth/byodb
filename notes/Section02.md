# Section 02: Indexing

Almost all queries can be broken down into three types:

- Scan the whole dataset (No index used)
- Point query: Query the index by a specific key
- Range query: Query the index by a range of keys

Databases indexes are mostly about range & point queries

## Hashtables

Hashtables cannot be used for a general-purpose KV store. They are not ordered, and they are not sorted. Real world applications usually
require ordered.

The resizing operation of hashtables is an expensive operation. Naive resiszing being `O(n)` runtime.

## B-Trees

Balance binary trees can be queried in `O(log(n))` time and can be range-queried. A B-Tree is roughly a balanced n-ary tree.

Why use an n-ary tree instead of a binary tree?

- Less space overhead. On average, each leaf node requires 1~2 pointers.
- Faster in memory. Mordern CPU memory caching and other factors, B-trees can be faster than binary trees even if big-O complexity is the same.
- Less disk I/O. B-trees are shorter, which means fewer disk seeks

## LSM-Trees

LSM-Trees, or Log-structured merge trees

Querying:

1. An LSM-tree contains multiple levels of data
2. Each level is sorted and split into multiple files
3. A point query starts at the top level, if the key is not found, search next level
4. Range query merges the results from all levels, higher levels are merged first

Updating:

5. Key is inserted into a file from the top level first
6. If size of file exceeds a threshold, merge it with the next level
7. File size threshold increases exponentially with each level, which means that the amount of data also increases exponentially

Queries In Action:

1. Each level is sorted, keys can be found via binary search; range queries are just sequential file I/O, therefore effiecient

Updates In Action:

2. Top-level file size is small; inserting into top level requires small amount of I/O
3. Data is eventually merged to a lower level. Merging is sequential I/O
4. Higher levels trigger merging more often, but the merge is also smaller
5. When merging a file into lower level, any lower files whose range intersects are replaced by the merge results (which can be multiple files)
6. Merges can be done in the background. However, low-level merging can suddenly cause high I/O usage, which can degrage system perf.
