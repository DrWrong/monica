package core

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"
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
	node := getNode(root, urlPath, &handlers, kwargs)
	if node == nil {
		return nil, errors.New("404 Not Found")
	}
	fmt.Printf("%+v", handlers)
	return node.Process(handlers, kwargs), nil
}

func getNode(topNode *UrlNode, urlPath string, handlers *[]Handler, kwargs map[string]string) *UrlNode {
	if topNode.Pattern.MatchString(urlPath) {
		fmt.Printf("examin path %s, using %+v\n", urlPath, topNode)
		*handlers = append(*handlers, topNode.Filters...)
		matchers := topNode.Pattern.FindStringSubmatch(urlPath)
		for i, name := range topNode.Pattern.SubexpNames() {
			if i != 0 {
				kwargs[name] = matchers[i]
			}
		}
		urlPath = topNode.Pattern.ReplaceAllString(urlPath, "")
		if urlPath == "" {
			if topNode.Handler == nil {
				return nil
			}

			*handlers = append(*handlers, topNode.Handler)
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

func DebugRoute() {
	debugRoute(0, root)
}

func debugRoute(level int, node *UrlNode) {
	fmt.Printf(
		"%s pattern: %+v handler: %+v filters: %+v \n",
		strings.Repeat("\t", level),
		node.Pattern,
		node.Handler,
		node.Filters,
	)
	for _, subNode := range node.SubNodes {
		debugRoute(level+1, subNode)
	}
}

func Group(pattern string, fn func(), handler Handler, filters ...Handler) {
	node := &UrlNode{
		Pattern: regexp.MustCompile(pattern),
		parent:  currentNode,
		Handler: handler,
		Filters: filters,
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
