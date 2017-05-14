package wrouter

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

type tMockDependency struct {
	works string
}

type tMockController struct {
}

func (m *tMockController) ArgsrouteAction(ctx *Context) string {
	fmt.Println("argsrouteactioncalled")
	return "i am the return string"

}

func (m *tMockController) InjectAction(ctx *Context, dep *tMockDependency) string {
	fmt.Println("controller dependency says", dep.works)
	return "i am the injection return string"
}

type tMockInjector struct {
}

func (m *tMockInjector) Supports(t string) bool {
	return t == "*wrouter.tMockDependency"
}

func (m *tMockInjector) Get(ctx *InjectorContext) interface{} {
	return &tMockDependency{works: "yes indeed"}
}

type tMockDependencyPost struct {
	works string
}
type tMockInjectorPost struct {
}

func (m *tMockInjectorPost) Supports(t string) bool {
	return t == "*wrouter.tMockDependencyPost"
}

func (m *tMockInjectorPost) Get(ctx *InjectorContext) interface{} {
	return &tMockDependencyPost{works: "post event"}
}

type tTestEventPre struct {
}

func (e *tTestEventPre) Exec(ctx *Context, d *tMockDependency) {
	fmt.Println(d.works)
}

type tTestPost struct {
}

func (e *tTestPost) Exec(ctx *Context, d *tMockDependencyPost) {
	fmt.Println(d.works, ctx.results[0].Interface().(string))
}

func TestRouterT(t *testing.T) {

	router := NewRouter()
	if router == nil {
		t.Error("NewRouter returned nil")
	}

	router.AddController(new(tMockController))
	router.AddInjector(new(tMockInjector))
	router.AddInjector(new(tMockInjectorPost))
	router.AddPreRequestEvent(new(tTestEventPre))
	router.AddPostRequestEvent(new(tTestPost))

	go http.ListenAndServe(":1337", router)

	for _, v := range router.routes {
		t.Log(v.Path)
	}

	expectedRoutes := []string{"tmock/argsroute", "tmock/inject"}

	time.Sleep(time.Second)

	for _, req := range expectedRoutes {
		_, err := http.Get("http://127.0.0.1:1337/" + req)
		if err != nil {
			t.Error(err.Error())
		}
	}

}
