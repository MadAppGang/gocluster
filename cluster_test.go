package cluster_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/MadAppGang/gocluster"
)

type TestPoint struct {
	Type       string
	Properties struct {
		//we don't need other data
		Name string
	}
	Geometry struct {
		Coordinates []float64
	}
}

func (tp *TestPoint)GetCoordinates() cluster.GeoCoordinates {
	return cluster.GeoCoordinates {
		Lon: tp.Geometry.Coordinates[0],
		Lat: tp.Geometry.Coordinates[1],
	}
}

func (tp *TestPoint)GeoPointID() string {
	return tp.Properties.Name
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

func TestNewCluster(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := cluster.NewCluster()
	assert.Equal(t, c.MinZoom, 0, "they shoud be equal")
	assert.Equal(t, c.MaxZoom, 18, "they shoud be equal")
	assert.Equal(t, c.PointSize, 50, "they shoud be equal")
	assert.Equal(t, c.TileSize, 512, "they shoud be equal")
	assert.Equal(t, c.MinBranch, 32, "they shoud be equal")
	assert.Equal(t, c.MaxBranch, 64, "they shoud be equal")
}

func TestCluster_GetTile00(t *testing.T) {
	points := importData("./testdata/places.json")
	if len(points) == 0 {
		t.Error("Getting empty test data")
	} else {
		t.Logf("Getting %v points to test\n", len(points))
	}

	c := cluster.NewCluster()
	geoPoints := make ([]cluster.GeoPoint, len(points))
	for i := range points {
		geoPoints[i] = points[i]
	}
	c.PointSize = 60
	c.MaxZoom = 1
	c.TileSize = 256
	c.ClusterPointsWithDeepLinking(geoPoints, false)
	result := c.GetTile(0,0,0)
	assert.NotEmpty(t,result)
	fmt.Printf("Getting points: %+v\n length %v \n",result, len(result))

	resultJSON, _ := json.MarshalIndent(result,  "", "  ")
	fmt.Println(string(resultJSON))
}


func Test_MercatorProjection(t *testing.T) {
	coor := cluster.GeoCoordinates{
		Lon:-79.04411780507252, //0.2804330060970208
		Lat:43.08771393436908, //0.36711590445377973
	}
	x,y := cluster.MercatorProjection(coor)
	assert.Equal(t,x,0.2804330060970208)
	assert.Equal(t,y,0.36711590445377973)

	coor = cluster.GeoCoordinates{
		Lon:-62.06181800038502,
		Lat:5.686896063275327,
	}
	x,y = cluster.MercatorProjection(coor)
	assert.Equal(t,x,0.32760606111004165)
	assert.Equal(t,y,0.4841770650015434)
}

func Test_FindNearest(t *testing.T) {
	//In radius: r = 0.1171875
	//id 0
	//numPoints 1
	//x 0.2804330060970208
	//y 0.36711590445377973
	//zoom 1

	//found 34 [98, 99, 103, 102, 91, 101, 100, 32, 128, 133, 130, 9, 11, 12, 161, 14, 13, 90, 160, 129, 131, 35, 34, 0, 31, 88, 89, 93, 92, 94, 95, 60, 96, 97]
}