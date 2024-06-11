// Copyright (c) 2024 Michael D Henderson. All rights reserved.

package lbmoves

import (
	"bytes"
	"fmt"
	"log"
	"regexp"
	"strings"
)

// hexReportToNodes converts a hex report into a linked list of nodes
// where each node contains all the arguments for each component of
// the hex report.
func hexReportToNodes(hexReport []byte, showDebug bool) (root *node) {
	if showDebug {
		log.Printf("parser: root: before split %s\n", string(hexReport))
	}

	var tail *node
	for _, component := range bytes.Split(hexReport, []byte{','}) {
		if component = bytes.TrimSpace(component); len(component) != 0 {
			if root == nil {
				root = &node{text: component}
				tail = root
			} else {
				tail.next = &node{text: component}
				tail = tail.next
			}
		}
	}

	if showDebug {
		log.Printf("parser: root: after split %s\n", printNodes(root))
	}

	// splitting like that has broke some things.
	// there are components that use commas as separators internally.
	// we need to find them and splice them back together. brute force it.
	for tmp := root; tmp != nil && tmp.next != nil; {
		if tmp.isFindQuantityItem() {
			for tmp.next.isQuantityItem() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isFordEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLakeEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isLowConiferMountainsEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isOceanEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isPassEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isPatrolledAndFound() {
			for tmp.next.isUnitId() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isRiverEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isStoneRoadEdge() {
			for tmp.next.isDirection() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		} else if tmp.isUnitId() && tmp.next.isUnitId() {
			for tmp.next.isUnitId() {
				tmp.addText(tmp.next)
				tmp.next = tmp.next.next
			}
		}
		tmp = tmp.next
	}

	if showDebug {
		log.Printf("parser: root: after consolidating %s\n", printNodes(root))
	}

	return root
}

func nodesToSteps(n *node) ([][]byte, error) {
	if n == nil {
		return nil, nil
	}
	var steps [][]byte
	for n != nil {
		text := bytes.TrimSpace(n.text)
		if len(text) != 0 {
			steps = append(steps, text)
		}
		n = n.next
	}
	return steps, nil
}

func printNodes(root *node) string {
	if root == nil {
		return "<nil>"
	}
	sb := &strings.Builder{}
	sb.WriteString("(")
	for n := root; n != nil; n = n.next {
		sb.WriteString("\n\t")
		sb.WriteString(n.String())
	}
	sb.WriteString("\n)")
	return sb.String()
}

type node struct {
	text []byte
	next *node
}

func (n *node) String() string {
	if n == nil {
		return "<nil>"
	}
	return fmt.Sprintf("(%s)", string(n.text))
}

func (n *node) addText(t *node) {
	if t == nil && len(t.text) == 0 {
		return
	}
	if len(n.text) != 0 {
		n.text = append(n.text, ' ')
	}
	n.text = append(n.text, t.text...)
}

func (n *node) isDirection() bool {
	if n == nil {
		return false
	} else if bytes.Equal(n.text, []byte{'N'}) {
		return true
	} else if bytes.Equal(n.text, []byte{'N', 'E'}) {
		return true
	} else if bytes.Equal(n.text, []byte{'S', 'E'}) {
		return true
	} else if bytes.Equal(n.text, []byte{'S'}) {
		return true
	} else if bytes.Equal(n.text, []byte{'S', 'W'}) {
		return true
	} else if bytes.Equal(n.text, []byte{'N', 'W'}) {
		return true
	}
	return false
}

func (n *node) isFindQuantityItem() bool {
	if n == nil {
		return false
	}
	return rxFindQuantityItem.Match(n.text)
}

func (n *node) isFordEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'F', 'o', 'r', 'd', ' '})
}

func (n *node) isLakeEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'O', ' '})
}

func (n *node) isLowConiferMountainsEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'L', 'c', 'm', ' '})
}

func (n *node) isOceanEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'O', ' '})
}

func (n *node) isPassEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'P', 'a', 's', 's', ' '})
}

func (n *node) isPatrolledAndFound() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'P', 'a', 't', 'r', 'o', 'l', 'l', 'e', 'd', ' ', 'a', 'n', 'd', ' ', 'f', 'o', 'u', 'n', 'd', ' '})
}

func (n *node) isQuantityItem() bool {
	if n == nil {
		return false
	}
	return rxQuantityItem.Match(n.text)
}

func (n *node) isRiverEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'R', 'i', 'v', 'e', 'r', ' '})
}

func (n *node) isStoneRoadEdge() bool {
	if n == nil {
		return false
	}
	return bytes.HasPrefix(n.text, []byte{'S', 't', 'o', 'n', 'e', ' ', 'R', 'o', 'a', 'd', ' '})
}

func (n *node) isUnitId() bool {
	if n == nil {
		return false
	}
	return rxUnitId.Match(n.text)
}

var (
	rxFindQuantityItem = regexp.MustCompile(`^Find [0-9]+ [a-zA-Z][a-zA-Z ]+`)
	rxQuantityItem     = regexp.MustCompile(`^[0-9]+ [a-zA-Z][a-zA-Z ]+`)
	rxUnitId           = regexp.MustCompile(`^[0-9][[0-9][0-9][0-9]([cefg][0-9])?`)
)
