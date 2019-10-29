package command

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTheSameValues(t *testing.T) {

	// Given
	arr := []float64{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}

	// When
	areTheSame := allValuesTheSame(arr)

	// Then
	assert.True(t, areTheSame)
}

func TestNotTheSameValues(t *testing.T) {

	// Given
	arr := []float64{1.0, 2.0, 5.0, 1.4, 1.6, 1.0}

	// When
	areTheSame := allValuesTheSame(arr)

	// Then
	assert.False(t, areTheSame)
}

func TestEmptyArray(t *testing.T) {

	// Given
	var arr []float64

	// When
	areTheSame := allValuesTheSame(arr)

	// Then
	assert.False(t, areTheSame)
}

func TestTooShortArray(t *testing.T) {

	// Given
	arr := []float64{1.0}

	// When
	areTheSame := allValuesTheSame(arr)

	// Then
	assert.False(t, areTheSame)
}
