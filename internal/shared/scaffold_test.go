package shared_test

import (
	"testing"
	_ "github.com/sky-flux/cms/internal/identity/domain"
	_ "github.com/sky-flux/cms/internal/content/domain"
	_ "github.com/sky-flux/cms/internal/media/domain"
	_ "github.com/sky-flux/cms/internal/site/domain"
	_ "github.com/sky-flux/cms/internal/platform/domain"
	_ "github.com/sky-flux/cms/internal/shared"
)

func TestDDDScaffoldExists(t *testing.T) {
	t.Log("all DDD bounded context packages are importable")
}
