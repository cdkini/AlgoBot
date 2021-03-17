TAGS = [
    "Array",
    "Backtracking",
    "Binary Search",
    "Bit Manipulation",
    "Breadth-first Search",
    "Depth-first Search",
    "Design",
    "Divide and Conquer",
    "Dynamic Programming",
    "Graph",
    "Greedy",
    "Hash Table",
    "Heap",
    "Linked List",
    "Math",
    "Recursion",
    "Sliding Window",
    "Sort",
    "Stack",
    "String",
    "Tree",
    "Trie",
    "Two Pointers",
    "Union Find"
]

def add_tags(questions):
    for tag in TAGS:
        path = _get_tag_path(tag)
        with open(path) as f:
            content = [c.strip() for c in f.readlines()]
            for row in content:
                if row.isnumeric() and int(row) in questions:
                    questions[int(row)].tags.append(tag)
    print(f"Successfully classified all questions under {len(TAGS)} tags")


def _get_tag_path(tag):
    arr = tag.split(" ")
    joined = "-".join(word.lower() for word in arr)
    return f"tags/{joined}.txt"
