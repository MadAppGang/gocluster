// goluster is very fast library for  geospatial point clustering server side (or client side)
// The cluster use hierarchical greedy clustering approach.
// The same approach used by Dave Leaver with his fantastic Leaflet.markercluster plugin.
// So this approach is extremely fast, the only drawback is that all clustered points are stored in memory
// This library is deeply inspired by MapBox's superclaster JS library and blog post: https://www.mapbox.com/blog/supercluster/
//
package cluster

import (
	"math"
	"github.com/dhconnelly/rtreego"
	"fmt"
	"errors"
)


//That Zoom level indicate impossible large zoom level (Cluster's max is 21)
const InfinityZoomLevel = 100


// GeoCoordinates represent position in the Earth
type GeoCoordinates struct {
	Lon float64
	Lat float64
}

// all object, that you want to cluster should implement this protocol
type GeoPoint interface {
	GetCoordinates() GeoCoordinates
	// Any kind of identification, will be placed in result index instead of object
	// If you do not want to store the whole object
	GeoPointID() string
}

// Interface of object that Cluster use to create new clustered point
// You could implement you own class and create your own objects, or use SimpleClusteredPointsProducer
type ClusteredPointsProducer interface {
	// Construct clustered point from geo point
	NewPoint(geoPoint GeoPoint) ClusteredPoint
	NewPointWithPoints(points []ClusteredPoint, x,y float64, zoom int) ClusteredPoint
}


// The objects, that you are getting as the result of clustering
// They could store GeoPoint or just id of geo point, depends oin your own implementation
type ClusteredPoint interface {
	//Get and set mercator projection coordinates and zoom level
	GetXYZ() (float64, float64, int)
	SetXY(x, y float64)
	SetZoom(zoom int)

	//Number of points in this cluster, 1 or more
	PointsCount() int

	//Implements spatial interface of rtreego
	Bounds() *rtreego.Rect
}



// Cluster struct get a list or stream of geo objects
// and produce all levels of clusters
// MinZoom - minimum  zoom level to generate clusters
// MaxZoom - maximum zoom level to generate clusters
// Zoom range is limited by 0 to 21, and MinZoom could not be larger, then MaxZoom
// PointSize - pixel size of marker, affects clustering radius
// TileSize - size of tile in pixels, affects clustering radius
// MinBranch - Minimum Branching factor
// MaxBranch - Minimum Branching factor
// minimum and maximum branching factor are settings for index, you could play with it to tweak performance for your case
// DeepLink is flag, if it's true, all resulting ClusteredPoints will have original points in it
// if DeepLink is false, just id is stored, saving you memory and some time
type Cluster struct {
	MinZoom   int
	MaxZoom   int
	PointSize int
	TileSize  int
	MinBranch int
	MaxBranch int
	Indexes []*rtreego.Rtree
}

// Create new Cluster instance with default parameters:
// MinZoom = 0
// MaxZoom = 18
// PointSize = 50
// TileSize = 512 (GMaps and OSM default)
// MinBranch = 32
// MinBranch = 64
func NewCluster() *Cluster {
	return &Cluster{
		MinZoom:   0,
		MaxZoom:   16,
		PointSize: 40,
		TileSize:  512,
		MinBranch: 32,
		MaxBranch: 64,
	}
}

// ClusterPoint get points and create multilevel clustered indexes
// this method use SimpleClusteredPointsProducer to produce clustered points
// so you will get SimpleClusteredPoints as result
// and DeepLink will be false
func (c *Cluster) ClusterPoints(points []GeoPoint) error {
	return c.ClusterPointsWithDeepLinking(points, false)
}

// ClusterPointsWithDeepLinking get points and create multilevel clustered indexes
// this method use SimpleClusteredPointsProducer to produce clustered points
// so you will get SimpleClusteredPoints as result
// you could set DeepLink here as second param for  SimpleClusteredPointsProducer
func (c *Cluster) ClusterPointsWithDeepLinking(points []GeoPoint, deepLink bool) error {
	pp := SimpleClusteredPointsProducer{}
	pp.DeepLink = deepLink
	return c.ClusterPointsWithProducer(points, &pp)
}

// ClusterPointsWithProducer get points and create multilevel clustered indexes
// you have to provide initialised point producer here
func (c *Cluster) ClusterPointsWithProducer(points []GeoPoint, pointProducer ClusteredPointsProducer) error {

	//limit max Zoom
	if c.MaxZoom > 21 { c.MaxZoom = 21 }
	//adding extra layer for infinite zoom (initial) layers data storage
	c.Indexes = make([]*rtreego.Rtree, c.MaxZoom-c.MinZoom+2)

	clusters := translatePointsToClusters(points, pointProducer)
	for z := c.MaxZoom; z >= c.MinZoom; z-- {

		//create index from clusters from previous iteration
		idx := rtreego.NewTree(2, c.MinBranch, c.MaxBranch)
		for _, p := range clusters {
			if err := insertIndex(idx,p); err != nil {
				fmt.Println(err.Error())
			}
		}
		c.Indexes[z+1] = idx
		//create clusters for level up using just created index
		clusters = c.clusterize(clusters, z, pointProducer)
		//checkCluster(clusters)
		fmt.Println("Clusterizing results: ",z,"   :    ", len(clusters),"   ::   ",clusters)
	}

	//index topmost points
	idx := rtreego.NewTree(2, c.MinBranch, c.MaxBranch)
	for _, p := range clusters {
		if err := insertIndex(idx,p); err != nil {
			fmt.Println(err.Error())
		}
	}
	c.Indexes[c.MinZoom] = idx
	return nil
}

func insertIndex(tree *rtreego.Rtree, obj rtreego.Spatial) (err error) {
	defer func () {
		if (recover() != nil) {
			str := fmt.Sprintf("Error inserting object %+v",obj)
			err = errors.New(str)
		}
	}()

	tree.Insert(obj)
	return nil
}

func checkCluster(points []ClusteredPoint) {
	for _, v := range points {
		if _,_,z := v.GetXYZ(); z == InfinityZoomLevel {
			fmt.Println("zoom level is wrong !!!!", z)
		} else {
			fmt.Println("ok with zoom level is wrong <<", z)
		}
	}

}



//return points for  Tile with coordinates x and y and for zoom z
func (c *Cluster)GetTile(x,y,z int) []ClusteredPoint {
	index := c.Indexes[z]
	z2 := 1 << uint(z)
	z2f := float64(z2)
	extent := c.TileSize
	r := c.PointSize
	p := r / extent
	top := float64(y - p)/z2f
	bottom := float64(y+1+p) / z2f

	bbox := newRect(
		float64(x-p)/ z2f ,
		float64(top),
		float64(x+1+p)/z2f,
		bottom,
	)
	result := index.SearchIntersect(bbox)
	if (x == 0) {
		bbox = newRect(
			float64(1-p)/z2f,
			float64(top),
			1.0,
			float64(bottom),
		)
		xp :=index.SearchIntersect(bbox)
		result = append(result, xp...)
	}
	if x == (z2-1) {
		bbox = newRect(0.0,float64(top),float64(p)/z2f,float64(bottom))
		zp := index.SearchIntersect(bbox)
		result = append(result,zp ...)
	}
	rr := make ([]ClusteredPoint, len(result))
	for i := range result {
		rr[i] = result[i].(ClusteredPoint)
	}
	return rr
}

func newRect(minX, minY, maxX, maxY float64) *rtreego.Rect {
	p := rtreego.Point{minX, minY}
	sizeX := maxX - minX
	sizeY := maxY - minY
	r , _ := rtreego.NewRect(p, []float64{sizeX, sizeY})
	return r
}


//clusterize points for zoom level
func (c *Cluster)clusterize(clusters []ClusteredPoint, zoom int, pp ClusteredPointsProducer) []ClusteredPoint {
	var result []ClusteredPoint
	var r float64 = float64(c.PointSize) / float64( c.TileSize * (1 << uint(zoom)))

	//iterate all clusters
	for _, p :=  range clusters {
		//skip points we have already clustered
		if _, _, z := p.GetXYZ(); z <= zoom {
			continue
		} else {
			p.SetZoom(zoom)
			fmt.Println("SSSSSSSETTTTTT new ZZZZZ:",zoom,p)
		}

		//find all neighbours
		tree := c.Indexes[zoom+1]

		wx, wy, _ := p.GetXYZ()
		bbox := rtreego.Point{ wx, wy }.ToRect(r / 2.0)
		fmt.Println("BBBBBBBBBB: ",r,"   :   ", bbox)
		neighbours := tree.SearchIntersect(bbox)

		nPoints := p.PointsCount()
		wx = wx * float64(nPoints)
		wy = wy * float64(nPoints)

		var foundNeighbours []ClusteredPoint
		for _, nb := range neighbours {
			b := nb.(ClusteredPoint)
			bx, by, bz := b.GetXYZ()
			if ((zoom) < bz) && (dist(p,b) < r) {
				wx += bx * float64(b.PointsCount())
				wy += by * float64(b.PointsCount())
				nPoints += nPoints
				b.SetZoom(zoom) //set the zoom to skip in other iterations
				foundNeighbours = append(foundNeighbours, b)
			}
		}

		var newClaster ClusteredPoint
		if len(neighbours) <= 1 {
			if len(neighbours) == 0 {
				panic("Neigbours number is 0, impossible")
			}
			newClaster = pp.NewPointWithPoints([]ClusteredPoint{p}, wx, wy, InfinityZoomLevel)
			//fmt.Println("newCluster", newClaster)
			//fmt.Println("Created point from point")
		} else {
			wx = wx / float64(nPoints)
			wy = wy / float64(nPoints)
			newClaster = pp.NewPointWithPoints(foundNeighbours, wx, wy, InfinityZoomLevel)
			fmt.Println("Created point from neighbous", len(neighbours), " : ", newClaster)
		}
		result = append(result, newClaster)
	}
	//checkCluster(clusters)
	return result
}

////////// End of Cluster implementation


// Simple ClusteredPointsProducer implementation, produce SimpleClusteredPoint
// if you set DeepLink to true, it will include GeoPoint in SimpleClusteredPoint
// if DeepLink is false, store only array of identificators
type SimpleClusteredPointsProducer struct {
	// flag that indicate, that GeoPoint should be included in produced Clustered points
	DeepLink  bool
}

//implementation of protocols requirement, produce SimpleClusteredPoint from GeoPoint
func (pp *SimpleClusteredPointsProducer)NewPoint(geoPoint GeoPoint) ClusteredPoint {
	cp := SimpleClusteredPoint{}
	if pp.DeepLink == true {
		cp.GeoPointObjects = []GeoPoint{geoPoint}
	}
	cp.GeoPointIDs = []string{geoPoint.GeoPointID()}
	cp.Coordinates = geoPoint.GetCoordinates()
	cp.X, cp.Y = MercatorProjection(cp.Coordinates)
	cp.Zoom = InfinityZoomLevel
	return &cp
}

func (pp *SimpleClusteredPointsProducer) NewPointWithPoints(points []ClusteredPoint, x,y float64, zoom int) ClusteredPoint {
	cp := SimpleClusteredPoint{}
	nPoints := 0
	wx := 0.0
	wy := 0.0
	//fmt.Printf(">>>>>> starting to process new point %+v",points)
	for _, gp := range points {
		p := gp.(*SimpleClusteredPoint)
		if pp.DeepLink {
			cp.GeoPointObjects = append(cp.GeoPointObjects, p.GeoPointObjects...)
		}
		cp.GeoPointIDs = append( cp.GeoPointIDs, p.GeoPointIDs...)
		px, py, _ := p.GetXYZ()
		wx = wx + (px * float64(p.PointsCount()))
		wy = wy + (py * float64(p.PointsCount()))
		nPoints = nPoints + p.PointsCount()
	}

	cp.SetXY(wx / float64(nPoints),wy / float64(nPoints))
	cp.SetZoom(zoom)
	return &cp
}

// Basic implementation of clustered point,
// you could compose this struct in your own type
// SimpleClusteredPoint embedding this type
type BasicClusteredPoint struct {
	//Mercator projection to map
	X float64
	//Mercator projection to map
	Y float64
	Zoom        int
}


func (bcp *BasicClusteredPoint)GetXYZ() (float64, float64, int) {
	return bcp.X, bcp.Y, bcp.Zoom
}

func (bcp *BasicClusteredPoint)SetXY(x, y float64) {
	bcp.X = x
	bcp.Y = y
}

func (bcp *BasicClusteredPoint)SetZoom(zoom int) {
	bcp.Zoom = zoom
}

func (bcp *BasicClusteredPoint)PointsCount() int{
	return 0
}


func (bcp *BasicClusteredPoint)Bounds() *rtreego.Rect{
	const df = 0.000001 //11cm
	p :=rtreego.Point{bcp.X, bcp.Y}.ToRect(df)
	return p
}




// General and simple implementation of  ClusteredPoint interface
// SimpleClusteredPoint is produced by SimpleClusteredPointsProducer
// But you could use it as well
type SimpleClusteredPoint struct {
	BasicClusteredPoint
	//Coordinates
	Coordinates GeoCoordinates
	GeoPointIDs     []string
	GeoPointObjects []GeoPoint
}

//override basic implementation, that always returns 0
func (cp *SimpleClusteredPoint)PointsCount() int{
	return len(cp.GeoPointIDs)
}


//////
// private stuff
func translatePointsToClusters(points []GeoPoint, pointProducer ClusteredPointsProducer) []ClusteredPoint {
	var result = make([]ClusteredPoint, len(points))
	for i, p := range points {
		cp := pointProducer.NewPoint(p)
		cp.SetZoom(InfinityZoomLevel)
		result[i] = cp
	}
	return result
}


//function lngX(lng) {
//return lng / 360 + 0.5;
//}
//function latY(lat) {
//var sin = Math.sin(lat * Math.PI / 180),
//y = (0.5 - 0.25 * Math.log((1 + sin) / (1 - sin)) / Math.PI);
//return y < 0 ? 0 :
//y > 1 ? 1 : y;
//}
// longitude/latitude to spherical mercator in [0..1] range
func MercatorProjection(coordinates GeoCoordinates) (float64, float64) {
	x := coordinates.Lon / 360.0 + 0.5
	sin := math.Sin(coordinates.Lat * math.Pi / 180.0)
	y := (0.5 - 0.25 * math.Log((1+sin) / (1-sin)) / math.Pi )
	if y < 0 { y = 0 }
	if y > 1 { y = 1 }
	return x, y
}

// Dist computes the Euclidean distance between two points p and q.
func dist(q, p ClusteredPoint) float64 {
	x1, y1, _ := p.GetXYZ()
	x2, y2, _ := q.GetXYZ()
	dx := x1-x2
	dy := y1-y2
	sum := (dx*dx)+(dy*dy)
	return math.Sqrt(sum)
}
