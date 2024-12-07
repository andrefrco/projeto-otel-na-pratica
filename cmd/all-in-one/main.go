package main

import (
	"net/http"

	"github.com/dosedetelemetria/projeto-otel-na-pratica/internal/app"
)

func main() {
	mux := http.NewServeMux()

	{
		a := app.NewUser()
		a.RegisterRoutes(mux)
	}

	{
		a := app.NewPlan()
		a.RegisterRoutes(mux)
	}

	{
		a, err := app.NewPayment()
		if err != nil {
			panic(err)
		}
		a.RegisterRoutes(mux)
	}

	{
		a := app.NewSubscription()
		a.RegisterRoutes(mux)
	}

	http.ListenAndServe(":8080", mux)
}