package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const validBody = `{"title":"Bolo de cenoura","ingredients":["2 cenouras","2 ovos"],"instructions":"Misture e asse por 40 minutos.","prep_minutes":50,"servings":8}`

func request(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}
func TestRecipeLifecycle(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recipes.json")
	s, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	h := handler(s)
	if w := request(h, "GET", "/recipes", ""); w.Code != 200 || strings.TrimSpace(w.Body.String()) != "[]" {
		t.Fatal(w.Code, w.Body.String())
	}
	w := request(h, "POST", "/recipes", validBody)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var r Recipe
	if err = json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatal(err)
	}
	url := "/recipes/" + r.ID
	if r.ID == "" || w.Header().Get("Location") != url {
		t.Fatal("ID ou Location ausente")
	}
	s2, err := NewStore(path)
	if err != nil {
		t.Fatal(err)
	}
	h = handler(s2)
	if w = request(h, "GET", url, ""); w.Code != 200 {
		t.Fatal(w.Code)
	}
	updated := strings.Replace(validBody, "Bolo de cenoura", "Bolo especial", 1)
	if w = request(h, "PUT", url, updated); w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
	if w = request(h, "GET", url, ""); !strings.Contains(w.Body.String(), "Bolo especial") {
		t.Fatal(w.Body.String())
	}
	if w = request(h, "DELETE", url, ""); w.Code != 204 || w.Body.Len() != 0 {
		t.Fatal(w.Code)
	}
	for _, method := range []string{"GET", "PUT", "DELETE"} {
		if w = request(h, method, url, validBody); w.Code != 404 {
			t.Fatalf("%s: %d", method, w.Code)
		}
	}
	s3, err := NewStore(path)
	if err != nil || len(s3.List()) != 0 {
		t.Fatal("exclusão não persistida", err)
	}
}
func TestInvalidInput(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "recipes.json"))
	h := handler(s)
	for _, body := range []string{`{}`, `null`, `{`, validBody + ` {}`, strings.Replace(validBody, `"servings":8`, `"servings":0`, 1), strings.Replace(validBody, `"2 ovos"`, `" "`, 1), strings.Replace(validBody, `"title":`, `"unknown":`, 1), strings.Repeat("x", (1<<20)+1)} {
		if w := request(h, "POST", "/recipes", body); w.Code != 400 {
			t.Fatalf("esperava 400, recebeu %d", w.Code)
		}
	}
	r := httptest.NewRequest("POST", "/recipes", strings.NewReader(validBody))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != 400 {
		t.Fatal(w.Code)
	}
	if len(s.List()) != 0 {
		t.Fatal("entrada inválida foi persistida")
	}
}
func TestConcurrentCreation(t *testing.T) {
	s, _ := NewStore(filepath.Join(t.TempDir(), "recipes.json"))
	h := handler(s)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := request(h, "POST", "/recipes", validBody)
			if w.Code != 201 {
				t.Errorf("status %d", w.Code)
			}
		}()
	}
	wg.Wait()
	if len(s.List()) != 20 {
		t.Fatal("receitas perdidas")
	}
}
func TestPersistenceFailureKeepsState(t *testing.T) {
	s := &Store{path: t.TempDir(), recipes: []Recipe{}}
	if err := s.Create(Recipe{ID: "1"}); err == nil {
		t.Fatal("esperava erro de escrita")
	}
	if len(s.List()) != 0 {
		t.Fatal("estado alterado após falha")
	}
}
func TestCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "recipes.json")
	if err := os.WriteFile(path, []byte("{broken"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewStore(path); err == nil {
		t.Fatal("esperava erro para JSON inválido")
	}
	// Reading a directory must fail instead of silently resetting stored data.
	if _, err := NewStore(filepath.Dir(path)); err == nil {
		t.Fatal("esperava falha de leitura")
	}
}
