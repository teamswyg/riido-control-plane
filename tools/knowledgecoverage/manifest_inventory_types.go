package main

type manifestDir struct {
	Group string `json:"group"`
	Count int    `json:"count"`
}

type manifestSample struct {
	Group string   `json:"group"`
	Paths []string `json:"paths"`
}
