// Copyright (c) 2019 Perlin
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package wavelet

import (
	"github.com/perlin-network/wavelet/store"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRounds(t *testing.T) {
	t.Parallel()

	storage := store.NewInmem()

	rm, err := NewRounds(storage, 10)
	if !assert.EqualError(t, errors.Cause(err), "key not found") {
		return
	}

	if !assert.NotNil(t, rm) {
		return
	}

	for i := 0; i < 5; i++ {
		r := &Round{
			Index: uint64(i + 1),
			Start: Transaction{},
			End:   Transaction{},
		}
		_, err := rm.Save(r)
		assert.NoError(t, err)
	}

	assert.Equal(t, uint64(5), rm.Latest().Index)
	assert.Equal(t, uint64(1), rm.Oldest().Index)

	newRM, err := NewRounds(storage, 10)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, rm.latest, newRM.latest)
	assert.Equal(t, rm.oldest, newRM.oldest)
	assert.Equal(t, len(rm.buffer), len(newRM.buffer))

	assert.Equal(t, uint64(5), newRM.Latest().Index)
	assert.Equal(t, uint64(1), newRM.Oldest().Index)
	assert.Equal(t, uint64(5), newRM.Count())
}

func TestRoundsCircular(t *testing.T) {
	t.Parallel()

	storage := store.NewInmem()

	rm, err := NewRounds(storage, 10)
	if !assert.EqualError(t, errors.Cause(err), "key not found") {
		return
	}

	if !assert.NotNil(t, rm) {
		return
	}

	for i := 0; i < 15; i++ {
		r := &Round{
			Index: uint64(i + 1),
			Start: Transaction{},
			End:   Transaction{},
		}
		_, err := rm.Save(r)
		assert.NoError(t, err)
	}

	assert.Equal(t, uint64(15), rm.Latest().Index)
	assert.Equal(t, uint64(6), rm.Oldest().Index)

	newRM, err := NewRounds(storage, 10)
	if !assert.NoError(t, err) {
		return
	}

	assert.Equal(t, rm.latest, newRM.latest)
	assert.Equal(t, rm.oldest, newRM.oldest)
	assert.Equal(t, len(rm.buffer), len(newRM.buffer))

	assert.Equal(t, uint32(4), newRM.latest)
	assert.Equal(t, uint32(5), newRM.oldest)
}
