package wrouter

import (
	"testing"
	"strconv"
)

var tMockCtrA string = "notDone"
var tMockCtrB string = "notDone"

type tMockController struct{ _ *tSubController }
type tSubController struct{}

func (t *tMockController) IndexAction()      {}
func (t *tMockController) Post_IndexAction() {}
func (t *tMockController) AnotherAction()    {}
func (t *tSubController) SubAction()         {}
func (t *tSubController) Delete_UserAction() {}

type tMockDependency struct{ c int }
type tMockDependencyInjector struct{}

func (t *tMockDependencyInjector) Supports(te string) bool { return te == "*wrouter.tMockDependency" }
func (t *tMockDependencyInjector) Get(ctx *InjectorContext) interface{} {
	return &tMockDependency{c: 15}
}

type tMockPreRequestEvent struct{}
type tMockPostRequestEvent struct{}

func (t *tMockPreRequestEvent) Exec(ctx *PreRequestEventContext)   { tMockCtrA = "done" }
func (t *tMockPostRequestEvent) Exec(ctx *PostRequestEventContext) { tMockCtrB = "done" }

func TestRouter(t *testing.T) {
	rt := NewRouter()

	rt.AddController(&tMockController{})

	if len(rt.routes) != 7 {
		t.Error("Error counting generated routes. Exptected 7, got "+strconv.Itoa(len(rt.routes))+". Please" +
			"Consider checking the generation of aliasses for the index-Actions")
	}

	rt.AppendPreRequestEvent(new(tMockPreRequestEvent))
	rt.AppendPostRequestEvent(new(tMockPostRequestEvent))


}