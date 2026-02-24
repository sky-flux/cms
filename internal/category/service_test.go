package category_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sky-flux/cms/internal/category"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Mock: CategoryRepository
// ---------------------------------------------------------------------------

type mockCategoryRepo struct {
	// List
	listCats []model.Category
	listErr  error

	// GetByID
	getByIDMap map[string]*model.Category
	getByIDErr error

	// GetChildren
	childrenMap map[string][]model.Category
	childrenErr error

	// Create
	createErr error

	// Update
	updateErr error

	// Delete
	deleteErr error

	// SlugExistsUnderParent
	slugExists    bool
	slugExistsErr error

	// UpdatePathPrefix
	pathPrefixRows int64
	pathPrefixErr  error

	// BatchUpdateSortOrder
	batchErr error

	// CountPosts
	postCountMap map[string]int64
	postCountErr error
}

func newMockRepo() *mockCategoryRepo {
	return &mockCategoryRepo{
		getByIDMap:   make(map[string]*model.Category),
		childrenMap:  make(map[string][]model.Category),
		postCountMap: make(map[string]int64),
	}
}

func (m *mockCategoryRepo) List(_ context.Context) ([]model.Category, error) {
	return m.listCats, m.listErr
}

func (m *mockCategoryRepo) GetByID(_ context.Context, id string) (*model.Category, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	cat, ok := m.getByIDMap[id]
	if !ok {
		return nil, apperror.NotFound("category not found", nil)
	}
	return cat, nil
}

func (m *mockCategoryRepo) GetChildren(_ context.Context, parentID string) ([]model.Category, error) {
	return m.childrenMap[parentID], m.childrenErr
}

func (m *mockCategoryRepo) Create(_ context.Context, cat *model.Category) error {
	if m.createErr == nil {
		cat.ID = "new-cat-id"
		cat.CreatedAt = time.Now()
		cat.UpdatedAt = time.Now()
	}
	return m.createErr
}

func (m *mockCategoryRepo) Update(_ context.Context, _ *model.Category) error {
	return m.updateErr
}

func (m *mockCategoryRepo) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockCategoryRepo) SlugExistsUnderParent(_ context.Context, _ string, _ *string, _ string) (bool, error) {
	return m.slugExists, m.slugExistsErr
}

func (m *mockCategoryRepo) UpdatePathPrefix(_ context.Context, _, _ string) (int64, error) {
	return m.pathPrefixRows, m.pathPrefixErr
}

func (m *mockCategoryRepo) BatchUpdateSortOrder(_ context.Context, _ []category.SortOrderItem) error {
	return m.batchErr
}

func (m *mockCategoryRepo) CountPosts(_ context.Context, categoryID string) (int64, error) {
	if m.postCountErr != nil {
		return 0, m.postCountErr
	}
	return m.postCountMap[categoryID], nil
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func newTestService(repo *mockCategoryRepo) *category.Service {
	c := cache.NewClient(nil)
	al := audit.NewNoopLogger()
	return category.NewService(repo, c, al)
}

// ---------------------------------------------------------------------------
// Tests: CreateCategory
// ---------------------------------------------------------------------------

func TestCreateCategory_Success(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	resp, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name: "Tech",
		Slug: "tech",
	})
	require.NoError(t, err)
	assert.Equal(t, "new-cat-id", resp.ID)
	assert.Equal(t, "Tech", resp.Name)
	assert.Equal(t, "/tech/", resp.Path)
	assert.Nil(t, resp.ParentID)
}

func TestCreateCategory_DuplicateSlug(t *testing.T) {
	repo := newMockRepo()
	repo.slugExists = true
	svc := newTestService(repo)

	_, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name: "Tech",
		Slug: "tech",
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestCreateCategory_WithParent(t *testing.T) {
	repo := newMockRepo()
	parentID := "parent-id"
	repo.getByIDMap["parent-id"] = &model.Category{
		ID:   "parent-id",
		Name: "Parent",
		Slug: "parent",
		Path: "/parent/",
	}
	svc := newTestService(repo)

	resp, err := svc.CreateCategory(context.Background(), &category.CreateCategoryReq{
		Name:     "Child",
		Slug:     "child",
		ParentID: &parentID,
	})
	require.NoError(t, err)
	assert.Equal(t, "/parent/child/", resp.Path)
	assert.Equal(t, &parentID, resp.ParentID)
}

// ---------------------------------------------------------------------------
// Tests: DeleteCategory
// ---------------------------------------------------------------------------

func TestDeleteCategory_HasChildren(t *testing.T) {
	repo := newMockRepo()
	repo.getByIDMap["cat-1"] = &model.Category{ID: "cat-1", Name: "Parent", Slug: "parent", Path: "/parent/"}
	repo.childrenMap["cat-1"] = []model.Category{
		{ID: "cat-2", Name: "Child", Slug: "child", Path: "/parent/child/"},
	}
	svc := newTestService(repo)

	err := svc.DeleteCategory(context.Background(), "cat-1")
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrConflict))
}

func TestDeleteCategory_Leaf(t *testing.T) {
	repo := newMockRepo()
	repo.getByIDMap["cat-1"] = &model.Category{ID: "cat-1", Name: "Leaf", Slug: "leaf", Path: "/leaf/"}
	// No children for cat-1.
	svc := newTestService(repo)

	err := svc.DeleteCategory(context.Background(), "cat-1")
	require.NoError(t, err)
}

// ---------------------------------------------------------------------------
// Tests: BuildTree (via ListTree)
// ---------------------------------------------------------------------------

func TestBuildTree(t *testing.T) {
	repo := newMockRepo()
	parentID := "parent-id"
	repo.listCats = []model.Category{
		{ID: "parent-id", Name: "Parent", Slug: "parent", Path: "/parent/", CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{ID: "child-id", Name: "Child", Slug: "child", Path: "/parent/child/", ParentID: &parentID, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	repo.postCountMap["parent-id"] = 5
	repo.postCountMap["child-id"] = 3
	svc := newTestService(repo)

	tree, err := svc.ListTree(context.Background())
	require.NoError(t, err)
	require.Len(t, tree, 1, "should have 1 root")

	root := tree[0]
	assert.Equal(t, "parent-id", root.ID)
	assert.Equal(t, int64(5), root.PostCount)
	require.Len(t, root.Children, 1, "root should have 1 child")

	child := root.Children[0]
	assert.Equal(t, "child-id", child.ID)
	assert.Equal(t, int64(3), child.PostCount)
}

// ---------------------------------------------------------------------------
// Tests: UpdateCategory
// ---------------------------------------------------------------------------

func TestUpdateCategory_Success(t *testing.T) {
	repo := newMockRepo()
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "Old", Slug: "old", Path: "/old/",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	svc := newTestService(repo)

	newName := "New Name"
	resp, err := svc.UpdateCategory(context.Background(), "cat-1", &category.UpdateCategoryReq{
		Name: &newName,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", resp.Name)
}

func TestUpdateCategory_CycleDetection(t *testing.T) {
	repo := newMockRepo()
	// cat-1 is parent of cat-2
	parentID := "cat-1"
	repo.getByIDMap["cat-1"] = &model.Category{
		ID: "cat-1", Name: "A", Slug: "a", Path: "/a/",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.getByIDMap["cat-2"] = &model.Category{
		ID: "cat-2", Name: "B", Slug: "b", Path: "/a/b/", ParentID: &parentID,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	svc := newTestService(repo)

	// Try to make cat-1's parent be cat-2 (creates cycle: cat-1 -> cat-2 -> cat-1).
	newParent := "cat-2"
	_, err := svc.UpdateCategory(context.Background(), "cat-1", &category.UpdateCategoryReq{
		ParentID: &newParent,
	})
	require.Error(t, err)
	assert.True(t, errors.Is(err, apperror.ErrValidation))
}

// ---------------------------------------------------------------------------
// Tests: Reorder
// ---------------------------------------------------------------------------

func TestReorder_Success(t *testing.T) {
	repo := newMockRepo()
	svc := newTestService(repo)

	err := svc.Reorder(context.Background(), []category.SortOrderItem{
		{ID: "cat-1", SortOrder: 2},
		{ID: "cat-2", SortOrder: 1},
	})
	require.NoError(t, err)
}
