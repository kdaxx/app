package app

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/kdaxx/container/v3"
)

type initNode struct {
	bean *container.BeanDefinition
	init Initializer
	name string

	nexts []*initNode
}

const (
	unvisited = iota
	visiting
	visited
)

func buildInitializerGraph(nodes []*initNode) error {

	findNodes := func(targetType reflect.Type) []*initNode {

		var result []*initNode

		for _, node := range nodes {

			beanType := node.bean.Type()

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
		eager, ok := node.init.(PreInitializer)
		if ok {
			// A Before B
			for _, targetType := range eager.InitializeBefore() {
				targets := findNodes(targetType)
				if len(targets) == 0 {
					return fmt.Errorf(
						"%s InitializeBefore target %v not found",
						node.name,
						targetType,
					)
				}
				for _, target := range targets {
					node.nexts = append(
						node.nexts,
						target,
					)
				}
			}
		}

		lazy, ok := node.init.(PostInitializer)
		if ok {
			// A After B
			// B -> A
			for _, targetType := range lazy.InitializeAfter() {
				targets := findNodes(targetType)
				if len(targets) == 0 {
					return fmt.Errorf(
						"%s InitializeAfter target %v not found",
						node.name,
						targetType,
					)
				}
				for _, target := range targets {

					target.nexts = append(
						target.nexts,
						node,
					)
				}
			}
		}

	}
	return nil
}

func topoSort(nodes []*initNode) ([]*initNode, error) {

	state := make(map[*initNode]int)
	stack := make([]*initNode, 0)
	stackIndex := make(map[*initNode]int)
	order := make([]*initNode, 0, len(nodes))
	var dfs func(*initNode) error

	dfs = func(node *initNode) error {

		state[node] = visiting

		stackIndex[node] = len(stack)

		stack = append(stack, node)

		for _, next := range node.nexts {

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

					sb.WriteString(stack[i].name)

					sb.WriteString(" -> ")
				}

				sb.WriteString(next.name)

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
