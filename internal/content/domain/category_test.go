package domain_test

import (
	"testing"

	"github.com/sky-flux/cms/internal/content/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCategory_Valid(t *testing.T) {
	c, err := domain.NewCategory("Technology", "technology", "")
	require.NoError(t, err)
	assert.Equal(t, "Technology", c.Name)
	assert.Equal(t, "technology", c.Slug)
	assert.Equal(t, "", c.ParentID)
}

func TestNewCategory_EmptyName(t *testing.T) {
	_, err := domain.NewCategory("", "slug", "")
	assert.ErrorIs(t, err, domain.ErrEmptyCategoryName)
}

func TestNewCategory_EmptySlug(t *testing.T) {
	_, err := domain.NewCategory("Name", "", "")
	assert.ErrorIs(t, err, domain.ErrEmptyCategorySlug)
}

func TestNewCategory_WithParent(t *testing.T) {
	c, err := domain.NewCategory("React", "react", "parent-id")
	require.NoError(t, err)
	assert.Equal(t, "parent-id", c.ParentID)
	assert.True(t, c.HasParent())
}

func TestCategory_HasParent_WithoutParent(t *testing.T) {
	c, _ := domain.NewCategory("Root", "root", "")
	assert.False(t, c.HasParent())
}

func TestCategory_IsCycleAncestor_DetectsCycle(t *testing.T) {
	// Ancestor path: [grandparent-id, parent-id]
	// Trying to set parent to one of the ancestors should detect cycle.
	ancestors := []string{"grandparent-id", "parent-id"}
	c, _ := domain.NewCategory("Child", "child", "parent-id")
	c.ID = "child-id"

	// Trying to set grandparent as a child of itself → cycle.
	assert.True(t, domain.WouldCreateCycle("grandparent-id", ancestors))
	assert.False(t, domain.WouldCreateCycle("new-parent-id", ancestors))
}
