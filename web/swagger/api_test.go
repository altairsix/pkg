package swagger_test

import (
	"testing"

	"encoding/json"
	"os"

	"github.com/altairsix/pkg/web"
	"github.com/altairsix/pkg/web/swagger"
)

func TestNew(t *testing.T) {
	api := swagger.New()
	r := web.NewRouter().WithObserver(api)
	r.GET("/a/b", nil)

	json.NewEncoder(os.Stdout).Encode(api)
}
