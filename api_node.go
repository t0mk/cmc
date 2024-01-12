package main

import (
	"fmt"
	"strings"
)

type Arg struct {
	Name string
	Type string
	Desc string
	Defa string
}

func (a Arg) String() string {
	return fmt.Sprintf("%s (%s): %s (default: %s)", a.Name, a.Type, a.Desc, a.Defa)
}

type Args map[string]Arg

func (a Args) Keys() string {
	r := []string{}
	for k := range a {
		r = append(r, k)
	}
	return fmt.Sprintf("%v", r)
}

func (a Args) String() string {
	r := ""
	for _, v := range a {
		r += " - " + v.String() + "\n"
	}
	return r
}

func exactInSlice(sli []string, val string) bool {
	for _, s := range sli {
		if s == val {
			return true
		}
	}
	return false
}

func (a Args) FindByPrefix(prefix string) (string, error) {
	found := []string{}
	for k := range a {
		lowK := strings.ToLower(k)
		lowPrefix := strings.ToLower(prefix)
		if strings.HasPrefix(lowK, lowPrefix) {
			found = append(found, k)
		}
	}
	if len(found) == 0 {
		return "", fmt.Errorf("no args match prefix \"%s\", possible args: %s", prefix, a.Keys())
	}
	if len(found) > 1 {
		if exactInSlice(found, prefix) {
			return prefix, nil
		}
		return "", fmt.Errorf("more args match prefix \"%s\", possible args: %s", prefix, a.Keys())
	}
	return found[0], nil
}

type ApiNode struct {
	Label    string
	Children []*ApiNode
	Url      string
	Args     Args
}

func (n *ApiNode) String() string {
	return n.PrefixedString("")
}

func (n *ApiNode) PrefixedString(p string) string {
	if n.Children == nil {
		return fmt.Sprintf(p+"╰ %s → %s\n", n.Label, n.Url)
	}
	r := p + "╰ " + n.Label + "\n"
	for _, c := range n.Children {
		r += c.PrefixedString(p + "   ")
	}
	return r
}

func GetLabels(nodes []*ApiNode) []string {
	r := []string{}
	for _, n := range nodes {
		r = append(r, n.Label)
	}
	return r
}

func (n *ApiNode) PickNext(prefix string) (*ApiNode, error) {
	matching := []*ApiNode{}
	childLabels := []string{}
	for _, c := range n.Children {
		childLabels = append(childLabels, c.Label)
		if strings.HasPrefix(c.Label, prefix) {
			matching = append(matching, c)
		}
	}
	if len(matching) == 0 {
		return nil, fmt.Errorf("no sub-paths match prefix \"%s\", child labels are: %v", prefix, childLabels)
	}
	if len(matching) > 1 {
		return nil, fmt.Errorf("more sub-paths %s match prefix \"%s\", child labels are: %v", GetLabels(matching), prefix, childLabels)
	}
	return matching[0], nil

}

func (n *ApiNode) ExpandExpression(pathExpression string) (string, error) {
	if pathExpression == "" {
		return "", nil
	}
	sli := strings.Split(pathExpression, "/")
	r := ""

	nn := n
	var err error

	for _, s := range sli {
		nn, err = nn.PickNext(s)
		if err != nil {
			return "", err
		}
		r += "/" + nn.Label
	}
	return r, nil
}

func (n *ApiNode) GetNode(pathExpression string) (*ApiNode, error) {
	if pathExpression == "" {
		return n, nil
	}
	sli := strings.Split(pathExpression, "/")

	nn := n
	var err error

	for _, s := range sli {
		nn, err = nn.PickNext(s)
		if err != nil {
			return nil, err
		}
	}
	return nn, nil
}

func (n *ApiNode) GetLeafNode(pathExpression string) (*ApiNode, error) {
	nn, err := n.GetNode(pathExpression)
	if err != nil {
		return nil, err
	}
	if nn.Children != nil {
		nodePath, _ := api.ExpandExpression(pathExpression)
		return nil, fmt.Errorf("node at \"%s\" is not URL endpoint. It's got sub-tree:\n%s", nodePath, nn)
	}
	return nn, nil

}

func (n *ApiNode) GetLeafValue(pathExpression string) (string, error) {
	nn, err := n.GetLeafNode(pathExpression)
	if err != nil {
		return "", err
	}
	return nn.Url, nil
}
