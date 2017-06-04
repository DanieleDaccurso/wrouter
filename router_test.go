package wrouter

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"math/rand"
)

var tMockCtrA string = "notDone"
var tMockCtrB string = "notDone"

type tMockController struct{ _ *tSubController }
type tSubController struct{}

func (t *tMockController) IndexAction(r *http.Request)                          {}
func (t *tMockController) Post_IndexAction(w http.ResponseWriter)               {}
func (t *tMockController) AnotherAction(r *http.Request, w http.ResponseWriter) {}
func (t *tSubController) SubAction(w http.ResponseWriter, r *http.Request)      {}
func (t *tSubController) Delete_UserAction(user *tMockUser)                     { user.deleted = true }

type tMockUserInjector struct{}
type tMockUser struct{ deleted bool }

var tUser *tMockUser = &tMockUser{false}

func (t *tMockUserInjector) Supports(ty string) bool              { return ty == "*wrouter.tMockUser" }
func (t *tMockUserInjector) Get(ctx *InjectorContext) interface{} { return tUser }

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
		t.Error("Error counting generated routes. Exptected 7, got " + strconv.Itoa(len(rt.routes)))
	}

	rt.AppendPreRequestEvent(new(tMockPreRequestEvent))
	rt.AppendPostRequestEvent(new(tMockPostRequestEvent))

	rt.AddInjector(new(tMockUserInjector))

	server := httptest.NewServer(rt)
	defer server.Close()
	t.Log("Test server started: " + server.URL)

	requests := make([]*http.Request, 7)
	requests[0], _ = http.NewRequest("GET", server.URL+"/tmock/index", nil)
	requests[1], _ = http.NewRequest("POST", server.URL+"/tmock/index", nil)
	requests[2], _ = http.NewRequest("GET", server.URL+"/tmock/", nil)
	requests[3], _ = http.NewRequest("POST", server.URL+"/tmock/", nil)
	requests[4], _ = http.NewRequest("GET", server.URL+"/tmock/another", nil)
	requests[5], _ = http.NewRequest("GET", server.URL+"/tmock/tsub/sub", nil)
	requests[6], _ = http.NewRequest("DELETE", server.URL+"/tmock/tsub/user", nil)

	for _, request := range requests {
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Error("Error in request: " + request.URL.String() + "; ERR: " + err.Error())
		}
		if response.Status != "200 OK" {
			t.Error("Error in request: " + request.URL.String() + "; ERR: Status=" + response.Status)
		}
	}

	invalidRequests := make([]*http.Request, 4)
	invalidRequests[0], _ = http.NewRequest("PUT", server.URL+"/tmock/index", nil)
	invalidRequests[1], _ = http.NewRequest("DELETE", server.URL+"/tmock/index", nil)
	invalidRequests[2], _ = http.NewRequest("HEAD", server.URL+"/tmock/", nil)
	invalidRequests[3], _ = http.NewRequest("POST", server.URL+"/doesntexist", nil)

	for _, request := range invalidRequests {
		response, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Error("Error in request: " + request.URL.String() + "; ERR: " + err.Error())
		}
		if response.Status == "200 OK" {
			t.Error("Error in request: " + request.URL.String() + "; GOT 200 OK on faulty route")
		}
	}

	// -- Minimal event tests. More tests in event package
	if tMockCtrA != "done" {
		t.Error("PRE REQUEST EXEC FAIL")
	}

	if tMockCtrB != "done" {
		t.Error("POST REQUEST EXEC FAIL")
	}

	// injector test
	if !tUser.deleted {
		t.Error("INJECTOR EXEC FAIL")
	}
}

func BenchmarkRqResolver(b *testing.B) {
	rt := NewRouter()
	rt.AddController(&tMockController{})
	rt.AddInjector(new(tMockUserInjector))

	server := httptest.NewServer(rt)
	defer server.Close()
	b.Log("Test server started: " + server.URL)

	requests := make([]*http.Request, 7)
	requests[0], _ = http.NewRequest("GET", server.URL+"/tmock/index", nil)
	requests[1], _ = http.NewRequest("POST", server.URL+"/tmock/index", nil)
	requests[2], _ = http.NewRequest("GET", server.URL+"/tmock/", nil)
	requests[3], _ = http.NewRequest("POST", server.URL+"/tmock/", nil)
	requests[4], _ = http.NewRequest("GET", server.URL+"/tmock/another", nil)
	requests[5], _ = http.NewRequest("GET", server.URL+"/tmock/tsub/sub", nil)
	requests[6], _ = http.NewRequest("DELETE", server.URL+"/tmock/tsub/user", nil)
	b.StartTimer()
	for i := 0; i <= 1000000; i ++ {
		http.DefaultClient.Do(requests[rand.Intn(len(requests))])
	}
	b.StopTimer()

}