package main

import (
	"projectDVMVRV/structures"
)

func main() {
	app := structures.CreateApp()
	app.RunApp()
	app.StopApp()
}
