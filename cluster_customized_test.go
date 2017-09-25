package cluster

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

type testClusterPoint struct {
	X,Y float64
	zoom int
	Id int //Index for pint, Id for cluster
	NumPoints int
	aggregatedClusters []testClusterPoint
}

func TestCluster_Customizer (t *testing.T) {
	points := importData("./testdata/places.json")
	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}

	customizer := testCustomizer{}

	c := NewClusterFromCustomizer(customizer)
	c.ClusterPoints(geoPoints)

	northWest := simplePoint{71.36718750000001, -83.79204408779539}
	southEast := simplePoint{-71.01562500000001, 83.7539108491127}

	var result []ClusterPoint = c.GetClusters(northWest, southEast, 2)
	for _, p := range(result) {
		assert.Equal(t, len(p.(testClusterPoint).aggregatedClusters), p.getNumPoints() - 1, "We get all points except itself")
	}
}


func (cp testClusterPoint)	Coordinates() (float64, float64) {
	return cp.X, cp.Y
}

func (cp testClusterPoint) getX() float64 {
	return cp.X
}

func (cp testClusterPoint) getY() float64 {
	return cp.Y
}

func (cp testClusterPoint) setX(x float64) ClusterPoint {
	cp.X = x
	return cp
}

func (cp testClusterPoint) setY(y float64) ClusterPoint {
	cp.Y = y
	return cp
}

func (cp testClusterPoint) setZoom(zoom int) ClusterPoint {
	cp.zoom = zoom
	return cp
}
func (cp testClusterPoint) getZoom() int{
	return cp.zoom
}

func (cp testClusterPoint) setNumPoints(numPoints int) ClusterPoint {
	cp.NumPoints = numPoints
	return cp
}
func (cp testClusterPoint) getNumPoints() int{
	return cp.NumPoints
}
func (cp testClusterPoint) setId(id int) ClusterPoint {
	cp.Id = id
	return cp
}

type testCustomizer struct {

}
func (dc testCustomizer) GeoPoint2ClusterPoint(point GeoPoint) ClusterPoint {
	cp := testClusterPoint{aggregatedClusters: make([]testClusterPoint, 0, 10)}
	return cp
}

func (dc testCustomizer) AggregateClusterPoints(point ClusterPoint, aggregated []ClusterPoint, zoom int) ClusterPoint {
	cp := point.(testClusterPoint)
	if (cap(cp.aggregatedClusters) < len(cp.aggregatedClusters) + len(aggregated)) {
		newAggregated := make([]testClusterPoint, len(cp.aggregatedClusters), max(len(cp.aggregatedClusters) + len(aggregated), 2 * cap(cp.aggregatedClusters)))
		copy(cp.aggregatedClusters, newAggregated)
		cp.aggregatedClusters = newAggregated
	}

	for _, p := range(aggregated) {
		testPoint := p.(testClusterPoint)
		if (cap(cp.aggregatedClusters) < len(cp.aggregatedClusters) + 1 + len(testPoint.aggregatedClusters)) {
			newAggregated := make([]testClusterPoint, len(cp.aggregatedClusters), max(len(cp.aggregatedClusters) + 1 + len(testPoint.aggregatedClusters), 2 * cap(cp.aggregatedClusters)))
			cp.aggregatedClusters = newAggregated
		}
		cp.aggregatedClusters = append(cp.aggregatedClusters, testPoint.aggregatedClusters...)
		cp.aggregatedClusters = append(cp.aggregatedClusters, testPoint)
	}
	return cp
}

func max (a, b int) int {
	if (a > b) {
		return a
	}
	return b
}


func importData(filename string) []*TestPoint {
	var points = struct {
		Type     string
		Features []*TestPoint
	}{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	json.Unmarshal(raw, &points)
	//fmt.Printf("Gett data: %+v\n",points)
	return points.Features
}
////Helpers
type simplePoint struct {
	Lon, Lat float64
}

func (sp simplePoint) GetCoordinates() GeoCoordinates {
	return GeoCoordinates{sp.Lon, sp.Lat}
}
type TestPoint struct {
	Type       string
	Properties struct {
		//we don't need other data
		Name       string
		PointCount int `json:"point_count"`
	}
	Geometry struct {
		Coordinates []float64
	}
}

func (tp *TestPoint) GetCoordinates() GeoCoordinates {
	return GeoCoordinates{
		Lon: tp.Geometry.Coordinates[0],
		Lat: tp.Geometry.Coordinates[1],
	}
}
