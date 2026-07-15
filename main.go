package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func respond(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if value != nil {
		_ = json.NewEncoder(w).Encode(value)
	}
}
func failure(w http.ResponseWriter, status int, message string) {
	respond(w, status, map[string]string{"error": message})
}
func decode(w http.ResponseWriter, r *http.Request) (RecipeInput, error) {
	var in RecipeInput
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		return in, errors.New("Content-Type deve ser application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if err := d.Decode(&in); err != nil {
		return in, errors.New("JSON inválido, campos desconhecidos ou corpo acima de 1 MiB")
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return in, errors.New("envie apenas um objeto JSON")
	}
	return in, nil
}
func handler(s *Store) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, map[string]string{"status": "ok"}) })
	mux.HandleFunc("GET /recipes", func(w http.ResponseWriter, r *http.Request) { respond(w, 200, s.List()) })
	mux.HandleFunc("GET /recipes/{id}", func(w http.ResponseWriter, r *http.Request) {
		for _, recipe := range s.List() {
			if recipe.ID == r.PathValue("id") {
				respond(w, 200, recipe)
				return
			}
		}
		failure(w, 404, errNotFound.Error())
	})
	save := func(w http.ResponseWriter, r *http.Request, create bool) {
		in, err := decode(w, r)
		if err != nil {
			failure(w, 400, err.Error())
			return
		}
		id := r.PathValue("id")
		if create {
			var b [16]byte
			if _, err = rand.Read(b[:]); err != nil {
				failure(w, 500, "erro ao gerar ID")
				return
			}
			id = hex.EncodeToString(b[:])
		}
		recipe, err := in.recipe(id)
		if err != nil {
			failure(w, 400, err.Error())
			return
		}
		if create {
			err = s.Create(recipe)
		} else {
			err = s.Update(id, recipe)
		}
		if errors.Is(err, errNotFound) {
			failure(w, 404, err.Error())
			return
		}
		if err != nil {
			log.Printf("persistência: %v", err)
			failure(w, 500, "erro ao salvar receita")
			return
		}
		status := 200
		if create {
			status = 201
			w.Header().Set("Location", "/recipes/"+id)
		}
		respond(w, status, recipe)
	}
	mux.HandleFunc("POST /recipes", func(w http.ResponseWriter, r *http.Request) { save(w, r, true) })
	mux.HandleFunc("PUT /recipes/{id}", func(w http.ResponseWriter, r *http.Request) { save(w, r, false) })
	mux.HandleFunc("DELETE /recipes/{id}", func(w http.ResponseWriter, r *http.Request) {
		err := s.Delete(r.PathValue("id"))
		if errors.Is(err, errNotFound) {
			failure(w, 404, err.Error())
			return
		}
		if err != nil {
			failure(w, 500, "erro ao excluir receita")
			return
		}
		w.WriteHeader(204)
	})
	return mux
}
func main() {
	path := os.Getenv("DATA_FILE")
	if path == "" {
		path = "data/recipes.json"
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	s, err := NewStore(path)
	if err != nil {
		log.Fatal(err)
	}
	server := &http.Server{Addr: addr, Handler: handler(s), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("API de receitas em %s", addr)
	log.Fatal(server.ListenAndServe())
}
