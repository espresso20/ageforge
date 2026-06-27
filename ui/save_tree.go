package ui

import (
	"sort"

	"github.com/espresso20/ageforge/game"
)

// treeRow is one rendered row of the save-lineage forest: the save it represents
// plus the box-drawing connector prefix that places it under its parent.
type treeRow struct {
	Info   game.SaveInfo
	Prefix string // tree connector prefix, e.g. "", "├─ ", "│  └─ "
}

// buildSaveTree orders saves into a depth-first lineage forest. Roots are saves
// with no ParentName, or whose ParentName is not among the saves (orphans), or
// that name themselves as parent (self-loop). Roots are ordered most-recent-first;
// each node's children likewise. Returns rows in render order with a box-drawing
// connector prefix per row.
//
// Connector scheme (standard): a node at depth d prefixes one segment per
// ancestor — "│  " when that ancestor still has a later sibling below it, else
// "   " — then its own segment: "├─ " when it has a later sibling at its level,
// else "└─ ". Depth-0 roots carry no connector (empty prefix).
//
// Cycles are broken with a visited set: a save is emitted at most once, and a
// child whose recursion would revisit an ancestor is simply not recursed into.
func buildSaveTree(saves []game.SaveInfo) []treeRow {
	// Index by name and gather children per parent. byName lets us tell a real
	// parent from an orphan's dangling ParentName.
	byName := make(map[string]game.SaveInfo, len(saves))
	for _, s := range saves {
		byName[s.Name] = s
	}

	children := make(map[string][]game.SaveInfo)
	var roots []game.SaveInfo
	for _, s := range saves {
		p := s.ParentName
		// Root when: no parent, parent missing among saves (orphan), or self-parent.
		if p == "" || p == s.Name {
			roots = append(roots, s)
			continue
		}
		if _, ok := byName[p]; !ok {
			roots = append(roots, s) // orphan — parent not present
			continue
		}
		children[p] = append(children[p], s)
	}

	recent := func(list []game.SaveInfo) {
		sort.SliceStable(list, func(i, j int) bool {
			return list[i].Timestamp.After(list[j].Timestamp)
		})
	}
	recent(roots)
	for k := range children {
		recent(children[k])
	}

	var rows []treeRow
	visited := make(map[string]bool)

	// ancestorMore holds one flag per *non-root* ancestor: whether that ancestor
	// still has a later sibling, deciding its column draws "│  " (pillar) or "   "
	// (blank). Roots contribute no column, so their direct children start with an
	// empty ancestorMore and render a bare "├─ "/"└─ " with no leading pillar.
	var walk func(node game.SaveInfo, ancestorMore []bool, isRoot, isLast bool)
	walk = func(node game.SaveInfo, ancestorMore []bool, isRoot, isLast bool) {
		if visited[node.Name] {
			return // cycle break — never emit or recurse a node twice
		}
		visited[node.Name] = true

		prefix := ""
		if !isRoot {
			for _, more := range ancestorMore {
				if more {
					prefix += "│  "
				} else {
					prefix += "   "
				}
			}
			if isLast {
				prefix += "└─ "
			} else {
				prefix += "├─ "
			}
		}
		rows = append(rows, treeRow{Info: node, Prefix: prefix})

		// The column this node passes to its descendants: roots add nothing; a
		// non-root leaves a pillar iff it still has a later sibling (!isLast).
		childAncestorMore := ancestorMore
		if !isRoot {
			childAncestorMore = append(append([]bool{}, ancestorMore...), !isLast)
		}
		kids := children[node.Name]
		for i, c := range kids {
			walk(c, childAncestorMore, false, i == len(kids)-1)
		}
	}

	for i, r := range roots {
		walk(r, nil, true, i == len(roots)-1)
	}

	// Stranded saves: a pure cycle (A↔B with no external root) leaves nodes that
	// no root reaches. Promote any still-unvisited save to a root, in input order,
	// so every save is rendered exactly once instead of vanishing.
	for _, s := range saves {
		if !visited[s.Name] {
			walk(s, nil, true, true)
		}
	}
	return rows
}
