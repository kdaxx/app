package app

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kdaxx/container/v3"
)

type Before interface {
	Before() []reflect.Type
}
type After interface {
	After() []reflect.Type
}

type InitNode struct {
	Bean *container.BeanDefinition
	Init any
	Name string

	Nexts []*InitNode
}

const (
	unvisited = iota
	visiting
	visited
)

func DFSSort(nodes []*InitNode) ([]*InitNode, error) {
	err := buildInitializerGraph(nodes)
	if err != nil {
		return nil, err
	}
	return topoSort(nodes)
}

func buildInitializerGraph(nodes []*InitNode) error {

	findNodes := func(targetType reflect.Type) []*InitNode {

		var result []*InitNode

		for _, node := range nodes {

			beanType := node.Bean.Type()

			// type implements
			if targetType.Kind() == reflect.Interface {

				if beanType.Implements(targetType) {
					result = append(result, node)
				}

				continue
			}
			// specify type
			if beanType == targetType {
				result = append(result, node)
			}
		}
		return result
	}

	// building dependencies route
	for _, node := range nodes {
		eager, ok := node.Init.(Before)
		if ok {
			// A Before B
			for _, targetType := range eager.Before() {
				targets := findNodes(targetType)
				if len(targets) == 0 {
					return fmt.Errorf(
						"%s InitializeBefore target %v not found",
						node.Name,
						targetType,
					)
				}
				for _, target := range targets {
					node.Nexts = append(
						node.Nexts,
						target,
					)
				}
			}
		}

		lazy, ok := node.Init.(After)
		if ok {
			// A After B
			// B -> A
			for _, targetType := range lazy.After() {
				targets := findNodes(targetType)
				if len(targets) == 0 {
					return fmt.Errorf(
						"%s InitializeAfter target %v not found",
						node.Name,
						targetType,
					)
				}
				for _, target := range targets {

					target.Nexts = append(
						target.Nexts,
						node,
					)
				}
			}
		}

	}
	return nil
}

func topoSort(nodes []*InitNode) ([]*InitNode, error) {

	state := make(map[*InitNode]int)
	stack := make([]*InitNode, 0)
	stackIndex := make(map[*InitNode]int)
	order := make([]*InitNode, 0, len(nodes))
	var dfs func(*InitNode) error

	dfs = func(node *InitNode) error {

		state[node] = visiting

		stackIndex[node] = len(stack)

		stack = append(stack, node)

		for _, next := range node.Nexts {

			switch state[next] {

			case unvisited:

				if err := dfs(next); err != nil {
					return err
				}

			case visiting:

				start := stackIndex[next]

				var sb strings.Builder

				sb.WriteString(
					"initializer dependency cycle detected:\n",
				)

				for i := start; i < len(stack); i++ {

					sb.WriteString(stack[i].Name)

					sb.WriteString(" -> ")
				}

				sb.WriteString(next.Name)

				return errors.New(sb.String())
			case visited:
				continue
			}
		}

		delete(stackIndex, node)

		stack = stack[:len(stack)-1]

		state[node] = visited

		order = append(order, node)

		return nil
	}

	for _, node := range nodes {

		if state[node] == unvisited {

			if err := dfs(node); err != nil {
				return nil, err
			}
		}
	}

	// reversed
	for i, j := 0, len(order)-1; i < j; i, j = i+1, j-1 {

		order[i], order[j] = order[j], order[i]
	}

	return order, nil
}
