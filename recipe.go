package main

import (
	"errors"
	"strings"
)

type Recipe struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Ingredients  []string `json:"ingredients"`
	Instructions string   `json:"instructions"`
	PrepMinutes  int      `json:"prep_minutes"`
	Servings     int      `json:"servings"`
}

type RecipeInput struct {
	Title        string   `json:"title"`
	Ingredients  []string `json:"ingredients"`
	Instructions string   `json:"instructions"`
	PrepMinutes  int      `json:"prep_minutes"`
	Servings     int      `json:"servings"`
}

func (in RecipeInput) recipe(id string) (Recipe, error) {
	r := Recipe{ID: id, Title: strings.TrimSpace(in.Title), Instructions: strings.TrimSpace(in.Instructions), PrepMinutes: in.PrepMinutes, Servings: in.Servings, Ingredients: make([]string, len(in.Ingredients))}
	if r.Title == "" || len(r.Title) > 200 {
		return r, errors.New("title deve ter entre 1 e 200 bytes")
	}
	if r.Instructions == "" {
		return r, errors.New("instructions é obrigatório")
	}
	if r.PrepMinutes <= 0 || r.Servings <= 0 {
		return r, errors.New("prep_minutes e servings devem ser positivos")
	}
	if len(in.Ingredients) == 0 {
		return r, errors.New("informe pelo menos um ingrediente")
	}
	for i, s := range in.Ingredients {
		r.Ingredients[i] = strings.TrimSpace(s)
		if r.Ingredients[i] == "" {
			return r, errors.New("ingredientes não podem estar vazios")
		}
	}
	return r, nil
}
