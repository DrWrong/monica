package core

import (
	"errors"
	"net/http"
	"regexp"
)

var (
	root        *UrlNode
	currentNode *UrlNode
)

type UrlNode struct {
	SubNodes []*UrlNode
	parent   *UrlNode
	Pattern  *regexp.Regexp
	Handler  Handler
	Filters  []Handler
}

func (node *UrlNode) Process(handlers []Handler, kwargs map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		c := NewContext(resp, req)
		c.handlers = handlers
		c.Kwargs = kwargs
		c.run()
	}
}

func GetProcessor(urlPath string) (func(http.ResponseWriter, *http.Request), error) {
	handlers := make([]Handler, 0, 0)
	kwargs := make(map[string]string, 0)
	node := getNode(root, urlPath, handlers, kwargs)
	if node == nil {
		return nil, errors.New("404 Not Found")
	}
	return node.Process(handlers, kwargs), nil
}

func getNode(topNode *UrlNode, urlPath string, handlers []Handler, kwargs map[string]string) *UrlNode {
	if topNode.Pattern.MatchString(urlPath) {
		handlers = append(handlers, topNode.Filters...)
		matchers := topNode.Pattern.FindStringSubmatch(urlPath)
		for i, name := range topNode.Pattern.SubexpNames() {
			if i != 0 {
				kwargs[name] = matchers[i]
			}
		}
		urlPath = topNode.Pattern.ReplaceAllString(urlPath, "")
		if urlPath == "" {
			handlers = append(handlers, topNode.Handler)
			return topNode
		}
		for _, subNode := range topNode.SubNodes {
			node := getNode(subNode, urlPath, handlers, kwargs)
			if node != nil {
				return node
			}
		}
	}
	return nil
}

func AddMiddleware(handlers ...Handler) {
	root.Filters = append(root.Filters, handlers...)
}

func Group(pattern string, fn func(), h ...Handler) {
	node := &UrlNode{
		Pattern: regexp.MustCompile(pattern),
		Filters: h,
		parent:  currentNode,
	}
	currentNode.SubNodes = append(currentNode.SubNodes, node)
	currentNode = node
	fn()
	currentNode = node.parent
}

func Handle(pattern string, handlers ...Handler) {

	node := &UrlNode{
		parent:  currentNode,
		Pattern: regexp.MustCompile(pattern),
	}
	node.Handler = handlers[len(handlers)-1]
	if len(handlers) > 1 {
		node.Filters = handlers[:len(handlers)-1]
	}
	currentNode.SubNodes = append(currentNode.SubNodes, node)
}

func init() {
	root = &UrlNode{
		Pattern: regexp.MustCompile("^"),
	}
	currentNode = root
}
