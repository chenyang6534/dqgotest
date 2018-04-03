package model

type BaseModel interface {
	Run(closeSig chan bool)
	
}