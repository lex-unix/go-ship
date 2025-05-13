package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMergeMaps(t *testing.T) {
	t.Run("does not overwrite existing keys in dst", func(t *testing.T) {
		dst := map[string]any{
			"key1": "val1",
			"key2": "val2",
			"key3": "val3",
		}
		src := map[string]any{
			"key1": "src-val1",
			"key2": "src-val2",
			"key3": "src-val3",
		}

		mergeMaps(dst, src)

		key1 := dst["key1"]
		key2 := dst["key2"]
		key3 := dst["key3"]

		assert.Len(t, dst, 3)
		assert.Equal(t, "val1", key1)
		assert.Equal(t, "val2", key2)
		assert.Equal(t, "val3", key3)
	})

	t.Run("adds new keys from src to dst", func(t *testing.T) {
		dst := map[string]any{
			"key1": "val1",
			"key2": "val2",
			"key3": "val3",
		}
		src := map[string]any{
			"key4": "src-val4",
			"key5": "src-val5",
		}

		mergeMaps(dst, src)

		key1 := dst["key1"]
		key2 := dst["key2"]
		key3 := dst["key3"]
		key4 := dst["key4"]
		key5 := dst["key5"]

		assert.Len(t, dst, 5)
		assert.Equal(t, "val1", key1)
		assert.Equal(t, "val2", key2)
		assert.Equal(t, "val3", key3)
		assert.Equal(t, "src-val4", key4)
		assert.Equal(t, "src-val5", key5)
	})
}
