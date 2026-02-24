package model

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"
)

func TestSetTimestamps_Insert(t *testing.T) {
	var createdAt, updatedAt time.Time
	q := (*bun.InsertQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.False(t, createdAt.IsZero(), "createdAt should be set on INSERT")
	assert.False(t, updatedAt.IsZero(), "updatedAt should be set on INSERT")
	assert.Equal(t, createdAt, updatedAt, "both timestamps should match on INSERT")
}

func TestSetTimestamps_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := original
	updatedAt := original
	q := (*bun.UpdateQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.Equal(t, original, createdAt, "createdAt should NOT change on UPDATE")
	assert.NotEqual(t, original, updatedAt, "updatedAt should change on UPDATE")
}

func TestSetTimestamps_Select(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	createdAt := original
	updatedAt := original
	q := (*bun.SelectQuery)(nil)

	SetTimestamps(&createdAt, &updatedAt, q)

	assert.Equal(t, original, createdAt, "createdAt should NOT change on SELECT")
	assert.Equal(t, original, updatedAt, "updatedAt should NOT change on SELECT")
}

func TestSetUpdatedAt_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := original
	q := (*bun.UpdateQuery)(nil)

	SetUpdatedAt(&updatedAt, q)

	assert.NotEqual(t, original, updatedAt, "updatedAt should change on UPDATE")
}

func TestSetUpdatedAt_Insert(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := original
	q := (*bun.InsertQuery)(nil)

	SetUpdatedAt(&updatedAt, q)

	assert.Equal(t, original, updatedAt, "updatedAt should NOT change on INSERT for SetUpdatedAt")
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  Alice@Example.COM  ", "alice@example.com"},
		{"user@test.com", "user@test.com"},
		{" BOB@MAIL.COM", "bob@mail.com"},
		{"", ""},
	}
	for _, tt := range tests {
		email := tt.input
		NormalizeEmail(&email)
		assert.Equal(t, tt.expected, email)
	}
}

func TestUser_BeforeAppendModel_Insert(t *testing.T) {
	u := &User{Email: "  Alice@Example.COM  "}
	q := (*bun.InsertQuery)(nil)

	err := u.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, u.CreatedAt.IsZero())
	assert.False(t, u.UpdatedAt.IsZero())
	assert.Equal(t, "alice@example.com", u.Email)
}

func TestUser_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	u := &User{
		Email:     " BOB@Test.com ",
		CreatedAt: original,
		UpdatedAt: original,
	}
	q := (*bun.UpdateQuery)(nil)

	err := u.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, u.CreatedAt, "createdAt should not change on UPDATE")
	assert.NotEqual(t, original, u.UpdatedAt)
	assert.Equal(t, "bob@test.com", u.Email)
}

func TestPost_BeforeAppendModel_Insert(t *testing.T) {
	p := &Post{Version: 1}
	q := (*bun.InsertQuery)(nil)

	err := p.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, p.CreatedAt.IsZero())
	assert.False(t, p.UpdatedAt.IsZero())
	assert.Equal(t, 1, p.Version, "Version should NOT increment on INSERT")
}

func TestPost_BeforeAppendModel_Update_IncrVersion(t *testing.T) {
	p := &Post{Version: 3}
	q := (*bun.UpdateQuery)(nil)

	err := p.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, 4, p.Version, "Version should increment to 4 on UPDATE")
	assert.False(t, p.UpdatedAt.IsZero())
}

func TestPostTranslation_BeforeAppendModel_Insert(t *testing.T) {
	pt := &PostTranslation{}
	q := (*bun.InsertQuery)(nil)

	err := pt.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.False(t, pt.CreatedAt.IsZero())
	assert.False(t, pt.UpdatedAt.IsZero())
}

func TestComment_BeforeAppendModel_Insert_Anonymous(t *testing.T) {
	c := &Comment{
		AuthorEmail: " Guest@Example.COM ",
		Content:     "hello",
	}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, "guest@example.com", c.AuthorEmail)
	assert.False(t, c.CreatedAt.IsZero())
}

func TestComment_BeforeAppendModel_Insert_Authenticated(t *testing.T) {
	uid := "user-123"
	c := &Comment{
		UserID:      &uid,
		AuthorEmail: " UPPER@test.COM ",
	}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, " UPPER@test.COM ", c.AuthorEmail, "should NOT normalize when user_id is set")
}

func TestComment_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Comment{CreatedAt: original, UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, c.CreatedAt)
	assert.NotEqual(t, original, c.UpdatedAt)
}

func TestConfig_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Config{UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.NotEqual(t, original, c.UpdatedAt)
}

func TestConfig_BeforeAppendModel_Insert(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	c := &Config{UpdatedAt: original}
	q := (*bun.InsertQuery)(nil)

	err := c.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.Equal(t, original, c.UpdatedAt, "SetUpdatedAt should not change on INSERT")
}

func TestSiteConfig_BeforeAppendModel_Update(t *testing.T) {
	original := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	sc := &SiteConfig{UpdatedAt: original}
	q := (*bun.UpdateQuery)(nil)

	err := sc.BeforeAppendModel(nil, q)

	assert.NoError(t, err)
	assert.NotEqual(t, original, sc.UpdatedAt)
}

func TestTimestampOnly_BeforeAppendModel(t *testing.T) {
	type hookable interface {
		BeforeAppendModel(ctx context.Context, query bun.Query) error
	}

	t.Run("Site", func(t *testing.T) {
		m := &Site{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
		assert.False(t, m.UpdatedAt.IsZero())

		orig := m.CreatedAt
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.UpdateQuery)(nil)))
		assert.Equal(t, orig, m.CreatedAt)
	})

	t.Run("MediaFile", func(t *testing.T) {
		m := &MediaFile{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Category", func(t *testing.T) {
		m := &Category{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Role", func(t *testing.T) {
		m := &Role{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("APIEndpoint", func(t *testing.T) {
		m := &APIEndpoint{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("AdminMenu", func(t *testing.T) {
		m := &AdminMenu{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("RoleTemplate", func(t *testing.T) {
		m := &RoleTemplate{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("UserTOTP", func(t *testing.T) {
		m := &UserTOTP{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("SiteMenu", func(t *testing.T) {
		m := &SiteMenu{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("SiteMenuItem", func(t *testing.T) {
		m := &SiteMenuItem{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})

	t.Run("Redirect", func(t *testing.T) {
		m := &Redirect{}
		assert.NoError(t, m.BeforeAppendModel(nil, (*bun.InsertQuery)(nil)))
		assert.False(t, m.CreatedAt.IsZero())
	})
}
