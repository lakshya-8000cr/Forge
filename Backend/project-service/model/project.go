package model  // struct for the project , you will notice in microservice , even small to small module will be independent 
// this will get imported innother file  

type Project struct {
	ID        int32
	Name      string
	ImageName string
	Status    string
}