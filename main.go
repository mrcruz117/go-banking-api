package main

func main() {
	server := NewAPIServer(":8088")
	server.Run()
}
