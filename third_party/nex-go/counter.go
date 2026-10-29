package nex

import (
	"golang.org/x/exp/constraints"
)

type numeric interface {
	constraints.Integer | constraints.Float | constraints.Complex
}


type Counter[T numeric] struct {
	Value T
}


func (c *Counter[T]) Next() T {
	c.Value++
	return c.Value
}


func NewCounter[T numeric](start T) *Counter[T] {
	return &Counter[T]{Value: start}
}
