package wrouter

import (
	"bytes"
	"net/http"
	"testing"
	"time"
)

type mockDependency struct {
	works string
}

type subController struct {

}

func (s *subController) SubAction() {

}

type mockController struct {
	_ *subController
}

func (m *mockController) RouteAction() {

}

func (m *mockController) ArgsrouteAction(h *http.Request, w http.ResponseWriter) {

}

func (m *mockController) FlippedargsAction(w http.ResponseWriter, h *http.Request) {

}

func (m *mockController) HasreturnAction() string {
	return "hello :)"
}

func (m *mockController) InjectAction(dep *mockDependency) {

}

func (m *mockController) POST_postAction() {

}

type mockInjector struct {
}

func (m *mockInjector) Supports(t string) bool {
	return t == "*wrouter.mockDependency"
}

func (m *mockInjector) Get(ctx *InjectorContext) interface{} {
	return &mockDependency{}
}

func TestRouter(t *testing.T) {
	router := NewRouter()
	if router == nil {
		t.Error("NewRouter returned nil")
	}

	router.AddController(new(mockController))
	router.AddInjector(new(mockInjector))

	go http.ListenAndServe(":1337", router)

	expectedRoutes := []string{"mock/argsroute", "mock/flippedargs", "mock/hasreturn", "mock/inject",
		"mock/route", "mock/post"}

	for _, v := range router.routes  {
		t.Log(v.Path)
	}

	time.Sleep(time.Second)

	for _, req := range expectedRoutes {
		resp, err := http.Get("http://127.0.0.1:1337/" + req)
		if err != nil {
			t.Error(err.Error())
		}
		if req != "mock/post" {
			if resp.Status != "200 OK" {
				t.Error("Route failed to return 200: " + req + ", returned: " + resp.Status)
			}
		} else {
			if resp.Status != "404 Not Found" {
				t.Error("POST route should have been 404 on GET request")
			}
		}

		if req == "mock/post" {
			var b bytes.Buffer
			resp, err := http.Post("http://127.0.0.1:1337/mock/post", "text/html", &b)
			if err != nil {
				t.Error("fail for POST request: " + err.Error())
			}
			if resp.Status != "200 OK" {
				t.Error("POST request is " + resp.Status)
			}
		}
	}
}
