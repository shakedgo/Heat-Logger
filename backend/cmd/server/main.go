package main

import router "heat-logger/internal/routes"

func main() {
	r := router.SetupRouter()

	r.Run(":8080")
}
