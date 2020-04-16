package main

type RedirectHeader struct {
	Name  string
	Value string
}

type RedirectResponse struct {
	Uri      string
	Status   int64
	Headers  []RedirectHeader
	Body     string
	Upstream string
}