package main

func main() {
	grpcManager := NewGRPCManager()
	api := NewAPI(grpcManager)
	defer api.Close()
}
