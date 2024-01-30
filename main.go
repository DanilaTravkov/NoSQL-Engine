package main

import (
	"projectDVMVRV/structures"
)

func main () {
	app := App.CreateApp()
	app.RunWebApp()
	app.StopApp()
}