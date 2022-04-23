package ui

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuildListener(t *testing.T) {
	// setup
	g := &global{
		buildCompleteListener: make([]chan<- bool, 0),
		listerLock:            &sync.Mutex{},
	}
	wait := make(chan bool)

	// calls and asserts
	g.NextBuildComplete(func(b bool) {
		assert.True(t, b)
		wait <- true
	})
	g.emitBuildComplete(true)

	select {
	case <-wait:
	case <-time.After(time.Millisecond):
		assert.Fail(t, "No result after 1ms")
	}
}

func TestBuildWhenAlreadyValid(t *testing.T) {
	// setup
	g := &global{
		buildCompleteListener: make([]chan<- bool, 0),
		listerLock:            &sync.Mutex{},
		hasValidBuild:         true,
	}
	wait := make(chan bool)

	g.NextBuildComplete(func(b bool) {
		assert.True(t, b)
		wait <- true
	})
	g.BuildSolution()

	select {
	case <-wait:
	case <-time.After(time.Millisecond):
		assert.Fail(t, "No result after 1ms")
	}
}
