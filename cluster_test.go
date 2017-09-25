package cluster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"encoding/json"
)



func TestNewCluster(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := NewCluster()
	assert.Equal(t, c.MinZoom, 0, "they shoud be equal")
	assert.Equal(t, c.MaxZoom, 16, "they shoud be equal")
	assert.Equal(t, c.PointSize, 40, "they shoud be equal")
	assert.Equal(t, c.TileSize, 512, "they shoud be equal")
	assert.Equal(t, c.NodeSize, 64, "they shoud be equal")
}

func TestCluster_GetTile00(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := NewCluster()
	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}
	c.PointSize = 60
	c.MaxZoom = 3
	c.TileSize = 256
	c.NodeSize = 64
	c.ClusterPoints(geoPoints)
	result := c.GetTile(0, 0, 0)
	assert.NotEmpty(t, result)

	expectedPoints := importPoints("./testdata/expect_tile0_0_0.json")
	assert.Equal(t, expectedPoints, result)
}

//validate original result from JS library
func TestCluster_GetTileDefault(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := NewCluster()
	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}
	c.ClusterPoints(geoPoints)
	result := c.GetTile(0, 0, 0)
	assert.NotEmpty(t, result)

	expectedPoints := importGeoJSONResultFeature("./testdata/places-z0-0-0.json")
	assert.Equal(t, len(result), len(expectedPoints))
	for i := range result {
		rp := result[i]
		ep := expectedPoints[i]
		assert.Equal(t, rp.getX(), ep.Geometry[0][0])
		assert.Equal(t, rp.getY(), ep.Geometry[0][1])
		if rp.getNumPoints() > 1 {
			assert.Equal(t, rp.getNumPoints(), ep.Tags.PointCount)
		}

	}
}

func TestCluster_GetClusters(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := NewCluster()
	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}
	c.PointSize = 40
	c.MaxZoom = 17
	c.TileSize = 512
	c.NodeSize = 64
	c.ClusterPoints(geoPoints)

	northWest := simplePoint{71.36718750000001, -83.79204408779539}
	southEast := simplePoint{-71.01562500000001, 83.7539108491127}

	var result []ClusterPoint = c.GetClusters(northWest, southEast, 2)
	assert.NotEmpty(t, result)

	expectedPoints := importData("./testdata/GetCluster.json")
	assert.Equal(t, len(result), len(expectedPoints))
	for i := range result {
		rp := result[i]
		ep := expectedPoints[i]

		t.Logf("Zoom", rp.getZoom())
		t.Logf("Coordinates are:", rp.getX(), ep.Geometry.Coordinates[0])
		assert.True(t, floatEquals(rp.getX(), ep.Geometry.Coordinates[0]))
		assert.True(t, floatEquals(rp.getY(), ep.Geometry.Coordinates[1]))
		if rp.getNumPoints() > 1 {
			assert.Equal(t, rp.getNumPoints(), ep.Properties.PointCount)
		}
	}

	//resultJSON, _ :=  json.MarshalIndent(result,"","    ")
	//fmt.Printf("getting points %v \n",string(resultJSON))
}


func TestCluster_AllClusters(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := NewCluster()
	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}
	c.PointSize = 40
	c.MaxZoom = 17
	c.TileSize = 512
	c.NodeSize = 64
	c.ClusterPoints(geoPoints)


	var result []ClusterPoint = c.AllClusters(2)
	assert.NotEmpty(t, result)

	//resultJSON, _ :=  json.MarshalIndent(result,"","    ")
	//fmt.Printf("getting points %v \n",string(resultJSON))

	assert.Equal(t, 100, len(result))

}
func Test_MercatorProjection(t *testing.T) {
	coor := GeoCoordinates{
		Lon: -79.04411780507252, //0.2804330060970208
		Lat: 43.08771393436908,  //0.36711590445377973
	}
	x, y := MercatorProjection(coor)
	assert.Equal(t, x, 0.2804330060970208)
	assert.Equal(t, y, 0.36711590445377973)

	coor = GeoCoordinates{
		Lon: -62.06181800038502,
		Lat: 5.686896063275327,
	}
	x, y = MercatorProjection(coor)
	assert.Equal(t, x, 0.32760606111004165)
	assert.Equal(t, y, 0.4841770650015434)
}

func ExampleCluster_GetTile() {
	points := importData("./testdata/places.json")

	c := NewCluster()
	c.PointSize = 60
	c.MaxZoom = 3
	c.TileSize = 256
	c.NodeSize = 64

	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}

	c.ClusterPoints(geoPoints)
	result := c.GetTile(0, 0, 4)
	fmt.Printf("%+v",result)
	// Output: [{X:-2418 Y:165 zoom:0 Id:62 NumPoints:1} {X:-3350 Y:253 zoom:0 Id:22 NumPoints:1}]

}

func ExampleCluster_GetClusters() {
	points := importData("./testdata/places.json")

	geoPoints := make([]GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}

	c := NewCluster()
	c.ClusterPoints(geoPoints)

	northWest := simplePoint{71.36718750000001, -83.79204408779539}
	southEast := simplePoint{-71.01562500000001, 83.7539108491127}

	var result []ClusterPoint = c.GetClusters(northWest, southEast, 2)
	fmt.Printf("%+v",result[:3])
	// Output: [{X:-14.473194953510028 Y:26.157965399212813 zoom:2 Id:107 NumPoints:1} {X:-12.408741828510014 Y:58.16339752811905 zoom:2 Id:159 NumPoints:1} {X:-9.269962828651519 Y:42.928736057812586 zoom:2 Id:127 NumPoints:1}]
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

type GeoJSONResultFeature struct {
	Geometry [][]float64
	Tags     struct {
		PointCount int `json:"point_count"`
	}
}

func (tp *TestPoint) GetCoordinates() GeoCoordinates {
	return GeoCoordinates{
		Lon: tp.Geometry.Coordinates[0],
		Lat: tp.Geometry.Coordinates[1],
	}
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

func importPoints(filename string) []ClusterPoint {
	var loaded []clusterPoint
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	json.Unmarshal(raw, &loaded)
	result := make([]ClusterPoint, len(loaded))
	for i, d := range loaded {
		result[i] = d
	}
	return result
}

func importGeoJSONResultFeature(filename string) []GeoJSONResultFeature {
	var points = struct {
		Features []GeoJSONResultFeature
	}{}
	raw, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	json.Unmarshal(raw, &points)
	return points.Features
}

var EPSILON float64 = 0.0000000001

func floatEquals(a, b float64) bool {
	if (a-b) < EPSILON && (b-a) < EPSILON {
		return true
	}
	return false
}
