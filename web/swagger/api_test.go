package swagger_test

import (
	"net/http"
	"testing"

	"github.com/altairsix/pkg/web"
	"github.com/altairsix/pkg/web/swagger"
)

func TestNew(t *testing.T) {
	api := swagger.New()
	r := web.NewRouter().WithObserver(api)
	r.GET("/a/b", nil)
}

func SkipServer(t *testing.T) {
	api := swagger.New()
	r := web.NewRouter().WithObserver(api)

	fn := func(c web.Context) error {
		return c.Text(http.StatusOK, "ok")
	}

	r.GET("/:sample", fn,
		swagger.Summary("this is the summary"),
		swagger.Description("and the description"),
		swagger.Query("foo", "does something", true),
	)

	http.ListenAndServe(":9090", api.Swagger.Handler(true))
}
