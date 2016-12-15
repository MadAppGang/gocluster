package cluster_test

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"

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

func importData(filename string) []TestPoint {
	var points = struct {
		Type     string
		Features []TestPoint
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
