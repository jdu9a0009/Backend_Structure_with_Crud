package service

import (
	"fmt"
	"math"
)

type Location struct {
	Lat float32
	Lon float32
}

const EarthRadius = 6371000

func CalculateDistance(loc1, loc2 Location) float64 {

	lat1 := loc1.Lat * math.Pi / 180.0
	lon1 := loc1.Lon * math.Pi / 180.0
	lat2 := loc2.Lat * math.Pi / 180.0
	lon2 := loc2.Lon * math.Pi / 180.0

	dLon := lon2 - lon1
	dLat := lat2 - lat1

	a := math.Pow(math.Sin(float64(dLat/2)), 2) + math.Cos(float64(lat1))*math.Cos(float64(lat2))*math.Pow(math.Sin(float64(dLon/2)), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Masofa (kilometrlar)
	distance := EarthRadius * c
	return distance
}

func CheckDistance(loc1, loc2 Location, distance int) bool {
	d := CalculateDistance(loc1, loc2)
	fmt.Println(d)
	res := float64(distance) >= d
	return res
}
