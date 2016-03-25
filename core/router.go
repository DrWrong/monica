package core

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/DrWrong/monica/log"
)

var (
	root        *UrlNode
	currentNode *UrlNode
	logger      *log.MonicaLogger
)

type UrlNode struct {
	SubNodes []*UrlNode
	parent   *UrlNode
	Pattern  *regexp.Regexp
	Handlers []Handler
}

func (node *UrlNode) isLeaf() bool {
	return len(node.SubNodes) == 0
}

func newProcessor(handlers []Handler, kwargs map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(resp http.ResponseWriter, req *http.Request) {
		c := NewContext(resp, req)
		c.handlers = handlers
		c.Kwargs = kwargs
		c.run()
	}

}

func GetProcessor(urlPath string) (func(http.ResponseWriter, *http.Request), error) {
	kwargs := make(map[string]string, 0)
	handlers := make([]Handler, 0, len(root.Handlers))
	handlers = append(handlers, root.Handlers...)
	for _, node := range root.SubNodes {
		if urlDispatcher(node, urlPath, &handlers, kwargs) {
			return newProcessor(handlers, kwargs), nil
		}
	}

	return nil, errors.New("404 not found")
}

func urlDispatcher(topNode *UrlNode, urlPath string, handlers *[]Handler, kwargs map[string]string) bool {
	if topNode.Pattern.MatchString(urlPath) {
		*handlers = append(*handlers, topNode.Handlers...)
		matchers := topNode.Pattern.FindStringSubmatch(urlPath)
		for i, name := range topNode.Pattern.SubexpNames() {
			if i != 0 {
				kwargs[name] = matchers[i]
			}
		}
		if topNode.isLeaf() {
			return true
		}

		urlPath = topNode.Pattern.ReplaceAllString(urlPath, "")
		for _, subNode := range topNode.SubNodes {
			if urlDispatcher(subNode, urlPath, handlers, kwargs) {
				return true
			}
		}
	}
	return false
}

func AddMiddleware(handlers ...Handler) {
	root.Handlers = append(root.Handlers, handlers...)
}

func DebugRoute() {
	debugRoute(0, root)
}

func debugRoute(level int, node *UrlNode) {
	fmt.Printf(
		"%s pattern: %+v handlers: %+v \n",
		strings.Repeat("\t", level),
		node.Pattern,
		node.Handlers,
	)
	for _, subNode := range node.SubNodes {
		debugRoute(level+1, subNode)
	}
}

func Group(pattern string, fn func(), handlers ...Handler) {
	node := &UrlNode{
		Pattern:  regexp.MustCompile(pattern),
		parent:   currentNode,
		Handlers: handlers,
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
	node.Handlers = handlers
	currentNode.SubNodes = append(currentNode.SubNodes, node)
}

func init() {
	root = &UrlNode{}
	currentNode = root
	logger = log.GetLogger("/monica/core/router")
}
